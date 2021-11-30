package integrations

import (
	common_config "github.com/prometheus/common/config"
	"github.com/prometheus/common/model"
	prom_config "github.com/prometheus/prometheus/config"
)

// DefaultControllerOptions holds the default settings for a Controller.
var DefaultControllerOptions = ControllerOptions{
	ScrapeIntegrations: true,
}

// ControllerOptions configures a Controller. ControllerOptions implements
// ControllerConfig.
type ControllerOptions struct {
	// When true, scrapes metrics from integrations.
	ScrapeIntegrations bool `yaml:"scrape_integrations,omitempty"`
	// Prometheus RW configs to use for self-scraping integrations.
	PrometheusRemoteWrite []*prom_config.RemoteWriteConfig `yaml:"prometheus_remote_write,omitempty"`

	// Configs are configurations of integration to create. Unmarshaled through
	// the custom UnmarshalYAML method of Controller.
	Configs []Config `yaml:"-"`

	// Extra labels to add for all integration telemetry.
	Labels model.LabelSet `yaml:"labels,omitempty"`

	// Override settings to self-communicate with agent.
	ClientConfig common_config.HTTPClientConfig `yaml:"client_config,omitempty"`
}

// ControllerConfig is an extension of Config used to configure controllers.
type ControllerConfig interface {
	Config
	ControllerOptions() ControllerOptions
}

// Name implements Config. Returns "integrations".
func (o *ControllerOptions) Name() string { return "integrations" }

// Identifier implements Config. Returns "integrations".
func (o *ControllerOptions) Identifier(IntegrationOptions) (string, error) {
	return "integrations", nil
}

// NewIntegration implements Config. Returns a Controller.
func (o *ControllerOptions) NewIntegration(iopts IntegrationOptions) (Integration, error) {
	return NewController(*o, iopts)
}

// ControllerOptions implements ControllerConfig.
func (o *ControllerOptions) ControllerOptions() ControllerOptions { return *o }
