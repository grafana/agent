package ssl_exporter

import (
	"context"
	"fmt"
	"net/http"
	"path"
	"time"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/gorilla/mux"
	"github.com/grafana/agent/pkg/integrations/v2"
	"github.com/grafana/agent/pkg/integrations/v2/autoscrape"
	"github.com/grafana/agent/pkg/integrations/v2/metricsutils"
	ssl_config "github.com/jamesalbert/ssl_exporter/config"
	"github.com/jamesalbert/ssl_exporter/prober"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/config"
	"github.com/prometheus/prometheus/discovery"
	"github.com/prometheus/prometheus/discovery/targetgroup"
)

type sslHandler struct {
	cfg     *Config
	modules *ssl_config.Config
	log     log.Logger
}

func (sh *sslHandler) Targets(ep integrations.Endpoint) []*targetgroup.Group {
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

	for _, t := range sh.cfg.SSLTargets {
		group.Targets = append(group.Targets, model.LabelSet{
			model.AddressLabel:     model.LabelValue(ep.Host),
			model.MetricsPathLabel: model.LabelValue(path.Join(ep.Prefix, "metrics")),
			"ssl_target":           model.LabelValue(t.Target),
			"__param_target":       model.LabelValue(t.Target),
			"ssl_module":           model.LabelValue(t.Module),
			"__param_module":       model.LabelValue(t.Module),
		})
	}

	return []*targetgroup.Group{group}
}

func (sh *sslHandler) ScrapeConfigs(sd discovery.Configs) []*autoscrape.ScrapeConfig {
	if !*sh.cfg.Common.Autoscrape.Enable {
		return nil
	}
	name := sh.cfg.Name()
	cfg := config.DefaultScrapeConfig
	cfg.JobName = fmt.Sprintf("%s/%s", name, name)
	cfg.Scheme = sh.cfg.globals.AgentBaseURL.Scheme
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

func (sh *sslHandler) Handler(prefix string) (http.Handler, error) {
	r := mux.NewRouter()
	r.Handle(path.Join(prefix, "metrics"), sh.createHandler(sh.cfg.SSLTargets))

	return r, nil
}

// Static typecheck tests
var (
	_ integrations.Integration        = (*sslHandler)(nil)
	_ integrations.HTTPIntegration    = (*sslHandler)(nil)
	_ integrations.MetricsIntegration = (*sslHandler)(nil)
)

func (sh *sslHandler) RunIntegration(ctx context.Context) error {
	<-ctx.Done()
	return nil
}

func (sh *sslHandler) createHandler(targets []SSLTarget) http.HandlerFunc {
	SSLTargets := make(map[string]SSLTarget)
	for _, target := range targets {
		SSLTargets[target.Name] = target
	}

	return func(w http.ResponseWriter, r *http.Request) {
		logger := sh.log
		logger = level.NewFilter(logger, level.AllowDebug())
		query := r.URL.Query()

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
}
