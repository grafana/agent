package ssl_exporter

import (
	"fmt"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/grafana/agent/pkg/integrations/v2/common"
	ssl_config "github.com/ribbybibby/ssl_exporter/v2/config"

	integrations_v2 "github.com/grafana/agent/pkg/integrations/v2"
)

// Config controls the ssl_exporter integration.
type Config struct {
	IncludeExporterMetrics bool                 `yaml:"include_exporter_metrics"`
	ConfigFile             string               `yaml:"config_file,omitempty"`
	SSLTargets             []SSLTarget          `yaml:"ssl_targets"`
	Common                 common.MetricsConfig `yaml:",inline"`

	globals integrations_v2.Globals
}

func (c *Config) Name() string {
	return "ssl_exporter"
}

func (c *Config) ApplyDefaults(globals integrations_v2.Globals) error {
	c.globals = globals
	c.Common.ApplyDefaults(globals.SubsystemOpts.Metrics.Autoscrape)
	return nil
}

func (c *Config) Identifier(globals integrations_v2.Globals) (string, error) {
	if c.Common.InstanceKey != nil {
		return *c.Common.InstanceKey, nil
	}
	// This is a bit odd and will error if you specify two ssl readers but with non-unique names
	return "ssl", nil
}

func (c *Config) GetExporterOptions(log log.Logger) (*Options, error) {
	var err error
	conf := ssl_config.DefaultConfig

	if c.ConfigFile != "" {
		conf, err = ssl_config.LoadConfig(c.ConfigFile)
		if err != nil {
			return nil, fmt.Errorf("failed to load ssl config from file %v: %w", c.ConfigFile, err)
		}
	}

	return &Options{
		Namespace:  c.Name(),
		SSLTargets: c.SSLTargets,
		SSLConfig:  conf,
		Logger:     log,
		Name:       c.Name(),
	}, nil
}

func (c *Config) NewIntegration(logger log.Logger, globals integrations_v2.Globals) (integrations_v2.Integration, error) {
	var err error
	level.Debug(logger).Log("msg", "initializing ssl_exporter", "config", c)

	exporterConfig, err := c.GetExporterOptions(logger)
	if err != nil {
		return nil, fmt.Errorf("failed to get exporter config: %w", err)
	}

	for _, target := range c.SSLTargets {
		if target.Name == "" || target.Target == "" {
			return nil, fmt.Errorf("failed to load ssl_targets; the `name` and `target` fields are mandatory")
		}
	}

	exporter, err := NewSSLExporter(*exporterConfig, c)
	if err != nil {
		return nil, fmt.Errorf("failed to create ssl exporter: %w", err)
	}
	return exporter, nil
}

// UnmarshalYAML implements yaml.Unmarshaler for Config.
func (c *Config) UnmarshalYAML(unmarshal func(interface{}) error) error {
	*c = DefaultConfig

	type plain Config
	return unmarshal((*plain)(c))
}

// SSLTarget represents a target to scrape.
type SSLTarget struct {
	Name   string `yaml:"name"`
	Target string `yaml:"target"`
	Module string `yaml:"module"`
}

// DefaultConfig holds the default settings for the ssl_exporter integration.
var DefaultConfig = Config{
	ConfigFile: "",
	SSLTargets: []SSLTarget{},
}
