// Package ssl_exporter embeds https://github.com/ribbybibby/ssl_exporter/v2
package ssl_exporter

import (
	"fmt"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/grafana/agent/pkg/integrations"
	ssl_config "github.com/ribbybibby/ssl_exporter/v2/config"
)

func init() {
	integrations.RegisterIntegration(&Config{})
}

// DefaultConfig holds the default settings for the ssl_exporter integration.
var DefaultConfig = Config{
	ConfigFile: "",
	SSLTargets: []SSLTarget{},
}

// SSLTarget represents a target to scrape.
type SSLTarget struct {
	Name   string `yaml:"name"`
	Target string `yaml:"target"`
	Module string `yaml:"module"`
}

// Config controls the ssl_exporter integration.
type Config struct {
	IncludeExporterMetrics bool        `yaml:"include_exporter_metrics"`
	ConfigFile             string      `yaml:"config_file,omitempty"`
	SSLTargets             []SSLTarget `yaml:"ssl_targets"`
}

func (c Config) GetExporterOptions(log log.Logger) (*Options, error) {
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
	}, nil
}

// UnmarshalYAML implements yaml.Unmarshaler for Config.
func (c *Config) UnmarshalYAML(unmarshal func(interface{}) error) error {
	*c = DefaultConfig

	type plain Config
	return unmarshal((*plain)(c))
}

// Name returns the name of the integration.
func (c *Config) Name() string {
	return "ssl_exporter"
}

// InstanceKey returns the hostname:port of the agent.
func (c *Config) InstanceKey(agentKey string) (string, error) {
	return agentKey, nil
}

// NewIntegration converts the config into an instance of an integration.
func (c *Config) NewIntegration(l log.Logger) (integrations.Integration, error) {
	return New(l, c)
}

// New creates a new ssl_exporter integration. The integration scrapes
// metrics from ssl certificates
func New(log log.Logger, c *Config) (integrations.Integration, error) {
	var err error
	level.Debug(log).Log("msg", "initializing ssl_exporter", "config", c)

	exporterConfig, err := c.GetExporterOptions(log)
	if err != nil {
		return nil, fmt.Errorf("failed to get exporter config: %w", err)
	}

	for _, target := range c.SSLTargets {
		if target.Name == "" || target.Target == "" {
			return nil, fmt.Errorf("failed to load ssl_targets; the `name` and `target` fields are mandatory")
		}
	}

	exporter, err := NewSSLExporter(*exporterConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create ssl exporter: %w", err)
	}

	return integrations.NewCollectorIntegration(
		c.Name(),
		integrations.WithCollectors(exporter),
		integrations.WithExporterMetricsIncluded(c.IncludeExporterMetrics),
	), nil
}
