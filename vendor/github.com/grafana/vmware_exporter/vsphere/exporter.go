package vsphere

import (
	"context"
	"net/http"

	"net/http/pprof"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/exporter-toolkit/web"
)

// Exporter holds the data needed to run a vSphere exporter
type Exporter struct {
	cfg    *Config
	logger log.Logger
	server *http.Server

	metricsHandlerFunc http.HandlerFunc
}

// NewExporter creates a new vSphere exporter from the given config
func NewExporter(logger log.Logger, cfg *Config) (*Exporter, error) {
	ctx := context.Background()
	x := &Exporter{
		cfg: cfg,
	}
	if logger == nil {
		logger = log.NewNopLogger()
	}
	x.logger = logger

	registry := prometheus.NewRegistry()
	defaultVSphere.ObjectDiscoveryInterval = cfg.ObjectDiscoveryInterval
	defaultVSphere.RefChunkSize = cfg.ChunkSize
	if cfg.CollectConcurrency > 0 {
		defaultVSphere.CollectConcurrency = cfg.CollectConcurrency
	}

	var e *endpoint
	if cfg.EnableExporterMetrics {
		goCollector := collectors.NewGoCollector()
		registry.MustRegister(goCollector)
		buildInfoCollector := collectors.NewBuildInfoCollector()
		registry.MustRegister(buildInfoCollector)
		e = newEndpoint(defaultVSphere, cfg.VSphereURL, logger, registry)
	} else {
		e = newEndpoint(defaultVSphere, cfg.VSphereURL, logger, nil)
	}

	vsphereCollector, err := newVSphereCollector(
		ctx,
		log.With(logger, "collector", "vsphere"),
		e,
	)
	if err != nil {
		return nil, err
	}
	registry.MustRegister(vsphereCollector)

	// create http server
	topMux := http.NewServeMux()
	h := newHandler(log.With(logger, "component", "handler"), registry)
	if cfg.EnableExporterMetrics {
		h = promhttp.InstrumentMetricHandler(registry, h)
	}
	// Register pprof handlers
	topMux.HandleFunc("/debug/pprof/", pprof.Index)
	topMux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	topMux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	topMux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	topMux.HandleFunc("/debug/pprof/trace", pprof.Trace)

	if cfg.TelemetryPath == "" {
		cfg.TelemetryPath = defaultConfig.TelemetryPath
	}
	topMux.Handle(cfg.TelemetryPath, h)
	x.metricsHandlerFunc = func(w http.ResponseWriter, r *http.Request) {
		h.ServeHTTP(w, r)
	}
	x.server = &http.Server{
		Addr:    cfg.ListenAddr,
		Handler: topMux,
	}
	return x, nil
}

// Start runs the exporter
func (e *Exporter) Start() error {
	level.Debug(e.logger).Log("msg", "starting the server")
	defer level.Debug(e.logger).Log("msg", "server stopped")

	flagConfig := &web.FlagConfig{
		WebConfigFile: &e.cfg.TLSConfigPath,
	}
	return web.ListenAndServe(e.server, flagConfig, log.With(e.logger, "component", "web"))
}

func (e *Exporter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	e.metricsHandlerFunc(w, r)
}

var _ http.Handler = (*Exporter)(nil)

type handler struct {
	logger      log.Logger
	promHandler http.Handler
}

func newHandler(logger log.Logger, registry *prometheus.Registry) http.Handler {
	promHandler := promhttp.HandlerFor(registry, promhttp.HandlerOpts{
		ErrorLog:            nil,
		ErrorHandling:       promhttp.PanicOnError,
		Registry:            nil,
		DisableCompression:  false,
		MaxRequestsInFlight: 0,
		Timeout:             0,
		EnableOpenMetrics:   false,
	})
	return &handler{
		logger:      logger,
		promHandler: promHandler,
	}
}

// ServeHTTP implements http.Handler.
func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	level.Debug(h.logger).Log("msg", "serving request")
	h.promHandler.ServeHTTP(w, r)
}
