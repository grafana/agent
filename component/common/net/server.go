package net

import (
	"crypto/tls"
	"fmt"
	"github.com/go-kit/log"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/model"
	"github.com/prometheus/exporter-toolkit/web"
	"github.com/weaveworks/common/instrument"
	"github.com/weaveworks/common/logging"
	"github.com/weaveworks/common/middleware"
	"github.com/weaveworks/common/signals"
	"golang.org/x/net/context"
	"golang.org/x/net/netutil"
	"net"
	"net/http"
)

// Listen on the named network
const (
	// DefaultNetwork  the host resolves to multiple IP addresses,
	// Dial will try each IP address in order until one succeeds
	DefaultNetwork = "tcp"
)

// SignalHandler used by Server.
type SignalHandler interface {
	// Starts the signals handler. This method is blocking, and returns only after signal is received,
	// or "Stop" is called.
	Loop()

	// Stop blocked "Loop" method.
	Stop()
}

// Server wraps a HTTP and gRPC server, and some common initialization.
//
// Servers will be automatically instrumented for Prometheus metrics.
type Server struct {
	cfg          ServerConfig
	handler      SignalHandler
	httpListener net.Listener

	HTTP       *mux.Router
	HTTPServer *http.Server

	Log        logging.Interface
	Registerer prometheus.Registerer
	Gatherer   prometheus.Gatherer
}

// NewWithDefaults creates a new Server, applying some defaults to the server configuration.
// If provided config is nil, a default configuration will be used instead.
func NewWithDefaults(logger log.Logger, metricsNamespace string, reg prometheus.Registerer, config *ServerConfig) (*Server, error) {
	if !model.IsValidMetricName(model.LabelValue(metricsNamespace)) {
		return nil, fmt.Errorf("metrics namespace is not prometheus compatiible: %s", metricsNamespace)
	}

	// Apply some defaults if nothing provided
	if config == nil {
		config = &ServerConfig{}
		// Set the config to the new combined config.
		// Avoid logging entire received request on failures
		config.ExcludeRequestInLog = true
		// Configure dedicated metrics registerer
		config.Registerer = reg
		// To prevent metric collisions because all metrics are going to be registered in the global Prometheus registry.
		config.MetricsNamespace = metricsNamespace
		// We don't want the /debug and /metrics endpoints running, since this is not the main Flow HTTP server.
		// We want this target to expose the least surface area possible, hence disabling WeaveWorks HTTP server metrics
		// and debugging functionality.
		config.RegisterInstrumentation = false
		// Add logger to weaveworks
		config.Log = logging.GoKit(logger)

	}

	// Apply some defaults if nothing provided
	if config.HTTP == nil {
		config.HTTP = &HTTPConfig{ListenPort: DefaultHTTPPort}
	}

	return New(*config)
}

