// Package server implements the HTTP and gRPC server used throughout Grafana
// Agent.
//
// It is a grafana/agent-specific fork of github.com/weaveworks/common/server.
package server

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	_ "net/http/pprof" // anonymous import to get the pprof handler registered
	"reflect"
	"sync"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/gorilla/mux"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/hashicorp/go-multierror"
	"github.com/oklog/run"
	otgrpc "github.com/opentracing-contrib/go-grpc"
	"github.com/opentracing/opentracing-go"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/weaveworks/common/middleware"
	"golang.org/x/net/netutil"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
)

// TODO(rfratto):
// - document removal of PathPrefix
// - document removal of TLS flags
// - should the flags for listen address be host:port instead of having two flags? (Yes)

// Server wraps an HTTP and gRPC server with some common initialization.
//
// Unless instrumentation is disabled in the Servers config, Prometheus metrics
// will be automatically generated for the server.
type Server struct {
	optsMut sync.Mutex
	opts    Flags

	// Listeners to use for connections. These will use TLS when TLS is enabled.
	httpListener net.Listener
	grpcListener net.Listener

	updateHTTPTLS func(TLSConfig) error
	updateGRPCTLS func(TLSConfig) error

	HTTP       *mux.Router
	HTTPServer *http.Server
	GRPC       *grpc.Server
}

