package tempo

import (
	"fmt"

	"github.com/go-kit/kit/log"
	"github.com/grafana/agent/pkg/build"
	"github.com/spf13/viper"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config"
	"go.opentelemetry.io/collector/config/configmodels"
	"go.opentelemetry.io/collector/exporter/otlpexporter"
	"go.opentelemetry.io/collector/extension/healthcheckextension"
	"go.opentelemetry.io/collector/extension/pprofextension"
	"go.opentelemetry.io/collector/extension/zpagesextension"
	"go.opentelemetry.io/collector/processor/attributesprocessor"
	"go.opentelemetry.io/collector/processor/batchprocessor"
	"go.opentelemetry.io/collector/processor/filterprocessor"
	"go.opentelemetry.io/collector/processor/memorylimiter"
	"go.opentelemetry.io/collector/processor/queuedprocessor"
	"go.opentelemetry.io/collector/processor/resourceprocessor"
	"go.opentelemetry.io/collector/processor/samplingprocessor/probabilisticsamplerprocessor"
	"go.opentelemetry.io/collector/processor/samplingprocessor/tailsamplingprocessor"
	"go.opentelemetry.io/collector/processor/spanprocessor"
	"go.opentelemetry.io/collector/receiver/jaegerreceiver"
	"go.opentelemetry.io/collector/receiver/otlpreceiver"
	"go.opentelemetry.io/collector/receiver/zipkinreceiver"
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

	info := service.ApplicationStartInfo{
		ExeName:  "grafana-agent",
		LongName: "Grafana Agent",
		Version:  build.Version,
		GitHash:  build.Revision,
	}

	cfgFactory := func(v *viper.Viper, factories config.Factories) (*configmodels.Config, error) {
		return nil, nil
	}

	componentFactories, err := tracingFactories()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize factories %w", err)
	}

	svc, err := service.New(service.Parameters{
		ApplicationStartInfo: info,
		Factories:            componentFactories,
		ConfigFactory:        service.ConfigFactory(cfgFactory),
	})
	if err != nil {
		return nil, fmt.Errorf("unable to create OpenTelemetry Collector service %w", err)
	}

	// jpe only allow trace pipelines?

	// jpe async start?  ruh-roh
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

func tracingFactories() (config.Factories, error) {
	extensions, err := component.MakeExtensionFactoryMap(
		&healthcheckextension.Factory{},
		&pprofextension.Factory{},
		&zpagesextension.Factory{},
	)
	if err != nil {
		return config.Factories{}, err
	}

	receivers, err := component.MakeReceiverFactoryMap(
		jaegerreceiver.NewFactory(),
		&zipkinreceiver.Factory{},
		otlpreceiver.NewFactory(),
	)
	if err != nil {
		return config.Factories{}, err
	}

	exporters, err := component.MakeExporterFactoryMap(
		&otlpexporter.Factory{},
	)
	if err != nil {
		return config.Factories{}, err
	}

	processors, err := component.MakeProcessorFactoryMap(
		attributesprocessor.NewFactory(),
		resourceprocessor.NewFactory(),
		queuedprocessor.NewFactory(),
		batchprocessor.NewFactory(),
		memorylimiter.NewFactory(),
		&tailsamplingprocessor.Factory{},
		&probabilisticsamplerprocessor.Factory{},
		spanprocessor.NewFactory(),
		filterprocessor.NewFactory(),
	)
	if err != nil {
		return config.Factories{}, err
	}

	return config.Factories{
		Extensions: extensions,
		Receivers:  receivers,
		Processors: processors,
		Exporters:  exporters,
	}, nil
}
