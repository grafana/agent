package tempo

import (
	"github.com/go-kit/kit/log"
)

// Config controls the configuration of the Tempo log scraper.
type Config struct {
	// Whether the Tempo subsystem should be enabled.
	Enabled bool `yaml:"-"`

	// OpenTelemetry Collector configuration: https://github.com/open-telemetry/opentelemetry-collector/blob/master/docs/design.md
	TracingPipelines map[string]interface{} `yaml:",inline"`
}

// UnmarshalYAML implements yaml.Unmarshaler.
func (c *Config) UnmarshalYAML(unmarshal func(interface{}) error) error {
	// If the Config is unmarshaled, it's present in the config and should be
	// enabled.
	c.Enabled = true

	type plain Config
	return unmarshal((*plain)(c))
}

// Tempo wraps the OpenTelemetry collector to enablet tracing pipelines
type Tempo struct {
}

// New creates and starts Loki log collection.
func New(c Config, l log.Logger) (*Tempo, error) {
	return &Tempo{}, nil
}

// Stop stops the OpenTelemetry collector subsystem
func (t *Tempo) Stop() {

}
