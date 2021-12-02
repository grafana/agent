package metricsutils

import (
	"context"
	"net/http"
	"path"

	"github.com/go-kit/log"
	"github.com/gorilla/mux"
	"github.com/grafana/agent/pkg/integrations/v2"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/config"
	"github.com/prometheus/prometheus/discovery"
	"github.com/prometheus/prometheus/discovery/targetgroup"
)

// NewMetricsHandlerIntegration returns a integrations.MetricsIntegration which
// will expose a /metrics endpoint for h.
func NewMetricsHandlerIntegration(
	_ log.Logger,
	c integrations.Config, common CommonConfig,
	globals integrations.Globals,
	h http.Handler,
) (integrations.MetricsIntegration, error) {
	if !common.Enabled {
		return nil, integrations.ErrDisabled
	}
	id, err := c.Identifier(globals)
	if err != nil {
		return nil, err
	}
	return &metricsHandlerIntegration{
		integrationName: c.Name(),
		instanceID:      id,
		common:          common,
		globals:         globals,
		handler:         h,
	}, nil
}

type metricsHandlerIntegration struct {
	integrationName, instanceID string

	common  CommonConfig
	globals integrations.Globals
	handler http.Handler
}

// Static typecheck tests
var (
	_ integrations.Integration        = (*metricsHandlerIntegration)(nil)
	_ integrations.HTTPIntegration    = (*metricsHandlerIntegration)(nil)
	_ integrations.MetricsIntegration = (*metricsHandlerIntegration)(nil)
)

// RunIntegration implements Integration.
func (i *metricsHandlerIntegration) RunIntegration(ctx context.Context) error {
	<-ctx.Done()
	return ctx.Err()
}

// Handler implements HTTPIntegration.
func (i *metricsHandlerIntegration) Handler(prefix string) (http.Handler, error) {
	r := mux.NewRouter()
	r.Handle(path.Join(prefix, "metrics"), promhttp.Handler())
	return r, nil
}

// Targets implements MetricsIntegration.
func (i *metricsHandlerIntegration) Targets(prefix string) []*targetgroup.Group {
	integrationNameValue := model.LabelValue("integrations/" + i.integrationName)

	return []*targetgroup.Group{{
		Targets: []model.LabelSet{{
			model.AddressLabel:     model.LabelValue(i.globals.AgentBaseURL.Host),
			model.MetricsPathLabel: model.LabelValue(path.Join(prefix, "metrics")),
		}},
		Labels: model.LabelSet{
			model.InstanceLabel: model.LabelValue(i.instanceID),
			model.JobLabel:      integrationNameValue,
			"agent_hostname":    model.LabelValue(i.globals.AgentIdentifier),

			// Meta labels that can be used during SD.
			// __meta_agent_integration_selfscrape
			"__meta_agent_integration_name":       model.LabelValue(i.integrationName),
			"__meta_agent_integration_instance":   model.LabelValue(i.instanceID),
			"__meta_agent_integration_selfscrape": model.LabelValue(boolToString(i.common.ScrapeIntegration, "")),
		},
		Source: i.integrationName,
	}}
}

func boolToString(b *bool, def string) string {
	switch {
	case b == nil:
		return def
	case *b:
		return "1"
	default:
		return "0"
	}
}

// ScrapeConfigs implements MetricsIntegration.
func (i *metricsHandlerIntegration) ScrapeConfigs(sd discovery.Configs) []*config.ScrapeConfig {
	if i.common.ScrapeIntegration != nil && !*i.common.ScrapeIntegration {
		return nil
	}

	cfg := config.DefaultScrapeConfig
	cfg.JobName = i.integrationName
	cfg.Scheme = i.globals.AgentBaseURL.Scheme
	cfg.HTTPClientConfig = i.globals.AgentHTTPClientConfig
	cfg.ServiceDiscoveryConfigs = sd
	if i.common.ScrapeInterval != 0 {
		cfg.ScrapeInterval = model.Duration(i.common.ScrapeInterval)
	}
	if i.common.ScrapeTimeout != 0 {
		cfg.ScrapeTimeout = model.Duration(i.common.ScrapeTimeout)
	}
	cfg.RelabelConfigs = i.common.RelabelConfigs
	cfg.MetricRelabelConfigs = i.common.MetricRelabelConfigs

	return []*config.ScrapeConfig{&cfg}
}
