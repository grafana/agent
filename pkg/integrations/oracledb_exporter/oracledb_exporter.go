package oracledbexporter

import (
	"os"
	"time"

	"github.com/go-kit/log"
	"github.com/grafana/agent/pkg/integrations"
	oe "github.com/observiq/oracledb_exporter/collector"
)

// DefaultConfig is the default config for the oracledb v2 integration
var DefaultConfig = Config{
	DSN:            os.Getenv("DATA_SOURCE_NAME"),
	ScrapeInterval: 0,
}

// Config is the configuration for the oracledb v2 integration
type Config struct {
	DSN                    string        `yaml:"dsn,omitempty"`
	SID                    string        `yaml:"sid,omitempty"`
	MaxIdleConns           int           `yaml:"max_idle_connections,omitempty"`
	MaxOpenConns           int           `yaml:"max_open_connections,omitempty"`
	ScrapeInterval         time.Duration `yaml:"scrape_interval,omitempty"`
	DefaultFileMetricsPath string        `yaml:"default_file_metrics_path,omitempty"`
	CustomMetricsPath      string        `yaml:"custom_metrics_path,omitempty"`
	QueryTimeout           string        `yaml:"query_timeout,omitempty"`
}

// UnmarshalYAML implements yaml.Unmarshaler for Config
func (c *Config) UnmarshalYAML(unmarshal func(interface{}) error) error {
	*c = DefaultConfig

	type plain Config
	return unmarshal((*plain)(c))
}

// Name returns the integration name this config is associated with.
func (c *Config) Name() string {
	return "oracledb"
}

// InstanceKey returns the addr of the oracle instance.
func (c *Config) InstanceKey(agentKey string) (string, error) {
	return c.DSN, nil
}

// NewIntegration returns the OracleDB Exporter Integration
func (c *Config) NewIntegration(logger log.Logger) (integrations.Integration, error) {
	return New(logger, c)
}

func init() {
	integrations.RegisterIntegration(&Config{})
}

// New creates a new oracledb integration. The integrationscrapes metrics
// from an OracleDB exporter running with the https://github.com/iamseth/oracledb_exporter
func New(logger log.Logger, c *Config) (integrations.Integration, error) {
	oeExporter, err := oe.NewExporter(logger, &oe.Config{
		DSN:                c.DSN,
		DefaultFileMetrics: c.DefaultFileMetricsPath,
		MaxIdleConns:       c.MaxIdleConns,
		MaxOpenConns:       c.MaxOpenConns,
		CustomMetrics:      c.CustomMetricsPath,
		QueryTimeout:       c.QueryTimeout,
		ScrapeInterval:     c.ScrapeInterval,
	})

	if err != nil {
		return nil, err
	}
	return integrations.NewCollectorIntegration(c.Name(), integrations.WithCollectors(oeExporter)), nil
}
