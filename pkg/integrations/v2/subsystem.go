package integrations

import (
	common_config "github.com/prometheus/common/config"
	"github.com/prometheus/common/model"
	prom_config "github.com/prometheus/prometheus/config"
)

// DefaultSubsystemOptions holds the default settings for a Controller.
var DefaultSubsystemOptions = SubsystemOptions{
	ScrapeIntegrations: true,
}

// SubsystemOptions controls how the integrations subsystem behaves.
type SubsystemOptions struct {
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
