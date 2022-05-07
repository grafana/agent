package ssl_exporter

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	ssl_config "github.com/ribbybibby/ssl_exporter/v2/config"
	"github.com/ribbybibby/ssl_exporter/v2/prober"
)

type sslHandler struct {
	cfg     *Config
	modules *ssl_config.Config
	log     log.Logger
}

func (sh *sslHandler) handler(w http.ResponseWriter, r *http.Request) {
	logger := sh.log

	query := r.URL.Query()

	SSLTargets := make(map[string]SSLTarget)
	for _, target := range sh.cfg.SSLTargets {
		SSLTargets[target.Name] = target
	}

	var target string
	targetName := query.Get("target")
	if len(query["target"]) != 1 || targetName == "" {
		http.Error(w, "'target' parameter must be specified once", 400)
		return
	}

	t, ok := SSLTargets[targetName]
	if ok {
		target = t.Target
	} else {
		target = targetName
	}

	moduleName := query.Get("module")
	if len(query["module"]) > 1 {
		http.Error(w, "'module' parameter must only be specified once", 400)
		return
	}
	if moduleName == "" {
		moduleName = "tcp"
	}

	module, ok := sh.modules.Modules[moduleName]
	if !ok {
		http.Error(w, fmt.Sprintf("Unknown module '%s'", moduleName), 400)
		return
	}

	probeFunc, ok := prober.Probers[module.Prober]
	if !ok {
		http.Error(w, fmt.Sprintf("Unknown prober %q", module.Prober), 400)
		return
	}

	var (
		probeSuccess = prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: prometheus.BuildFQName("ssl", "", "probe_success"),
				Help: "If the probe was a success",
			},
		)
		proberType = prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: prometheus.BuildFQName("ssl", "", "prober"),
				Help: "The prober used by the exporter to connect to the target",
			},
			[]string{"prober"},
		)
	)

	logger = log.With(logger, "module", moduleName, "target", target)
	level.Debug(logger).Log("msg", "Starting scrape")

	start := time.Now()
	registry := prometheus.NewRegistry()
	registry.MustRegister(probeSuccess, proberType)

	proberType.WithLabelValues(module.Prober).Set(1)

	err := probeFunc(r.Context(), logger, target, module, registry)
	if err != nil {
		level.Error(logger).Log("msg", err)
		probeSuccess.Set(0)
	} else {
		probeSuccess.Set(1)
	}

	h := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
	h.ServeHTTP(w, r)
	duration := time.Since(start).Seconds()
	level.Debug(logger).Log("msg", "Finished scrape", "duration_seconds", duration)
}

func (sh sslHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	sh.handler(w, r)
}
