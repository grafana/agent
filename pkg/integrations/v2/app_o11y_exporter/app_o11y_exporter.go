package app_o11y_exporter

import (
	"context"
	"fmt"
	"net/http"
	"path"

	"github.com/go-kit/log"
	"github.com/gorilla/mux"
	"github.com/grafana/agent/pkg/integrations/v2"
	"github.com/grafana/agent/pkg/integrations/v2/app_o11y_exporter/config"
	"github.com/grafana/agent/pkg/integrations/v2/app_o11y_exporter/exporters"
	"github.com/grafana/agent/pkg/integrations/v2/app_o11y_exporter/receiver"
	"github.com/grafana/agent/pkg/integrations/v2/autoscrape"
	"github.com/grafana/agent/pkg/integrations/v2/common"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/model"
	promConfig "github.com/prometheus/prometheus/config"
	"github.com/prometheus/prometheus/discovery"
	"github.com/prometheus/prometheus/discovery/targetgroup"
)

type Config struct {
	ExporterConfig config.AppExporterConfig `yaml:",inline"`
	Common         common.MetricsConfig     `yaml:"metrics,omitempty"`
}

func (c *Config) ApplyDefaults(globals integrations.Globals) error {
	c.Common.ApplyDefaults(globals.SubsystemOpts.Metrics.Autoscrape)
	if id, err := c.Identifier(globals); err == nil {
		c.Common.InstanceKey = &id
	}
	return nil
}

func (c *Config) Name() string { return "app_o11y_exporter" }

func (c *Config) Identifier(globals integrations.Globals) (string, error) {
	if c.Common.InstanceKey != nil {
		return *c.Common.InstanceKey, nil
	}
	return globals.AgentIdentifier, nil
}

type appo11yIntegration struct {
	integrationName, instanceId string
	globals                     integrations.Globals
	logger                      log.Logger
	conf                        config.AppExporterConfig
	common                      common.MetricsConfig
	receiver                    receiver.AppReceiver
	exporters                   []exporters.AppReceiverExporter
	reg                         *prometheus.Registry
}

// Static typecheck tests
var (
	_ integrations.Integration        = (*appo11yIntegration)(nil)
	_ integrations.HTTPIntegration    = (*appo11yIntegration)(nil)
	_ integrations.MetricsIntegration = (*appo11yIntegration)(nil)
)

func (c *Config) NewIntegration(l log.Logger, globals integrations.Globals) (integrations.Integration, error) {
	id, err := c.Identifier(globals)
	if err != nil {
		return nil, err
	}

	reg := prometheus.NewRegistry()

	lokiExceptionExporter, err := exporters.NewLokiExceptionExporter(
		globals.Logs.Instance(c.ExporterConfig.LogsInstance),
		c.ExporterConfig.SourceMap,
	)

	if err != nil {
		return nil, err
	}

	lokiExporter := exporters.NewLokiExporter(
		globals.Logs.Instance(c.ExporterConfig.LogsInstance),
		exporters.LokiExporterConfig{SendEntryTimeout: 2000},
	)

	measurementsPromExporter := exporters.NewPrometheusMetricsExporter(reg, c.ExporterConfig.Measurements)

	var exp = []exporters.AppReceiverExporter{
		// Logs
		lokiExporter,
		// Exceptions
		lokiExceptionExporter,
		// Measurements
		measurementsPromExporter,
	}

	receiver := receiver.NewAppReceiver(c.ExporterConfig, exp)

	for _, e := range exp {
		e.Init()
	}

	return &appo11yIntegration{
		logger:          l,
		integrationName: c.Name(),
		instanceId:      id,
		common:          c.Common,
		globals:         globals,
		conf:            c.ExporterConfig,
		receiver:        receiver,
		exporters:       exp,
		reg:             reg,
	}, nil
}

func (i *appo11yIntegration) Handler(prefix string) (http.Handler, error) {
	r := mux.NewRouter()
	r.Handle(path.Join(prefix, "metrics"), promhttp.HandlerFor(i.reg, promhttp.HandlerOpts{}))
	return r, nil
}

func (ai *appo11yIntegration) Targets(ep integrations.Endpoint) []*targetgroup.Group {
	integrationNameValue := model.LabelValue("integrations/" + ai.integrationName)

	group := &targetgroup.Group{
		Labels: model.LabelSet{
			model.InstanceLabel: model.LabelValue(ai.instanceId),
			model.JobLabel:      integrationNameValue,
			"agent_hostname":    model.LabelValue(ai.globals.AgentIdentifier),

			// Meta labels that can be used during SD.
			"__meta_agent_integration_name":     model.LabelValue(ai.integrationName),
			"__meta_agent_integration_instance": model.LabelValue(ai.instanceId),
		},
		Source: fmt.Sprintf("%s/%s", ai.integrationName, ai.instanceId),
	}

	group.Targets = append(group.Targets, model.LabelSet{
		model.AddressLabel:     model.LabelValue(ep.Host),
		model.MetricsPathLabel: model.LabelValue(path.Join(ep.Prefix, "/metrics")),
	})

	return []*targetgroup.Group{group}
}

func (ai *appo11yIntegration) ScrapeConfigs(sd discovery.Configs) []*autoscrape.ScrapeConfig {
	if !*ai.common.Autoscrape.Enable {
		return nil
	}

	cfg := promConfig.DefaultScrapeConfig
	cfg.JobName = fmt.Sprintf("%s/%s", ai.integrationName, ai.instanceId)
	cfg.Scheme = ai.globals.AgentBaseURL.Scheme
	cfg.HTTPClientConfig = ai.globals.SubsystemOpts.ClientConfig
	cfg.ServiceDiscoveryConfigs = sd
	cfg.ScrapeInterval = ai.common.Autoscrape.ScrapeInterval
	cfg.ScrapeTimeout = ai.common.Autoscrape.ScrapeTimeout
	cfg.RelabelConfigs = ai.common.Autoscrape.RelabelConfigs
	cfg.MetricRelabelConfigs = ai.common.Autoscrape.MetricRelabelConfigs

	return []*autoscrape.ScrapeConfig{{
		Instance: ai.common.Autoscrape.MetricsInstance,
		Config:   cfg,
	}}
}

func (ai *appo11yIntegration) RunIntegration(ctx context.Context) error {
	r := mux.NewRouter()
	r.Handle("/collect", ai.receiver.ReceiverHandler(&ai.logger))

	srv := &http.Server{
		Addr: fmt.Sprintf("%s:%d", ai.conf.Server.Host, ai.conf.Server.Port),
	}
	go func() {
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			ai.logger.Log("Error on ListenAndServe(): %v", err)
		}
	}()

	<-ctx.Done()
	if err := srv.Shutdown(ctx); err != nil {
		return err
	}

	return nil
}

func init() {
	integrations.Register(&Config{}, integrations.TypeSingleton)
}
