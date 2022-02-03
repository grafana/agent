package snmp_exporter

import (
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/snmp_exporter/collector"
	snmp_config "github.com/prometheus/snmp_exporter/config"
)

var (
	// Metrics about the SNMP exporter itself.
	snmpDuration = promauto.NewSummaryVec(
		prometheus.SummaryOpts{
			Name: "snmp_collection_duration_seconds",
			Help: "Duration of collections by the SNMP exporter",
		},
		[]string{"module"},
	)
	snmpRequestErrors = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "snmp_request_errors_total",
			Help: "Errors in requests to the SNMP exporter",
		},
	)
)

func handler(w http.ResponseWriter, r *http.Request, logger log.Logger, modules *snmp_config.Config) {
	query := r.URL.Query()

	target := query.Get("target")
	if len(query["target"]) != 1 || target == "" {
		http.Error(w, "'target' parameter must be specified once", 400)
		snmpRequestErrors.Inc()
		return
	}

	moduleName := query.Get("module")
	if len(query["module"]) > 1 {
		http.Error(w, "'module' parameter must only be specified once", 400)
		snmpRequestErrors.Inc()
		return
	}
	if moduleName == "" {
		moduleName = "if_mib"
	}

	module, ok := (*modules)[moduleName]
	if !ok {
		http.Error(w, fmt.Sprintf("Unknown module '%s'", moduleName), 400)
		snmpRequestErrors.Inc()
		return
	}

	logger = log.With(logger, "module", moduleName, "target", target)
	level.Debug(logger).Log("msg", "Starting scrape")

	start := time.Now()
	registry := prometheus.NewRegistry()
	c := collector.New(r.Context(), target, module, logger)
	registry.MustRegister(c)
	// Delegate http serving to Prometheus client library, which will call collector.Collect.
	h := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
	h.ServeHTTP(w, r)
	duration := time.Since(start).Seconds()
	snmpDuration.WithLabelValues(moduleName).Observe(duration)
	level.Debug(logger).Log("msg", "Finished scrape", "duration_seconds", duration)
}

type snmpHandler struct {
	cfg     *Config
	modules *snmp_config.Config
	log     log.Logger
}

func (sh snmpHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	handler(w, r, sh.log, sh.modules)
}
