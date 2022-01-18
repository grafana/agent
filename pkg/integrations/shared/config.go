// Package config provides common configuration structs config among
// implementations of integrations.Integration.
package shared

import (
	"time"

	"github.com/prometheus/prometheus/pkg/relabel"
)

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

type V1IntegrationConfig interface {
	Cfg() Config
	Cmn() Common
}

type V1Integrations interface {
	ActiveConfigs() []Config
}

// Type determines a specific type of integration.
type Type int

const (
	// TypeSingleton is an integration that can only be defined exactly once in
	// the config, unmarshaled through "<integration name>"
	TypeSingleton Type = iota

	// TypeMultiplex is an integration that can only be defined through an array,
	// unmarshaled through "<integration name>_configs"
	TypeMultiplex

	// TypeEither is an integration that can be unmarshaled either as a singleton
	// or as an array, but not both.
	//
	// Deprecated. Use either TypeSingleton or TypeMultiplex for new integrations.
	TypeEither
)
