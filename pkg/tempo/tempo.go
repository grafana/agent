package tempo

import (
	"fmt"

	"github.com/go-kit/kit/log"
	"github.com/spf13/viper"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config/configmodels"
	"go.opentelemetry.io/collector/service"
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
	svc *service.Application
}

// New creates and starts Loki log collection.
func New(c Config, l log.Logger) (*Tempo, error) {
	info := component.ApplicationStartInfo{
		ExeName:  "jpe",
		LongName: "jpe",
		Version:  "?",
		GitHash:  "?",
	}

	cfgFactory := func(v *viper.Viper, factories component.Factories) (*configmodels.Config, error) {
		return nil, nil
	}

	componentFactories := component.Factories{}

	svc, err := service.New(service.Parameters{
		ApplicationStartInfo: info,
		Factories:            componentFactories,
		ConfigFactory:        service.ConfigFactory(cfgFactory),
	})
	if err != nil {
		return nil, fmt.Errorf("unable to create OpenTelemetry Collector service %w", err)
	}

	err = svc.Start()
	if err != nil {
		return nil, fmt.Errorf("unable to start OpenTelemetry Collector service %w", err)
	}

	return &Tempo{}, nil
}

// Stop stops the OpenTelemetry collector subsystem
func (t *Tempo) Stop() {
	if t.svc != nil {
		// jpe - how to stop.  service doesn't have a way to stop it.  listens to signals channel on its own
	}
}
