// Package config provides common configuration structs shared among
// implementations of integrations.Integration.
package config

import (
	"time"

	"github.com/prometheus/prometheus/pkg/relabel"
)

// DefaultCommon is the default common settings for all integrations.
var DefaultCommon = Common{
	// Integrations are enabled by default when defined.
	Enabled: true,
}

// Common is a set of common options shared by all integrations. It should be
// utilised by an integration's config by inlining the common options:
//
//   type IntegrationConfig struct {
//     Common config.Common `yaml:",inline"`
//   }
type Common struct {
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

	// TODO(rfratto): remove this field.
	WALTruncateFrequency time.Duration `yaml:"wal_truncate_frequency,omitempty"`
}

// UnmarshalYAML implements yaml.Unmarshaler.
func (c *Common) UnmarshalYAML(unmarshal func(interface{}) error) error {
	*c = DefaultCommon

	type common Common
	return unmarshal((*common)(c))
}

// ScrapeConfig is a subset of options used by integrations to inform how samples
// should be scraped. It is utilized by the integrations.Manager to define a full
// Prometheus-compatible ScrapeConfig.
type ScrapeConfig struct {
	// JobName should be a unique name indicating the collection of samples to be
	// scraped. It will be prepended by "integrations/" when used by the integrations
	// manager.
	JobName string

	// MetricsPath is the path relative to the integration where metrics are exposed.
	// It should match a route added to the router provided in Integration.RegisterRoutes.
	// The path will be prepended by "/integrations/<integration name>" when read by
	// the integrations manager.
	MetricsPath string
}
