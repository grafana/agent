package host_info

import (
	"fmt"
	"time"

	"go.opentelemetry.io/collector/component"
)

// Config defines the configuration options for the host_info connector.
type Config struct {
	// HostIdentifiers defines the list of resource attributes used to derive
	// a unique `grafana.host.id` value. In most cases, this should be [ "host.id" ]
	HostIdentifiers      []string      `mapstructure:"host_identifiers"`
	MetricsFlushInterval time.Duration `mapstructure:"metrics_flush_interval"`
}

var _ component.ConfigValidator = (*Config)(nil)

// Validate checks if the configuration is valid
func (c Config) Validate() error {
	if len(c.HostIdentifiers) == 0 {
		return fmt.Errorf("at least one host identifier is required")
	}

	if c.MetricsFlushInterval > 5*time.Minute || c.MetricsFlushInterval < 15*time.Second {
		return fmt.Errorf("%q is not a valid flush interval", c.MetricsFlushInterval)
	}

	return nil
}