// New makes a new Server
func New(cfg ServerConfig) (*Server, error) {
	// If user doesn't supply a logging implementation, by default instantiate
	// logrus.
	logger := cfg.Log
	if logger == nil {
		logger = logging.NewLogrus(cfg.LogLevel)
	}

	// If user doesn't supply a registerer/gatherer, use Prometheus' by default.
	reg := cfg.Registerer
	if reg == nil {
		reg = prometheus.DefaultRegisterer
	}
	gatherer := cfg.Gatherer
	if gatherer == nil {
		gatherer = prometheus.DefaultGatherer
	}

	tcpConnections := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: cfg.MetricsNamespace,
		Name:      "tcp_connections",
		Help:      "Current number of accepted TCP connections.",
	}, []string{"protocol"})
	reg.MustRegister(tcpConnections)

	tcpConnectionsLimit := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: cfg.MetricsNamespace,
		Name:      "tcp_connections_limit",
		Help:      "The max number of TCP connections that can be accepted (0 means no limit).",
	}, []string{"protocol"})
	reg.MustRegister(tcpConnectionsLimit)

	network := cfg.HTTPListenNetwork
	if network == "" {
		network = DefaultNetwork
	}
	// Setup listeners first, so we can fail early if the port is in use.
	httpListener, err := net.Listen(network, fmt.Sprintf("%s:%d", cfg.HTTP.ListenAddress, cfg.HTTP.ListenPort))
	if err != nil {
		return nil, err
	}
	httpListener = middleware.CountingListener(httpListener, tcpConnections.WithLabelValues("http"))

	tcpConnectionsLimit.WithLabelValues("http").Set(float64(cfg.HTTP.ConnLimit))
	if cfg.HTTP.ConnLimit > 0 {
		httpListener = netutil.LimitListener(httpListener, cfg.HTTP.ConnLimit)
	}

	cipherSuites, err := stringToCipherSuites(cfg.CipherSuites)
	if err != nil {
		return nil, err
	}
	minVersion, err := stringToTLSVersion(cfg.MinVersion)
	if err != nil {
		return nil, err
	}

	// Setup TLS
	var httpTLSConfig *tls.Config
	if len(cfg.HTTPTLSConfig.TLSCertPath) > 0 && len(cfg.HTTPTLSConfig.TLSKeyPath) > 0 {
		// Note: ConfigToTLSConfig from prometheus/exporter-toolkit is awaiting security review.
		httpTLSConfig, err = web.ConfigToTLSConfig(&web.TLSConfig{
			TLSCertPath:  cfg.HTTPTLSConfig.TLSCertPath,
			TLSKeyPath:   cfg.HTTPTLSConfig.TLSKeyPath,
			ClientAuth:   cfg.HTTPTLSConfig.ClientAuth,
			ClientCAs:    cfg.HTTPTLSConfig.ClientCAs,
			CipherSuites: cipherSuites,
			MinVersion:   minVersion,
		})
		if err != nil {
			return nil, fmt.Errorf("error generating http tls config: %v", err)
		}
	}

	// Prometheus histograms for requests.
	requestDuration := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: cfg.MetricsNamespace,
		Name:      "request_duration_seconds",
		Help:      "Time (in seconds) spent serving HTTP requests.",
		Buckets:   instrument.DefBuckets,
	}, []string{"method", "route", "status_code", "ws"})
	reg.MustRegister(requestDuration)

	receivedMessageSize := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: cfg.MetricsNamespace,
		Name:      "request_message_bytes",
		Help:      "Size (in bytes) of messages received in the request.",
		Buckets:   middleware.BodySizeBuckets,
	}, []string{"method", "route"})
	reg.MustRegister(receivedMessageSize)

	sentMessageSize := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: cfg.MetricsNamespace,
		Name:      "response_message_bytes",
		Help:      "Size (in bytes) of messages sent in response.",
		Buckets:   middleware.BodySizeBuckets,
	}, []string{"method", "route"})
	reg.MustRegister(sentMessageSize)

	inflightRequests := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: cfg.MetricsNamespace,
		Name:      "inflight_requests",
		Help:      "Current number of inflight requests.",
	}, []string{"method", "route"})
	reg.MustRegister(inflightRequests)

	logger.WithField("http", httpListener.Addr()).Infof("server listening on addresses")

	// Setup HTTP server
	var router *mux.Router
	if cfg.Router != nil {
		router = cfg.Router
	} else {
		router = mux.NewRouter()
	}
	if cfg.PathPrefix != "" {
		// Expect metrics and pprof handlers to be prefixed with server's path prefix.
		// e.g. /loki/metrics or /loki/debug/pprof
		router = router.PathPrefix(cfg.PathPrefix).Subrouter()
	}
	if cfg.RegisterInstrumentation {
		RegisterInstrumentationWithGatherer(router, gatherer)
	}

	var sourceIPs *middleware.SourceIPExtractor
	if cfg.LogSourceIPs {
		sourceIPs, err = middleware.NewSourceIPs(cfg.LogSourceIPsHeader, cfg.LogSourceIPsRegex)
		if err != nil {
			return nil, fmt.Errorf("error setting up source IP extraction: %v", err)
		}
	}

	defaultHTTPMiddleware := []middleware.Interface{
		middleware.Tracer{
			RouteMatcher: router,
			SourceIPs:    sourceIPs,
		},
		middleware.Log{
			Log:                   logger,
			SourceIPs:             sourceIPs,
			LogRequestAtInfoLevel: cfg.LogRequestAtInfoLevel,
		},
		middleware.Instrument{
			RouteMatcher:     router,
			Duration:         requestDuration,
			RequestBodySize:  receivedMessageSize,
			ResponseBodySize: sentMessageSize,
			InflightRequests: inflightRequests,
		},
	}
	httpMiddleware := []middleware.Interface{}
	if cfg.DoNotAddDefaultHTTPMiddleware {
		httpMiddleware = cfg.HTTPMiddleware
	} else {
		httpMiddleware = append(defaultHTTPMiddleware, cfg.HTTPMiddleware...)
	}

	httpServer := &http.Server{
		ReadTimeout:  cfg.HTTP.ServerReadTimeout,
		WriteTimeout: cfg.HTTP.ServerWriteTimeout,
		IdleTimeout:  cfg.HTTP.ServerIdleTimeout,
		Handler:      middleware.Merge(httpMiddleware...).Wrap(router),
	}
	if httpTLSConfig != nil {
		httpServer.TLSConfig = httpTLSConfig
	}

	handler := cfg.SignalHandler
	if handler == nil {
		handler = signals.NewHandler(logger)
	}

	return &Server{
		cfg:          cfg,
		httpListener: httpListener,
		handler:      handler,

		HTTP:       router,
		HTTPServer: httpServer,
		Log:        logger,
		Registerer: reg,
		Gatherer:   gatherer,
	}, nil
}

