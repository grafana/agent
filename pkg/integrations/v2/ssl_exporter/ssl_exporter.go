package ssl_exporter

import (
	"context"
	"fmt"
	"net/http"
	"path"
	"sort"
	"sync"

	"github.com/prometheus/common/version"

	"github.com/gorilla/mux"

	"github.com/grafana/agent/pkg/integrations/v2/autoscrape"
	"github.com/prometheus/prometheus/config"
	"github.com/prometheus/prometheus/discovery"

	"github.com/grafana/agent/pkg/integrations/v2/metricsutils"
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/discovery/targetgroup"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	sslv1 "github.com/grafana/agent/pkg/integrations/ssl_exporter"
	"github.com/grafana/agent/pkg/integrations/v2"
	"github.com/prometheus/client_golang/prometheus"
	ssl_config "github.com/ribbybibby/ssl_exporter/v2/config"
	"github.com/ribbybibby/ssl_exporter/v2/prober"
)

func init() {
	integrations.Register(&Config{}, integrations.TypeMultiplex)
}

type Exporter struct {
	sync.Mutex
	probeSuccess prometheus.Gauge
	proberType   *prometheus.GaugeVec

	options   Options
	namespace string
	log       log.Logger
	cfg       *Config
}

type Options struct {
	Namespace   string
	MetricsPath string
	ProbePath   string
	Registry    *prometheus.Registry
	SSLTargets  []SSLTarget
	SSLConfig   *ssl_config.Config
	Logger      log.Logger
	Name        string
}

// Static typecheck tests
var (
	_ integrations.Integration        = (*Exporter)(nil)
	_ integrations.HTTPIntegration    = (*Exporter)(nil)
	_ integrations.MetricsIntegration = (*Exporter)(nil)
)

func NewSSLExporter(opts Options, cfg *Config) (*Exporter, error) {
	e := &Exporter{
		options:   opts,
		namespace: opts.Namespace,
		probeSuccess: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: prometheus.BuildFQName(opts.Namespace, "", "probe_success"),
				Help: "If the probe was a success",
			},
		),
		proberType: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: prometheus.BuildFQName(opts.Namespace, "", "prober"),
				Help: "The prober used by the exporter to connect to the target",
			},
			[]string{"prober"},
		),
		cfg: cfg,
	}

	return e, nil
}

func (e *Exporter) RunIntegration(ctx context.Context) error {
	<-ctx.Done()
	return nil
}

func (e *Exporter) Handler(prefix string) (http.Handler, error) {
	r := mux.NewRouter()
	r.Handle(path.Join(prefix, "metrics"), e)

	return r, nil
}

func (e *Exporter) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	query := request.URL.Query()

	targetName := query.Get("target")

	target := &SSLTarget{
		Name:   targetName,
		Target: targetName,
		Module: e.options.SSLConfig.DefaultModule,
	}
	module, found := e.options.SSLConfig.Modules[target.Module]
	if !found {
		level.Error(e.log).Log("msg", fmt.Sprintf("Unknown module '%s'", target.Module))
		return
	}

	probeFunc, found := prober.Probers[module.Prober]
	ctx := context.Background()
	// set high-level metric not collected in the prober
	registry := prometheus.NewRegistry()
	registry.MustRegister(e.probeSuccess, e.proberType)
	registry.MustRegister(version.NewCollector("ssl"))

	err := probeFunc(ctx, e.log, target.Target, module, registry)
	if err != nil {
		level.Error(e.log).Log("msg", fmt.Sprintf("error probing module '%s'", target.Module))
		return
	}

	// Delegate http serving to Prometheus client library, which will call collector.Collect.
	h := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
	h.ServeHTTP(writer, request)
}

func (e *Exporter) ScrapeConfigs(sd discovery.Configs) []*autoscrape.ScrapeConfig {
	if !*e.cfg.Common.Autoscrape.Enable {
		return nil
	}
	name := e.cfg.Name()
	cfg := config.DefaultScrapeConfig
	cfg.JobName = fmt.Sprintf("%s/%s", name, name)
	cfg.Scheme = e.cfg.globals.AgentBaseURL.Scheme
	cfg.ServiceDiscoveryConfigs = sd
	cfg.ScrapeInterval = e.cfg.Common.Autoscrape.ScrapeInterval
	cfg.ScrapeTimeout = e.cfg.Common.Autoscrape.ScrapeTimeout
	cfg.RelabelConfigs = e.cfg.Common.Autoscrape.RelabelConfigs
	cfg.MetricRelabelConfigs = e.cfg.Common.Autoscrape.MetricRelabelConfigs

	return []*autoscrape.ScrapeConfig{{
		Instance: e.cfg.Common.Autoscrape.MetricsInstance,
		Config:   cfg,
	}}
}

