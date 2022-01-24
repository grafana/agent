package app_o11y_exporter //nolint:golint

import (
	"context"
	"fmt"
	"net/http"
	"path"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
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

// Config structs controls the configuration of the app o11y
// integration
type Config struct {
	ExporterConfig config.AppExporterConfig `yaml:",inline"`
	Common         common.MetricsConfig     `yaml:",inline"`
}

// ApplyDefaults applies runtime-specific defaults to c.
func (c *Config) ApplyDefaults(globals integrations.Globals) error {
	c.Common.ApplyDefaults(globals.SubsystemOpts.Metrics.Autoscrape)
	if id, err := c.Identifier(globals); err == nil {
		c.Common.InstanceKey = &id
	}
	return nil
}

// Name returns the name of the integration that this config represents
func (c *Config) Name() string { return "app_o11y_exporter" }

// Identifier uniquely identifies the app o11y integration
func (c *Config) Identifier(globals integrations.Globals) (string, error) {
	if c.Common.InstanceKey != nil {
		return *c.Common.InstanceKey, nil
	}
	return globals.AgentIdentifier, nil
}

type appo11yIntegration struct {
	integrationName, instanceID string
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

// NewIntegration converts this config into an instance of an integratin
func (c *Config) NewIntegration(l log.Logger, globals integrations.Globals) (integrations.Integration, error) {
	id, err := c.Identifier(globals)
	if err != nil {
		return nil, err
	}

	reg := prometheus.NewRegistry()

	if err != nil {
		return nil, err
	}

	lokiExporter := exporters.NewLokiExporter(
		l,
		exporters.LokiExporterConfig{
			LokiInstance:     globals.Logs.Instance(c.ExporterConfig.LogsInstance),
			Labels:           c.ExporterConfig.LogsLabels,
			SendEntryTimeout: c.ExporterConfig.LogsSendTimeout,
		},
	)

	var exp = []exporters.AppReceiverExporter{
		// Loki
		lokiExporter,
	}

	receiver := receiver.NewAppReceiver(c.ExporterConfig, exp)

	for _, e := range exp {
		err = e.Init()
		if err != nil {
			return nil, err
		}
	}

	return &appo11yIntegration{
		logger:          l,
		integrationName: c.Name(),
		instanceID:      id,
		common:          c.Common,
		globals:         globals,
		conf:            c.ExporterConfig,
		receiver:        receiver,
		exporters:       exp,
		reg:             reg,
	}, nil
}

// Handler is the http endpoint for exposing Prometheus metrics
func (i *appo11yIntegration) Handler(prefix string) (http.Handler, error) {
	r := mux.NewRouter()
	r.Handle(path.Join(prefix, "metrics"), promhttp.HandlerFor(i.reg, promhttp.HandlerOpts{}))
	return r, nil
}

// Targets implements MetricsIntegration
func (i *appo11yIntegration) Targets(ep integrations.Endpoint) []*targetgroup.Group {
	integrationNameValue := model.LabelValue("integrations/" + i.integrationName)

	group := &targetgroup.Group{
		Labels: model.LabelSet{
			model.InstanceLabel: model.LabelValue(i.instanceID),
			model.JobLabel:      integrationNameValue,
			"agent_hostname":    model.LabelValue(i.globals.AgentIdentifier),

			// Meta labels that can be used during SD.
			"__meta_agent_integration_name":     model.LabelValue(i.integrationName),
			"__meta_agent_integration_instance": model.LabelValue(i.instanceID),
		},
		Source: fmt.Sprintf("%s/%s", i.integrationName, i.instanceID),
	}

	group.Targets = append(group.Targets, model.LabelSet{
		model.AddressLabel:     model.LabelValue(ep.Host),
		model.MetricsPathLabel: model.LabelValue(path.Join(ep.Prefix, "/metrics")),
	})

	return []*targetgroup.Group{group}
}

// ScrapeConfigs implements MetricsIntegration
func (i *appo11yIntegration) ScrapeConfigs(sd discovery.Configs) []*autoscrape.ScrapeConfig {
	if !*i.common.Autoscrape.Enable {
		return nil
	}

	cfg := promConfig.DefaultScrapeConfig
	cfg.JobName = fmt.Sprintf("%s/%s", i.integrationName, i.instanceID)
	cfg.Scheme = i.globals.AgentBaseURL.Scheme
	cfg.HTTPClientConfig = i.globals.SubsystemOpts.ClientConfig
	cfg.ServiceDiscoveryConfigs = sd
	cfg.ScrapeInterval = i.common.Autoscrape.ScrapeInterval
	cfg.ScrapeTimeout = i.common.Autoscrape.ScrapeTimeout
	cfg.RelabelConfigs = i.common.Autoscrape.RelabelConfigs
	cfg.MetricRelabelConfigs = i.common.Autoscrape.MetricRelabelConfigs

	return []*autoscrape.ScrapeConfig{{
		Instance: i.common.Autoscrape.MetricsInstance,
		Config:   cfg,
	}}
}

// RunIntegration implements Integratin
func (i *appo11yIntegration) RunIntegration(ctx context.Context) error {
	r := mux.NewRouter()
	r.Handle("/collect", i.receiver.ReceiverHandler(i.logger))

	srv := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", i.conf.Server.Host, i.conf.Server.Port),
		Handler: r,
	}
	errChan := make(chan error)

	go func() {
		level.Info(i.logger).Log("msg", "starting app o11y receiver", "host", i.conf.Server.Host, "port", i.conf.Server.Port)
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			errChan <- err
		}
	}()

	select {
	case <-ctx.Done():
		if err := srv.Shutdown(ctx); err != nil {
			return err
		}
	case err := <-errChan:
		close(errChan)
		return err
	}

	return nil
}

func init() {
	integrations.Register(&Config{}, integrations.TypeMultiplex)
}
