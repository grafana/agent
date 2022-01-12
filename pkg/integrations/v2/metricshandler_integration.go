package integrations

import (
	"context"
	"fmt"
	"net/http"
	"path"

	"github.com/go-kit/log"
	"github.com/gorilla/mux"
	"github.com/grafana/agent/pkg/integrations/v2/autoscrape"
	"github.com/grafana/agent/pkg/integrations/v2/common"
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/config"
	"github.com/prometheus/prometheus/discovery"
	"github.com/prometheus/prometheus/discovery/targetgroup"
)

// NewMetricsHandlerIntegration returns a integrations.MetricsIntegration which
// will expose a /metrics endpoint for handler.
func NewMetricsHandlerIntegration(
	_ log.Logger,
	c Config,
	mc common.MetricsConfig,
	globals Globals,
	h http.Handler,
) (MetricsIntegration, error) {
	id, err := c.Identifier(globals)
	if err != nil {
		return nil, err
	}
	return &MetricsHandlerIntegration{
		IntegrationName: c.Name(),
		InstanceID:      id,

		Common:  mc,
		Globals: globals,
		handler: h,

		targets: []handlerTarget{{MetricsPath: "metrics"}},
	}, nil
}

type MetricsHandlerIntegration struct {
	IntegrationName, InstanceID string

	Common  common.MetricsConfig
	Globals Globals
	handler http.Handler
	targets []handlerTarget

	RunFunc func(ctx context.Context) error
}

type handlerTarget struct {
	// Path relative to handler prefix where metrics are available.
	MetricsPath string
	// Extra labels to inject into the target. Labels here that take precedence
	// over labels with the same name from the generated target group.
	Labels model.LabelSet
}

// Static typecheck tests
var (
	_ Integration        = (*MetricsHandlerIntegration)(nil)
	_ HTTPIntegration    = (*MetricsHandlerIntegration)(nil)
	_ MetricsIntegration = (*MetricsHandlerIntegration)(nil)
)

// RunIntegration implements Integration.
func (i *MetricsHandlerIntegration) RunIntegration(ctx context.Context) error {
	// Call our runFunc if defined (used from integrationShim), otherwise
	// fallback to no-op.
	if i.RunFunc != nil {
		return i.RunFunc(ctx)
	}

	<-ctx.Done()
	return nil
}

// handler implements HTTPIntegration.
func (i *MetricsHandlerIntegration) Handler(prefix string) (http.Handler, error) {
	r := mux.NewRouter()
	r.Handle(path.Join(prefix, "metrics"), i.handler)
	return r, nil
}

// targets implements MetricsIntegration.
func (i *MetricsHandlerIntegration) Targets(ep Endpoint) []*targetgroup.Group {
	integrationNameValue := model.LabelValue("integrations/" + i.IntegrationName)

	group := &targetgroup.Group{
		Labels: model.LabelSet{
			model.InstanceLabel: model.LabelValue(i.InstanceID),
			model.JobLabel:      integrationNameValue,
			"agent_hostname":    model.LabelValue(i.Globals.AgentIdentifier),

			// Meta labels that can be used during SD.
			"__meta_agent_integration_name":       model.LabelValue(i.IntegrationName),
			"__meta_agent_integration_instance":   model.LabelValue(i.InstanceID),
			"__meta_agent_integration_autoscrape": model.LabelValue(boolToString(*i.Common.Autoscrape.Enable)),
		},
		Source: fmt.Sprintf("%s/%s", i.IntegrationName, i.InstanceID),
	}

	for _, t := range i.targets {
		group.Targets = append(group.Targets, model.LabelSet{
			model.AddressLabel:     model.LabelValue(ep.Host),
			model.MetricsPathLabel: model.LabelValue(path.Join(ep.Prefix, t.MetricsPath)),
		}.Merge(t.Labels))
	}

	return []*targetgroup.Group{group}
}

func boolToString(b bool) string {
	switch b {
	case true:
		return "1"
	default:
		return "0"
	}
}

// ScrapeConfigs implements MetricsIntegration.
func (i *MetricsHandlerIntegration) ScrapeConfigs(sd discovery.Configs) []*autoscrape.ScrapeConfig {
	if !*i.Common.Autoscrape.Enable {
		return nil
	}

	cfg := config.DefaultScrapeConfig
	cfg.JobName = fmt.Sprintf("%s/%s", i.IntegrationName, i.InstanceID)
	cfg.Scheme = i.Globals.AgentBaseURL.Scheme
	cfg.HTTPClientConfig = i.Globals.SubsystemOpts.ClientConfig
	cfg.ServiceDiscoveryConfigs = sd
	cfg.ScrapeInterval = i.Common.Autoscrape.ScrapeInterval
	cfg.ScrapeTimeout = i.Common.Autoscrape.ScrapeTimeout
	cfg.RelabelConfigs = i.Common.Autoscrape.RelabelConfigs
	cfg.MetricRelabelConfigs = i.Common.Autoscrape.MetricRelabelConfigs

	return []*autoscrape.ScrapeConfig{{
		Instance: i.Common.Autoscrape.MetricsInstance,
		Config:   cfg,
	}}
}