func (e *Exporter) Targets(ep integrations.Endpoint) []*targetgroup.Group {
	integrationNameValue := model.LabelValue("integrations/" + e.cfg.Name())

	group := &targetgroup.Group{
		Labels: model.LabelSet{
			model.InstanceLabel: model.LabelValue(""),
			model.JobLabel:      integrationNameValue,
			"agent_hostname":    model.LabelValue(e.cfg.globals.AgentIdentifier),

			// Meta labels that can be used during SD.
			"__meta_agent_integration_name":       model.LabelValue(e.cfg.Name()),
			"__meta_agent_integration_instance":   model.LabelValue(e.cfg.Name()),
			"__meta_agent_integration_autoscrape": model.LabelValue(metricsutils.BoolToString(*e.cfg.Common.Autoscrape.Enable)),
		},
		Source: fmt.Sprintf("%s/%s", e.cfg.Name(), e.cfg.Name()),
	}

	for _, lbl := range e.cfg.Common.ExtraLabels {
		group.Labels[model.LabelName(lbl.Name)] = model.LabelValue(lbl.Value)
	}

	for _, t := range e.cfg.SSLTargets {
		group.Targets = append(group.Targets, model.LabelSet{
			model.AddressLabel:     model.LabelValue(ep.Host),
			model.MetricsPathLabel: model.LabelValue(path.Join(ep.Prefix, "metrics")),
			"snmp_target":          model.LabelValue(t.Target),
			"__param_target":       model.LabelValue(t.Target),
		})
	}

	return []*targetgroup.Group{group}
}

func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	for _, desc := range sslv1.Descs {
		ch <- desc
	}
}

func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
	e.Lock()
	defer e.Unlock()

	logger := e.options.Logger

	for _, target := range e.options.SSLTargets {
		ctx := context.Background()

		var moduleName string
		if target.Module != "" {
			moduleName = e.options.SSLConfig.DefaultModule
			if moduleName == "" {
				level.Error(logger).Log("msg", "Module parameter must be set")
				continue
			}
		}

		module, ok := e.options.SSLConfig.Modules[target.Module]
		if !ok {
			level.Error(logger).Log("msg", fmt.Sprintf("Unknown module '%s'", target.Module))
			continue
		}

		probeFunc, ok := prober.Probers[module.Prober]
		if !ok {
			level.Error(logger).Log("msg", fmt.Sprintf("Unknown prober %q", module.Prober))
			continue
		}

		e.options.Registry = prometheus.NewRegistry()
		e.options.Registry.MustRegister(e.probeSuccess, e.proberType)
		e.options.Registry.MustRegister(version.NewCollector("ssl"))
		e.proberType.WithLabelValues(module.Prober).Set(1)

		// set high-level metric not collected in the prober
		err := probeFunc(ctx, logger, target.Target, module, e.options.Registry)
		if err != nil {
			level.Error(logger).Log("msg", err)
			e.probeSuccess.Set(0)
		} else {
			e.probeSuccess.Set(1)
		}

		// gather all the metrics we've collected in the prober
		metricFams, err := e.options.Registry.Gather()
		if err != nil {
			level.Error(logger).Log("msg", err)
			continue
		}
		for _, mf := range metricFams {
			for _, m := range mf.Metric {
				// get desc from name
				desc, ok := sslv1.Descs[*mf.Name]
				if !ok {
					level.Error(logger).Log("msg", fmt.Sprintf("Unknown metric %q", *mf.Name))
					continue
				}

				// ensure label order
				sort.Slice(m.Label, func(i, j int) bool {
					iPrec := sslv1.LabelOrder[*m.Label[i].Name]
					jPrec := sslv1.LabelOrder[*m.Label[j].Name]
					return iPrec < jPrec
				})
				labelValues := []string{}
				for _, l := range m.Label {
					labelValues = append(labelValues, *l.Value)
				}

				// create prometheus metric
				metric, err := prometheus.NewConstMetric(desc, prometheus.GaugeValue, *m.Gauge.Value, labelValues...)
				if err != nil {
					level.Error(logger).Log("msg", err)
					continue
				}
				ch <- metric
			}
		}
	}
}
