// Package config provides common configuration structs shared among
// implementations of integrations.Integration.
package config

import (
	"time"

	"github.com/prometheus/prometheus/pkg/relabel"
)

// Common is a set of common options shared by all integrations. It should be
// utilised by an integration's config by inlining the common options:
//
//   type IntegrationConfig struct {
//     Common config.Common `yaml:",inline"`
//   }
type Common struct {
	Enabled              bool              `yaml:"enabled,omitempty"`
	InstanceKey          *string           `yaml:"instance,omitempty"`
	ScrapeIntegration    *bool             `yaml:"scrape_integration,omitempty"`
	ScrapeInterval       time.Duration     `yaml:"scrape_interval,omitempty"`
	ScrapeTimeout        time.Duration     `yaml:"scrape_timeout,omitempty"`
	RelabelConfigs       []*relabel.Config `yaml:"relabel_configs,omitempty"`
	MetricRelabelConfigs []*relabel.Config `yaml:"metric_relabel_configs,omitempty"`
	WALTruncateFrequency time.Duration     `yaml:"wal_truncate_frequency,omitempty"`
}
