package tempo

import (
	"fmt"

	"github.com/spf13/viper"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config"
	"go.opentelemetry.io/collector/config/configmodels"
	"go.opentelemetry.io/collector/exporter/otlpexporter"
	"go.opentelemetry.io/collector/processor/batchprocessor"
	"go.opentelemetry.io/collector/processor/queuedprocessor"
	"go.opentelemetry.io/collector/receiver/jaegerreceiver"
	"go.opentelemetry.io/collector/receiver/otlpreceiver"
	"go.opentelemetry.io/collector/receiver/zipkinreceiver"
)

// Config controls the configuration of the Tempo trace pipeline.
type Config struct {
	// Whether the Tempo subsystem should be enabled.
	Enabled bool `yaml:"-"`

	RemoteWrite RWConfig `yaml:"remote_write"`

	// Receivers: https://github.com/open-telemetry/opentelemetry-collector/tree/master/receiver
	Receivers map[string]interface{} `yaml:"receivers"`
}

// RWConfig controls the configuration of exporting to Grafana Cloud
type RWConfig struct {
	URL       string                  `yaml:"url"`
	BasicAuth BasicAuthConfig         `yaml:"basic_auth"`
	Batch     *batchprocessor.Config  `yaml:"batch,omitempty"`
	Queue     *queuedprocessor.Config `yaml:"queue,omitempty"`
}

// BasicAuthConfig controls the configuration of basic auth to Grafana cloud
type BasicAuthConfig struct {
	Username     string `yaml:"username"`
	Password     string `yaml:"password"`
	PasswordFile string `yaml:"password_file"`
}

// UnmarshalYAML implements yaml.Unmarshaler.
func (c *Config) UnmarshalYAML(unmarshal func(interface{}) error) error {
	// If the Config is unmarshaled, it's present in the config and should be
	// enabled.
	c.Enabled = true

	type plain Config
	return unmarshal((*plain)(c))
}

func (c *Config) otelConfig() (*configmodels.Config, error) {
	otelMapStructure := map[string]interface{}{}

	// basic auth header
	/*encodedAuth := base64.StdEncoding.EncodeToString([]byte(cfg.TenantID + ":" + cfg.Token))
	otlpConfig := factory.CreateDefaultConfig().(*otlpexporter.Config)

	// config
	otlpConfig.Endpoint = cfg.Endpoint
	otlpConfig.Headers = map[string]string{
		"Authorization": "Basic " + encodedAuth,
	}*/

	// add receivers
	otelMapStructure["receivers"] = c.Receivers

	// now build the otel configmodel from the mapstructure
	v := viper.New()
	err := v.MergeConfigMap(otelMapStructure)
	if err != nil {
		return nil, fmt.Errorf("failed to merge in mapstructure config %w", err)
	}

	factories, err := tracingFactories()
	if err != nil {
		return nil, fmt.Errorf("failed to create factories %w", err)
	}

	otelCfg, err := config.Load(v, factories)
	if err != nil {
		return nil, fmt.Errorf("failed to load OTel config %w", err)
	}

	return otelCfg, nil
}

// tracingFactories() only creates the needed factories.  if we decide to add support for a new
// processor, exporter, receiver we need to add it here
func tracingFactories() (config.Factories, error) {
	extensions, err := component.MakeExtensionFactoryMap()
	if err != nil {
		return config.Factories{}, err
	}

	receivers, err := component.MakeReceiverFactoryMap(
		jaegerreceiver.NewFactory(),
		&zipkinreceiver.Factory{},
		otlpreceiver.NewFactory(), // jpe - opencensus?
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
		queuedprocessor.NewFactory(),
		batchprocessor.NewFactory(),
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