// New creates a new Server with the given config.
//
// r is used to register Server-specific metrics. If r is nil, no metrics will
// be registered.
//
// g is used for collecting metrics from the instrumentation handlers, when
// enabled. If g is nil, a /metrics endpoint will not be registered.
func New(l log.Logger, r prometheus.Registerer, g prometheus.Gatherer, cfg Config) (srv *Server, err error) {
	// TODO(rfratto): make a argument and remove from Config struct in v0.26.0.
	opts := cfg.Flags

	if l == nil {
		l = log.NewNopLogger()
	}
	wrappedLogger := GoKitLogger(l)

	// Create metrics for the server
	var (
		tcpConnections = prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "agent_tcp_connections",
			Help: "Current number of accepted TCP connections.",
		}, []string{"protocol"})
		tcpConnectionsLimit = prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "agent_tcp_connections_limit",
			Help: "The maximum number of TCP connections that can be accepted (0 = unlimited)",
		}, []string{"protocol"})
		requestDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name: "agent_request_duration_seconds",
			Help: "Time in seconds spent serving HTTP requests.",
		}, []string{"method", "route", "status_code", "ws"})
		receivedMessageSize = prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "agent_request_message_bytes",
			Help:    "Size (in bytes) of messages received in the request.",
			Buckets: middleware.BodySizeBuckets,
		}, []string{"method", "route"})
		sentMessageSize = prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "agent_response_message_bytes",
			Help:    "Size (in bytes) of messages sent in response.",
			Buckets: middleware.BodySizeBuckets,
		}, []string{"method", "route"})
		inflightRequests = prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "agent_inflight_requests",
			Help: "Current number of inflight requests.",
		}, []string{"method", "route"})
	)
	if r != nil {
		// Register all of our metrics
		cc := []prometheus.Collector{
			tcpConnections, tcpConnectionsLimit, requestDuration, receivedMessageSize,
			sentMessageSize, inflightRequests,
		}
		for _, c := range cc {
			if err := r.Register(c); err != nil {
				return nil, fmt.Errorf("failed registering server metrics: %w", err)
			}
		}
	}

	// Create listeners first so we can fail early if the port is in use.

	// HTTP listener setup
	httpAddress := opts.HTTP.GetListenAddress()
	if httpAddress == "" {
		return nil, fmt.Errorf("http address not set")
	}
	httpListener, err := net.Listen(opts.HTTP.ListenNetwork, httpAddress)
	if err != nil {
		return nil, fmt.Errorf("creating HTTP listener: %w", err)
	}
	httpListener = middleware.CountingListener(httpListener, tcpConnections.WithLabelValues("http"))
	defer func() {
		if err != nil {
			_ = httpListener.Close()
		}
	}()

	tcpConnectionsLimit.WithLabelValues("http").Set(float64(opts.HTTP.ConnLimit))
	if opts.HTTP.ConnLimit > 0 {
		httpListener = netutil.LimitListener(httpListener, opts.HTTP.ConnLimit)
	}

	// gRPC listener setup
	grpcAddress := opts.GRPC.GetListenAddress()
	if grpcAddress == "" {
		return nil, fmt.Errorf("gRPC address not set")
	}
	grpcListener, err := net.Listen(opts.GRPC.ListenNetwork, grpcAddress)
	if err != nil {
		return nil, fmt.Errorf("creating gRPC listener: %w", err)
	}
	grpcListener = middleware.CountingListener(grpcListener, tcpConnections.WithLabelValues("grpc"))
	defer func() {
		if err != nil {
			_ = grpcListener.Close()
		}
	}()

	tcpConnectionsLimit.WithLabelValues("grpc").Set(float64(opts.GRPC.ConnLimit))
	if opts.GRPC.ConnLimit > 0 {
		grpcListener = netutil.LimitListener(httpListener, opts.GRPC.ConnLimit)
	}

	// Configure TLS
	var (
		updateHTTPTLS func(TLSConfig) error
		updateGRPCTLS func(TLSConfig) error
	)
	if opts.HTTP.UseTLS {
		httpTLSListener, err := newTLSListener(httpListener, cfg.HTTP.TLSConfig)
		if err != nil {
			return nil, fmt.Errorf("generating HTTP TLS config: %w", err)
		}
		httpListener = httpTLSListener
		updateHTTPTLS = httpTLSListener.ApplyConfig
	}
	if opts.GRPC.UseTLS {
		grpcTLSListener, err := newTLSListener(grpcListener, cfg.GRPC.TLSConfig)
		if err != nil {
			return nil, fmt.Errorf("generating GRPC TLS config: %w", err)
		}
		grpcListener = grpcTLSListener
		updateGRPCTLS = grpcTLSListener.ApplyConfig
	}

	level.Info(l).Log(
		"msg", "server listening on addresses",
		"http", httpListener.Addr(), "grpc", grpcListener.Addr(),
		"http_tls_enabled", opts.HTTP.UseTLS, "grpc_tls_enabled", opts.GRPC.UseTLS,
	)

	// Configure gRPC server
	serverLog := middleware.GRPCServerLog{
		WithRequest: true,
		Log:         wrappedLogger,
	}
	grpcOptions := []grpc.ServerOption{
		grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
			serverLog.UnaryServerInterceptor,
			otgrpc.OpenTracingServerInterceptor(opentracing.GlobalTracer()),
			middleware.UnaryServerInstrumentInterceptor(requestDuration),
		)),
		grpc.StreamInterceptor(grpc_middleware.ChainStreamServer(
			serverLog.StreamServerInterceptor,
			otgrpc.OpenTracingStreamServerInterceptor(opentracing.GlobalTracer()),
			middleware.StreamServerInstrumentInterceptor(requestDuration),
		)),
		grpc.KeepaliveParams(keepalive.ServerParameters{
			MaxConnectionIdle:     opts.GRPC.MaxConnectionIdle,
			MaxConnectionAge:      opts.GRPC.MaxConnectionAge,
			MaxConnectionAgeGrace: opts.GRPC.MaxConnectionAgeGrace,
			Time:                  opts.GRPC.KeepaliveTime,
			Timeout:               opts.GRPC.KeepaliveTimeout,
		}),
		grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
			MinTime:             opts.GRPC.MinTimeBetweenPings,
			PermitWithoutStream: opts.GRPC.PingWithoutStreamAllowed,
		}),
		grpc.MaxRecvMsgSize(opts.GRPC.MaxRecvMsgSize),
		grpc.MaxSendMsgSize(opts.GRPC.MaxSendMsgSize),
		grpc.MaxConcurrentStreams(uint32(opts.GRPC.MaxConcurrentStreams)),
		grpc.StatsHandler(middleware.NewStatsHandler(receivedMessageSize, sentMessageSize, inflightRequests)),
	}
	grpcServer := grpc.NewServer(grpcOptions...)

	router := mux.NewRouter()
	if opts.RegisterInstrumentation && g != nil {
		router.Handle("/metrics", promhttp.HandlerFor(g, promhttp.HandlerOpts{
			EnableOpenMetrics: true,
		}))
		router.PathPrefix("/debug/pprof").Handler(http.DefaultServeMux)
	}

	var sourceIPs *middleware.SourceIPExtractor
	if opts.LogSourceIPs {
		sourceIPs, err = middleware.NewSourceIPs(opts.LogSourceIPsHeader, opts.LogSourceIPsRegex)
		if err != nil {
			return nil, fmt.Errorf("error setting up source IP extraction: %v", err)
		}
	}

	httpMiddleware := []middleware.Interface{
		middleware.Tracer{
			RouteMatcher: router,
			SourceIPs:    sourceIPs,
		},
		middleware.Log{
			Log:       wrappedLogger,
			SourceIPs: sourceIPs,
		},
		middleware.Instrument{
			RouteMatcher:     router,
			Duration:         requestDuration,
			RequestBodySize:  receivedMessageSize,
			ResponseBodySize: sentMessageSize,
			InflightRequests: inflightRequests,
		},
	}

	httpServer := &http.Server{
		ReadTimeout:  opts.HTTP.ReadTimeout,
		WriteTimeout: opts.HTTP.WriteTimeout,
		IdleTimeout:  opts.HTTP.IdleTimeout,
		Handler:      middleware.Merge(httpMiddleware...).Wrap(router),
	}

	return &Server{
		opts:         opts,
		httpListener: httpListener,
		grpcListener: grpcListener,

		updateHTTPTLS: updateHTTPTLS,
		updateGRPCTLS: updateGRPCTLS,

		HTTP:       router,
		HTTPServer: httpServer,
		GRPC:       grpcServer,
	}, nil
}

