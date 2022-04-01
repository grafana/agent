package snmp_exporter

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/snmp_exporter/collector"
	snmp_config "github.com/prometheus/snmp_exporter/config"
)

var (
	// SnmpDuration collects duration
	SnmpDuration = promauto.NewSummaryVec(
		prometheus.SummaryOpts{
			Name: "snmp_collection_duration_seconds",
			Help: "Duration of collections by the SNMP exporter",
		},
		[]string{"module"},
	)
	// SnmpRequestErrors counters errors
	SnmpRequestErrors = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "snmp_request_errors_total",
			Help: "Errors in requests to the SNMP exporter",
		},
	)
)

type snmpHandler struct {
	cfg     *Config
	modules *snmp_config.Config
	log     log.Logger
}

func (sh *snmpHandler) handler(w http.ResponseWriter, r *http.Request) {
	logger := sh.log

	query := r.URL.Query()

	snmpTargets := make(map[string]SNMPTarget)
	for _, target := range sh.cfg.SnmpTargets {
		snmpTargets[target.Name] = target
	}

	var target string
	targetName := query.Get("target")
	if len(query["target"]) != 1 || targetName == "" {
		http.Error(w, "'target' parameter must be specified once", 400)
		SnmpRequestErrors.Inc()
		return
	}

	t, ok := snmpTargets[targetName]
	if ok {
		target = t.Target
	} else {
		target = targetName
	}

	moduleName := query.Get("module")
	if len(query["module"]) > 1 {
		http.Error(w, "'module' parameter must only be specified once", 400)
		SnmpRequestErrors.Inc()
		return
	}
	if moduleName == "" {
		moduleName = "if_mib"
	}

	module, ok := (*sh.modules)[moduleName]
	if !ok {
		http.Error(w, fmt.Sprintf("Unknown module '%s'", moduleName), 400)
		SnmpRequestErrors.Inc()
		return
	}

	// override module connection details with custom walk params if provided
	walkParams := query.Get("walk_params")
	if len(query["walk_params"]) > 1 {
		http.Error(w, "'walk_params' parameter must only be specified once", 400)
		SnmpRequestErrors.Inc()
		return
	}

	if walkParams != "" {
		if wp, ok := sh.cfg.WalkParams[walkParams]; ok {
			// module.WalkParams = wp
			if wp.Version != 0 {
				module.WalkParams.Version = wp.Version
			}
			if wp.MaxRepetitions != 0 {
				module.WalkParams.MaxRepetitions = wp.MaxRepetitions
			}
			if wp.Retries != 0 {
				module.WalkParams.Retries = wp.Retries
			}
			if wp.Timeout != 0 {
				module.WalkParams.Timeout = wp.Timeout
			}
			module.WalkParams.Auth = wp.Auth
		} else {
			http.Error(w, fmt.Sprintf("Unknown walk_params '%s'", walkParams), 400)
			SnmpRequestErrors.Inc()
			return
		}
		logger = log.With(logger, "module", moduleName, "target", target, "walk_params", walkParams)
	} else {
		logger = log.With(logger, "module", moduleName, "target", target)
	}
	level.Debug(logger).Log("msg", "Starting scrape")

	start := time.Now()
	registry := prometheus.NewRegistry()
	c := collector.New(r.Context(), target, module, logger)
	registry.MustRegister(c)
	// Delegate http serving to Prometheus client library, which will call collector.Collect.
	h := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
	h.ServeHTTP(w, r)
	duration := time.Since(start).Seconds()
	SnmpDuration.WithLabelValues(moduleName).Observe(duration)
	level.Debug(logger).Log("msg", "Finished scrape", "duration_seconds", duration)
}

func (sh snmpHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	sh.handler(w, r)
}
