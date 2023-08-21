// Package http implements the HTTP service for Flow.
package http

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"path"
	"sort"
	"strings"
	"sync"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/gorilla/mux"
	"github.com/grafana/agent/component"
	"github.com/grafana/agent/service"
	"github.com/grafana/ckit/memconn"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gorilla/mux/otelmux"
	"go.opentelemetry.io/otel/trace"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

// ServiceName defines the name used for the HTTP service.
const ServiceName = "http"

// Options are used to configure the HTTP service. Options are constant for the
// lifetime of the HTTP service.
type Options struct {
	Logger   log.Logger           // Where to send logs.
	Tracer   trace.TracerProvider // Where to send traces.
	Gatherer prometheus.Gatherer  // Where to collect metrics from.

	ReadyFunc  func() bool
	ReloadFunc func() error

	HTTPListenAddr   string // Address to listen for HTTP traffic on.
	MemoryListenAddr string // Address to accept in-memory traffic on.
	EnablePProf      bool   // Whether pprof endpoints should be exposed.
}

type Service struct {
	log      log.Logger
	tracer   trace.TracerProvider
	gatherer prometheus.Gatherer
	opts     Options

	memLis *memconn.Listener

	componentHttpPathPrefix string
}

var _ service.Service = (*Service)(nil)

// New returns a new, unstarted instance of the HTTP service.
func New(opts Options) *Service {
	var (
		l = opts.Logger
		t = opts.Tracer
		r = opts.Gatherer
	)

	if l == nil {
		l = log.NewNopLogger()
	}
	if t == nil {
		t = trace.NewNoopTracerProvider()
	}
	if r == nil {
		r = prometheus.NewRegistry()
	}

	return &Service{
		log:      l,
		tracer:   t,
		gatherer: r,
		opts:     opts,

		memLis: memconn.NewListener(l),

		componentHttpPathPrefix: "/api/v0/component/",
	}
}

// Definition returns the definition of the HTTP service.
func (s *Service) Definition() service.Definition {
	return service.Definition{
		Name:       ServiceName,
		ConfigType: nil, // http does not accept configuration
		DependsOn:  nil, // http has no dependencies.
	}
}

// Run starts the HTTP service. It will run until the provided context is
// canceled or there is a fatal error.
func (s *Service) Run(ctx context.Context, host service.Host) error {
	var wg sync.WaitGroup
	defer wg.Wait()

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	netLis, err := net.Listen("tcp", s.opts.HTTPListenAddr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", s.opts.HTTPListenAddr, err)
	}

	r := mux.NewRouter()
	r.Use(otelmux.Middleware(
		"grafana-agent",
		otelmux.WithTracerProvider(s.tracer),
	))

	r.Handle(
		"/metrics",
		promhttp.HandlerFor(s.gatherer, promhttp.HandlerOpts{}),
	)
	if s.opts.EnablePProf {
		r.PathPrefix("/debug/pprof").Handler(http.DefaultServeMux)
	}

	r.PathPrefix(s.componentHttpPathPrefix).Handler(s.componentHandler(host))

	if s.opts.ReadyFunc != nil {
		r.HandleFunc("/-/ready", func(w http.ResponseWriter, _ *http.Request) {
			if s.opts.ReadyFunc() {
				w.WriteHeader(http.StatusOK)
				fmt.Fprintln(w, "Agent is ready.")
			} else {
				w.WriteHeader(http.StatusServiceUnavailable)
				fmt.Fprintln(w, "Agent is not ready.")
			}
		})
	}

	if s.opts.ReloadFunc != nil {
		r.HandleFunc("/-/reload", func(w http.ResponseWriter, _ *http.Request) {
			level.Info(s.log).Log("msg", "reload requested via /-/reload endpoint")
			defer level.Info(s.log).Log("msg", "config reloaded")

			err := s.opts.ReloadFunc()
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			fmt.Fprintln(w, "config reloaded")
		}).Methods(http.MethodGet, http.MethodPost)
	}

	// Wire custom service handlers for services which depend on the http
	// service.
	//
	// NOTE(rfratto): keep this at the bottom of all other routes, otherwise a
	// service with a colliding path takes precedence over a predefined route.
	for _, route := range s.getServiceRoutes(host) {
		r.PathPrefix(route.Base).Handler(route.Handler)
	}

	srv := &http.Server{Handler: h2c.NewHandler(r, &http2.Server{})}

	level.Info(s.log).Log("msg", "now listening for http traffic", "addr", s.opts.HTTPListenAddr)

	listeners := []net.Listener{netLis, s.memLis}
	for _, lis := range listeners {
		wg.Add(1)
		go func(lis net.Listener) {
			defer wg.Done()
			defer cancel()

			if err := srv.Serve(lis); err != nil {
				level.Info(s.log).Log("msg", "http server closed", "addr", lis.Addr(), "err", err)
			}
		}(lis)
	}

	defer func() { _ = srv.Shutdown(ctx) }()

	<-ctx.Done()
	return nil
}

