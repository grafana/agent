package snmp_exporter

import (
	"context"
	"fmt"
	"net/http"
	"path"
	"time"

	"github.com/gorilla/mux"
	"github.com/grafana/agent/pkg/integrations/v2"
	"github.com/grafana/agent/pkg/integrations/v2/autoscrape"
	"github.com/grafana/agent/pkg/integrations/v2/metricsutils"
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/config"
	"github.com/prometheus/prometheus/discovery"
	"github.com/prometheus/prometheus/discovery/targetgroup"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	v1snmp "github.com/grafana/agent/pkg/integrations/snmp_exporter"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/snmp_exporter/collector"
	snmp_config "github.com/prometheus/snmp_exporter/config"
)

type snmpHandler struct {
	cfg     *Config
	modules *snmp_config.Config
	log     log.Logger
}

func (sh *snmpHandler) Targets(ep integrations.Endpoint) []*targetgroup.Group {
	integrationNameValue := model.LabelValue("integrations/" + sh.cfg.Name())
	key, _ := sh.cfg.InstanceKey("")

	group := &targetgroup.Group{
		Labels: model.LabelSet{
			model.InstanceLabel: model.LabelValue(key),
			model.JobLabel:      integrationNameValue,
			"agent_hostname":    model.LabelValue(sh.cfg.globals.AgentIdentifier),

			// Meta labels that can be used during SD.
			"__meta_agent_integration_name":       model.LabelValue(sh.cfg.Name()),
			"__meta_agent_integration_instance":   model.LabelValue(sh.cfg.Name()),
			"__meta_agent_integration_autoscrape": model.LabelValue(metricsutils.BoolToString(*sh.cfg.Common.Autoscrape.Enable)),
		},
		Source: fmt.Sprintf("%s/%s", sh.cfg.Name(), sh.cfg.Name()),
	}

	for _, lbl := range sh.cfg.Common.ExtraLabels {
		group.Labels[model.LabelName(lbl.Name)] = model.LabelValue(lbl.Value)
	}

	for _, t := range sh.cfg.SnmpTargets {
		group.Targets = append(group.Targets, model.LabelSet{
			model.AddressLabel:     model.LabelValue(ep.Host),
			model.MetricsPathLabel: model.LabelValue(path.Join(ep.Prefix, "metrics")),
			"snmp_target":          model.LabelValue(t.Target),
			"__param_target":       model.LabelValue(t.Target),
		})
	}

	return []*targetgroup.Group{group}
}

func (sh *snmpHandler) ScrapeConfigs(sd discovery.Configs) []*autoscrape.ScrapeConfig {
	if !*sh.cfg.Common.Autoscrape.Enable {
		return nil
	}
	name := sh.cfg.Name()
	cfg := config.DefaultScrapeConfig
	cfg.JobName = fmt.Sprintf("%s/%s", name, name)
	cfg.Scheme = sh.cfg.globals.AgentBaseURL.Scheme
	cfg.HTTPClientConfig = sh.cfg.globals.SubsystemOpts.ClientConfig
	cfg.ServiceDiscoveryConfigs = sd
	cfg.ScrapeInterval = sh.cfg.Common.Autoscrape.ScrapeInterval
	cfg.ScrapeTimeout = sh.cfg.Common.Autoscrape.ScrapeTimeout
	cfg.RelabelConfigs = sh.cfg.Common.Autoscrape.RelabelConfigs
	cfg.MetricRelabelConfigs = sh.cfg.Common.Autoscrape.MetricRelabelConfigs

	return []*autoscrape.ScrapeConfig{{
		Instance: sh.cfg.Common.Autoscrape.MetricsInstance,
		Config:   cfg,
	}}
}

func (sh *snmpHandler) Handler(prefix string) (http.Handler, error) {
	r := mux.NewRouter()
	r.Handle(path.Join(prefix, "metrics"), sh.createHandler(sh.cfg.SnmpTargets))

	return r, nil
}

// Static typecheck tests
var (
	_ integrations.Integration        = (*snmpHandler)(nil)
	_ integrations.HTTPIntegration    = (*snmpHandler)(nil)
	_ integrations.MetricsIntegration = (*snmpHandler)(nil)
)

func (sh *snmpHandler) RunIntegration(ctx context.Context) error {
	<-ctx.Done()
	return nil
}

func (sh *snmpHandler) createHandler(targets []SNMPTarget) http.HandlerFunc {
	snmpTargets := make(map[string]SNMPTarget)
	for _, target := range targets {
		snmpTargets[target.Name] = target
	}

	return func(w http.ResponseWriter, r *http.Request) {
		logger := sh.log
		query := r.URL.Query()
		targetName := query.Get("target")

		var target string
		if len(query["target"]) != 1 || targetName == "" {
			http.Error(w, "'target' parameter must be specified once", 400)
			v1snmp.SnmpRequestErrors.Inc()
			return
		}

		t, ok := snmpTargets[targetName]
		if ok {
			target = t.Target
		} else {
			target = targetName
		}

		var moduleName string
		if query.Has("module") {
			if len(query["module"]) > 1 {
				http.Error(w, "'module' parameter must only be specified once", 400)
				v1snmp.SnmpRequestErrors.Inc()
				return
			}
			moduleName = query.Get("module")
		} else {
			moduleName = t.Module
		}

		if moduleName == "" {
			moduleName = "if_mib"
		}

		module, ok := (*sh.modules)[moduleName]
		if !ok {
			http.Error(w, fmt.Sprintf("Unknown module '%s'", moduleName), 400)
			v1snmp.SnmpRequestErrors.Inc()
			return
		}

		// override module connection details with custom walk params if provided
		var walkParams string
		if query.Has("walk_params") {
			if len(query["walk_params"]) > 1 {
				http.Error(w, "'walk_params' parameter must only be specified once", 400)
				v1snmp.SnmpRequestErrors.Inc()
				return
			}
			walkParams = query.Get("walk_params")
		} else {
			walkParams = t.WalkParams
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
				v1snmp.SnmpRequestErrors.Inc()
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
		v1snmp.SnmpDuration.WithLabelValues(moduleName).Observe(duration)
		level.Debug(logger).Log("msg", "Finished scrape", "duration_seconds", duration)
	}
}

func (sh *snmpHandler) handler(w http.ResponseWriter, r *http.Request) {
	logger := sh.log

	query := r.URL.Query()

	target := query.Get("target")
	if len(query["target"]) != 1 || target == "" {
		http.Error(w, "'target' parameter must be specified once", 400)
		v1snmp.SnmpRequestErrors.Inc()
		return
	}

	moduleName := query.Get("module")
	if len(query["module"]) > 1 {
		http.Error(w, "'module' parameter must only be specified once", 400)
		v1snmp.SnmpRequestErrors.Inc()
		return
	}
	if moduleName == "" {
		moduleName = "if_mib"
	}

	module, ok := (*sh.modules)[moduleName]
	if !ok {
		http.Error(w, fmt.Sprintf("Unknown module '%s'", moduleName), 400)
		v1snmp.SnmpRequestErrors.Inc()
		return
	}

	// override module connection details with custom walk params if provided
	walkParams := query.Get("walk_params")
	if len(query["walk_params"]) > 1 {
		http.Error(w, "'walk_params' parameter must only be specified once", 400)
		v1snmp.SnmpRequestErrors.Inc()
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
			v1snmp.SnmpRequestErrors.Inc()
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
	v1snmp.SnmpDuration.WithLabelValues(moduleName).Observe(duration)
	level.Debug(logger).Log("msg", "Finished scrape", "duration_seconds", duration)
}

func (sh snmpHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	sh.handler(w, r)
}