// HTTPAddress returns the HTTP net.Addr of this Server.
func (s *Server) HTTPAddress() net.Addr { return s.httpListener.Addr() }

// GRPCAddress returns the GRPC net.Addr of this Server.
func (s *Server) GRPCAddress() net.Addr { return s.grpcListener.Addr() }

// ApplyConfig applies changes to the Server block. ApplyConfig will fail if
// the cfg.Flags field has been changed.
//
// v0.26.0 will remove YAML support for cfg.Flags and remove it out of the
// Config struct to simplify dynamic updating.
func (s *Server) ApplyConfig(cfg Config) error {
	s.optsMut.Lock()
	defer s.optsMut.Unlock()

	// N.B. LogLevel/LogFormat support dynamic updating but are never used in
	// *Server, so they're ignored here.

	if s.updateHTTPTLS != nil {
		if err := s.updateHTTPTLS(cfg.HTTP.TLSConfig); err != nil {
			return fmt.Errorf("updating HTTP TLS settings: %w", err)
		}
	}
	if s.updateGRPCTLS != nil {
		if err := s.updateGRPCTLS(cfg.GRPC.TLSConfig); err != nil {
			return fmt.Errorf("updating gRPC TLS settings: %w", err)
		}
	}

	if !reflect.DeepEqual(s.opts, cfg.Flags) {
		return fmt.Errorf("cannot dynamically update values for deprecated YAML fields")
	}
	return nil
}

// Run the server until en error is received or the given context is canceled.
// Run may not be re-called after it exits.
func (s *Server) Run(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var g run.Group

	g.Add(func() error {
		<-ctx.Done()
		return nil
	}, func(_ error) {
		cancel()
	})

	g.Add(func() error {
		err := s.HTTPServer.Serve(s.httpListener)
		if errors.Is(err, http.ErrServerClosed) {
			err = nil
		}
		return err
	}, func(_ error) {
		ctx, cancel := context.WithTimeout(context.Background(), s.opts.GracefulShutdownTimeout)
		defer cancel()
		_ = s.HTTPServer.Shutdown(ctx)
	})

	g.Add(func() error {
		err := s.GRPC.Serve(s.grpcListener)
		if errors.Is(err, grpc.ErrServerStopped) {
			err = nil
		}
		return err
	}, func(_ error) {
		s.GRPC.GracefulStop()
	})

	return g.Run()
}

// Close forcibly closes the server's listeners.
func (s *Server) Close() error {
	errs := multierror.Append(
		s.httpListener.Close(),
		s.grpcListener.Close(),
	)
	return errs.ErrorOrNil()
}
