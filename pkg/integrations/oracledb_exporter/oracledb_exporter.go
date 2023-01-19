package oracledbexporter

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/go-kit/log"
	"github.com/grafana/agent/pkg/integrations"
	oe "github.com/observiq/oracledb_exporter/collector"
	_ "github.com/sijms/go-ora/v2"
	"github.com/xo/dburl"
)

// DefaultConfig is the default config for the oracledb v2 integration
var DefaultConfig = Config{
	ConnectionString: os.Getenv("DATA_SOURCE_NAME"),
	MaxOpenConns:     10,
	MaxIdleConns:     0,
	QueryTimeout:     "5",
	ScrapeInterval:   0,
}

// Config is the configuration for the oracledb v2 integration
type Config struct {
	ConnectionString  string        `yaml:"connection_string,omitempty"`
	MaxIdleConns      int           `yaml:"max_idle_connections,omitempty"`
	MaxOpenConns      int           `yaml:"max_open_connections,omitempty"`
	ScrapeInterval    time.Duration `yaml:"scrape_interval,omitempty"`
	CustomMetricsPath string        `yaml:"custom_metrics_path,omitempty"`
	QueryTimeout      string        `yaml:"query_timeout,omitempty"`
}

// Validate returnsif the configuration is valid
func (c *Config) Validate() error {
	if c.ConnectionString == "" {
		return errors.New("no connection string was provided")
	}
	return nil
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
	u, err := dburl.Parse(c.ConnectionString)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s:%s", u.Hostname(), u.Port()), nil
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
	if err := c.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	oeExporter, err := oe.NewExporter(logger, &oe.Config{
		DSN:            c.ConnectionString,
		MaxIdleConns:   c.MaxIdleConns,
		MaxOpenConns:   c.MaxOpenConns,
		CustomMetrics:  c.CustomMetricsPath,
		QueryTimeout:   c.QueryTimeout,
		ScrapeInterval: c.ScrapeInterval,
	})

	if err != nil {
		return nil, err
	}
	return integrations.NewCollectorIntegration(c.Name(), integrations.WithCollectors(oeExporter)), nil
}
