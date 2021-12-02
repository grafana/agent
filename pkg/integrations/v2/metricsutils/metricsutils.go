// Package metricsutils provides utilities for creating metrics integrations.
package metricsutils

import (
	"time"

	"github.com/prometheus/prometheus/pkg/relabel"
)

// DefaultCommonConfig is the default common settings for metrics integrations.
var DefaultCommonConfig = CommonConfig{
	// Integrations are enabled by default when defined.
	Enabled: true,
}

// CommonConfig is a set of common options shared by all integrations. It should be
// utilised by an integration's config by inlining the common options:
//
//   type IntegrationConfig struct {
//     Common config.CommonConfig `yaml:",inline"`
//   }
type CommonConfig struct {
	// Enabled controls whether a present integration should run.
	//
	// Enabled is DEPRECATED and will be removed in a future version. Users
	// should change to removing or commenting out integrations instead of
	// using `enabled: false` to prevent it from running.
	Enabled bool `yaml:"enabled,omitempty"`

	InstanceKey          *string           `yaml:"instance,omitempty"`
	ScrapeIntegration    *bool             `yaml:"scrape_integration,omitempty"`
	ScrapeInterval       time.Duration     `yaml:"scrape_interval,omitempty"`
	ScrapeTimeout        time.Duration     `yaml:"scrape_timeout,omitempty"`
	RelabelConfigs       []*relabel.Config `yaml:"relabel_configs,omitempty"`
	MetricRelabelConfigs []*relabel.Config `yaml:"metric_relabel_configs,omitempty"`
}

// UnmarshalYAML implements yaml.Unmarshaler.
func (c *CommonConfig) UnmarshalYAML(unmarshal func(interface{}) error) error {
	*c = DefaultCommonConfig

	type commonConfig CommonConfig
	return unmarshal((*commonConfig)(c))
}
