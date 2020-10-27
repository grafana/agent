package tempo

import (
	"encoding/base64"
	"errors"
	"fmt"
	"io/ioutil"

	"github.com/grafana/agent/pkg/tempo/promsdprocessor"
	prom_config "github.com/prometheus/common/config"
	"github.com/spf13/viper"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config"
	"go.opentelemetry.io/collector/config/configmodels"
	"go.opentelemetry.io/collector/exporter/otlpexporter"
	"go.opentelemetry.io/collector/processor/attributesprocessor"
	"go.opentelemetry.io/collector/processor/batchprocessor"
	"go.opentelemetry.io/collector/processor/queuedprocessor"
	"go.opentelemetry.io/collector/receiver/jaegerreceiver"
	"go.opentelemetry.io/collector/receiver/opencensusreceiver"
	"go.opentelemetry.io/collector/receiver/otlpreceiver"
	"go.opentelemetry.io/collector/receiver/zipkinreceiver"
)

// Config controls the configuration of the Tempo trace pipeline.
type Config struct {
	// Whether the Tempo subsystem should be enabled.
	Enabled bool `yaml:"-"`

	RemoteWrite RWConfig `yaml:"remote_write"`

	// Receivers: https://github.com/open-telemetry/opentelemetry-collector/tree/1405654d4e907b3215cece0ce04e46a6c1576382/receiver
	Receivers map[string]interface{} `yaml:"receivers"`

	// Attributes: https://github.com/open-telemetry/opentelemetry-collector/tree/1405654d4e907b3215cece0ce04e46a6c1576382/processor/attributesprocessor
	Attributes map[string]interface{} `yaml:"attributes"`

	// prom service discovery
	ScrapeConfigs []interface{} `yaml:"scrape_configs"`
}

// RWConfig controls the configuration of exporting to Grafana Cloud
type RWConfig struct {
	Endpoint  string                 `yaml:"endpoint"`
	Insecure  bool                   `yaml:"insecure"`
	BasicAuth *prom_config.BasicAuth `yaml:"basic_auth,omitempty"`
	Batch     map[string]interface{} `yaml:"batch,omitempty"` // https://github.com/open-telemetry/opentelemetry-collector/blob/1405654d4e907b3215cece0ce04e46a6c1576382/processor/batchprocessor/config.go#L24
	Queue     map[string]interface{} `yaml:"queue,omitempty"` // https://github.com/open-telemetry/opentelemetry-collector/blob/1405654d4e907b3215cece0ce04e46a6c1576382/processor/queuedprocessor/config.go#L24
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

	if !c.Enabled {
		return nil, errors.New("tempo config not enabled")
	}

	if len(c.Receivers) == 0 {
		return nil, errors.New("must have at least one configured receiver")
	}

	if len(c.RemoteWrite.Endpoint) == 0 {
		return nil, errors.New("must have a configured remote_write.endpoint")
	}

	// exporter
	headers := map[string]string{}
	if c.RemoteWrite.BasicAuth != nil {
		password := string(c.RemoteWrite.BasicAuth.Password)

		if len(c.RemoteWrite.BasicAuth.PasswordFile) > 0 {
			buff, err := ioutil.ReadFile(c.RemoteWrite.BasicAuth.PasswordFile)
			if err != nil {
				return nil, fmt.Errorf("unable to load password file %s: %w", c.RemoteWrite.BasicAuth.PasswordFile, err)
			}
			password = string(buff)
		}

		encodedAuth := base64.StdEncoding.EncodeToString([]byte(c.RemoteWrite.BasicAuth.Username + ":" + password))
		headers = map[string]string{
			"authorization": "Basic " + encodedAuth,
		}
	}

	otelMapStructure["exporters"] = map[string]interface{}{
		"otlp": map[string]interface{}{
			"endpoint": c.RemoteWrite.Endpoint,
			"headers":  headers,
			"insecure": c.RemoteWrite.Insecure,
		},
	}

	// processors
	processors := map[string]interface{}{}
	processorNames := []string{}
	if c.ScrapeConfigs != nil {
		processorNames = append(processorNames, promsdprocessor.TypeStr)
		processors[promsdprocessor.TypeStr] = map[string]interface{}{
			"scrape_configs": c.ScrapeConfigs,
		}
	}

	if c.Attributes != nil {
		processors["attributes"] = c.Attributes
		processorNames = append(processorNames, "attributes")
	}

	// todo: when we update otel collector to the latest we can just use the settings on the exporter
	//       https://github.com/open-telemetry/opentelemetry-collector/tree/master/exporter/otlpexporter
	if c.RemoteWrite.Batch != nil {
		processors["batch"] = c.RemoteWrite.Batch
		processorNames = append(processorNames, "batch")
	}

	if c.RemoteWrite.Queue != nil {
		processors["queued_retry"] = c.RemoteWrite.Queue
		processorNames = append(processorNames, "queued_retry")
	}
	otelMapStructure["processors"] = processors

	// receivers
	otelMapStructure["receivers"] = c.Receivers
	receiverNames := []string{}
	for name := range c.Receivers {
		receiverNames = append(receiverNames, name)
	}

	// pipelines
	otelMapStructure["service"] = map[string]interface{}{
		"pipelines": map[string]interface{}{
			"traces": map[string]interface{}{
				"exporters":  []string{"otlp"},
				"processors": processorNames,
				"receivers":  receiverNames,
			},
		},
	}

	// now build the otel configmodel from the mapstructure
	v := viper.New()
	err := v.MergeConfigMap(otelMapStructure)
	if err != nil {
		return nil, fmt.Errorf("failed to merge in mapstructure config: %w", err)
	}

	factories, err := tracingFactories()
	if err != nil {
		return nil, fmt.Errorf("failed to create factories: %w", err)
	}

	otelCfg, err := config.Load(v, factories)
	if err != nil {
		return nil, fmt.Errorf("failed to load OTel config: %w", err)
	}

	return otelCfg, nil
}

// tracingFactories() only creates the needed factories.  if we decide to add support for a new
// processor, exporter, receiver we need to add it here
func tracingFactories() (component.Factories, error) {
	extensions, err := component.MakeExtensionFactoryMap()
	if err != nil {
		return component.Factories{}, err
	}

	receivers, err := component.MakeReceiverFactoryMap(
		jaegerreceiver.NewFactory(),
		zipkinreceiver.NewFactory(),
		otlpreceiver.NewFactory(),
		opencensusreceiver.NewFactory(),
	)
	if err != nil {
		return component.Factories{}, err
	}

	exporters, err := component.MakeExporterFactoryMap(
		otlpexporter.NewFactory(),
	)
	if err != nil {
		return component.Factories{}, err
	}

	processors, err := component.MakeProcessorFactoryMap(
		queuedprocessor.NewFactory(),
		batchprocessor.NewFactory(),
		attributesprocessor.NewFactory(),
		promsdprocessor.NewFactory(),
	)
	if err != nil {
		return component.Factories{}, err
	}

	return component.Factories{
		Extensions: extensions,
		Receivers:  receivers,
		Processors: processors,
		Exporters:  exporters,
	}, nil
}