// getServiceRoutes returns a sorted list of service routes for services which
// depend on the HTTP service.
//
// Longer paths are prioritized over shorter paths so that a service with a
// more specific base route takes precedence.
func (s *Service) getServiceRoutes(host service.Host) []serviceRoute {
	var routes serviceRoutes

	for _, consumer := range host.GetServiceConsumers(ServiceName) {
		if consumer.Type != service.ConsumerTypeService {
			continue
		}

		sh, ok := consumer.Value.(ServiceHandler)
		if !ok {
			continue
		}
		base, handler := sh.ServiceHandler(host)

		routes = append(routes, serviceRoute{
			Base:    base,
			Handler: handler,
		})
	}

	sort.Sort(routes)
	return routes
}

func (s *Service) componentHandler(host service.Host) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Trim the path prefix to get our full path.
		trimmedPath := strings.TrimPrefix(r.URL.Path, s.componentHttpPathPrefix)

		// splitURLPath should only fail given an unexpected path.
		componentID, componentPath, err := splitURLPath(host, trimmedPath)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "failed to parse URL path %q: %s\n", r.URL.Path, err)
		}

		info, err := host.GetComponent(componentID, component.InfoOptions{})
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		component, ok := info.Component.(Component)
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		handler := component.Handler()
		if handler == nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		// Send just the remaining path to our component so each component can
		// handle paths from their own root path.
		r.URL.Path = componentPath
		handler.ServeHTTP(w, r)
	}
}

// Update implements [service.Service]. It is a no-op since the HTTP service
// does not support runtime configuration.
func (s *Service) Update(newConfig any) error {
	return fmt.Errorf("HTTP service does not support configuration")
}

// Data returns an instance of [Data]. Calls to Data are cachable by the
// caller.
//
// Data must only be called after parsing command-line flags.
func (s *Service) Data() any {
	return Data{
		HTTPListenAddr:   s.opts.HTTPListenAddr,
		MemoryListenAddr: s.opts.MemoryListenAddr,
		BaseHTTPPath:     s.componentHttpPathPrefix,

		DialFunc: func(ctx context.Context, network, address string) (net.Conn, error) {
			switch address {
			case s.opts.MemoryListenAddr:
				return s.memLis.DialContext(ctx)
			default:
				return (&net.Dialer{}).DialContext(ctx, network, address)
			}
		},
	}
}

// Data includes information associated with the HTTP service.
type Data struct {
	// Address that the HTTP service is configured to listen on.
	HTTPListenAddr string

	// Address that the HTTP service is configured to listen on for in-memory
	// traffic when [DialFunc] is used to establish a connection.
	MemoryListenAddr string

	// BaseHTTPPath is the base path where component HTTP routes are exposed.
	BaseHTTPPath string

	// DialFunc is a function which establishes in-memory network connection when
	// address is MemoryListenAddr. If address is not MemoryListenAddr, DialFunc
	// establishes an outbound network connection.
	DialFunc func(ctx context.Context, network, address string) (net.Conn, error)
}

// HTTPPathForComponent returns the full HTTP path for a given global component
// ID.
func (d Data) HTTPPathForComponent(componentID string) string {
	merged := path.Join(d.BaseHTTPPath, componentID)
	if !strings.HasSuffix(merged, "/") {
		return merged + "/"
	}
	return merged
}

// Component is a Flow component which also contains a custom HTTP handler.
type Component interface {
	component.Component

	// Handler should return a valid HTTP handler for the component.
	// All requests to the component will have the path trimmed such that the component is at the root.
	// For example, f a request is made to `/component/{id}/metrics`, the component
	// will receive a request to just `/metrics`.
	Handler() http.Handler
}

// ServiceHandler is a Service which exposes custom HTTP handlers.
type ServiceHandler interface {
	service.Service

	// ServiceHandler returns the base route and HTTP handlers to register for
	// the provided service.
	//
	// This method is only called for services that declare a dependency on
	// the http service.
	//
	// The http service prioritizes longer base routes. Given two base routes of
	// /foo and /foo/bar, an HTTP URL of /foo/bar/baz will be routed to the
	// longer base route (/foo/bar).
	ServiceHandler(host service.Host) (base string, handler http.Handler)
}