// RegisterInstrumentationWithGatherer on the given router.
func RegisterInstrumentationWithGatherer(router *mux.Router, gatherer prometheus.Gatherer) {
	router.Handle("/metrics", promhttp.HandlerFor(gatherer, promhttp.HandlerOpts{
		EnableOpenMetrics: true,
	}))
}

// Run the server; blocks until SIGTERM (if signal handling is enabled), an error is received, or Stop() is called.
func (s *Server) Run() error {
	errChan := make(chan error, 1)

	// Wait for a signal
	go func() {
		s.handler.Loop()
		select {
		case errChan <- nil:
		default:
		}
	}()

	go func() {
		var err error
		if s.HTTPServer.TLSConfig == nil {
			err = s.HTTPServer.Serve(s.httpListener)
		} else {
			err = s.HTTPServer.ServeTLS(s.httpListener, s.cfg.HTTPTLSConfig.TLSCertPath, s.cfg.HTTPTLSConfig.TLSKeyPath)
		}
		if err == http.ErrServerClosed {
			err = nil
		}

		select {
		case errChan <- err:
		default:
		}
	}()

	return <-errChan
}

// HTTPListenAddr exposes `net.Addr` that `Server` is listening to for HTTP connections.
func (s *Server) HTTPListenAddr() net.Addr {
	return s.httpListener.Addr()

}

// Stop unblocks Run().
func (s *Server) Stop() {
	s.handler.Stop()
}

// Shutdown the server, gracefully.  Should be defered after New().
func (s *Server) Shutdown() {
	ctx, cancel := context.WithTimeout(context.Background(), s.cfg.GracefulShutdownTimeout)
	defer cancel() // releases resources if httpServer.Shutdown completes before timeout elapses

	_ = s.HTTPServer.Shutdown(ctx)
}

// MountAndRun mounts the handlers and starting the server.
func (ts *Server) MountAndRun(mountRoute func(router *mux.Router)) error {
	ts.Log.Infof("starting server")
	mountRoute(ts.HTTP)

	go func() {
		err := ts.Run()
		if err != nil {
			ts.Log.Errorf("server shutdown with error: %v", err)
		}
	}()

	return nil
}

// StopAndShutdown stops and shuts down the underlying server.
func (s *Server) StopAndShutdown() {
	s.Stop()
	s.Shutdown()
}
