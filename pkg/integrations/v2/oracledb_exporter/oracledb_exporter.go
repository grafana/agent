package oracledbexporter

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/go-kit/log"
	"github.com/grafana/agent/pkg/integrations/v2/common"
	"github.com/grafana/agent/pkg/integrations/v2/metricsutils"
	oe "github.com/observiq/oracledb_exporter/collector"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	// required driver for integration
	_ "github.com/sijms/go-ora/v2"
	"github.com/xo/dburl"

	integrations_v2 "github.com/grafana/agent/pkg/integrations/v2"
)

// DefaultConfig is the default config for the oracledb v2 integration
var DefaultConfig = Config{
	ConnectionString:      os.Getenv("DATA_SOURCE_NAME"),
	MaxOpenConns:          10,
	MaxIdleConns:          0,
	QueryTimeout:          5,
	MetricsScrapeInterval: 0,
	Common:                common.MetricsConfig{},
}

var (
	errNoConnectionString = errors.New("no connection string was provided")
	errNoHostname         = errors.New("no hostname in connection string")
)

// Config is the configuration for the oracledb v2 integration
type Config struct {
	ConnectionString      string               `yaml:"connection_string"`
	MaxIdleConns          int                  `yaml:"max_idle_connections"`
	MaxOpenConns          int                  `yaml:"max_open_connections"`
	MetricsScrapeInterval time.Duration        `yaml:"metrics_scrape_interval"`
	QueryTimeout          int                  `yaml:"query_timeout"`
	Common                common.MetricsConfig `yaml:",inline"`
}

// UnmarshalYAML implements yaml.Unmarshaler for Config
func (c *Config) UnmarshalYAML(unmarshal func(interface{}) error) error {
	*c = DefaultConfig
	type plain Config
	return unmarshal((*plain)(c))
}

// Validate returns errors if the configuration is invalid and the exporter could not function with it
func (c *Config) Validate() error {
	if c.ConnectionString == "" {
		return errNoConnectionString
	}
	u, err := dburl.Parse(c.ConnectionString)
	if err != nil {
		return fmt.Errorf("unable to parse connection string: %w", err)
	}

	if u.Scheme != "oracle" {
		return fmt.Errorf("unexpected scheme of type '%s'. Was expecting 'oracle': %w", u.Scheme, err)
	}

	// hostname is required for identification
	if u.Hostname() == "" {
		return errNoHostname
	}
	return nil
}

// Name returns the integration name this config is associated with.
func (c *Config) Name() string {
	return "oracledb"
}

// ApplyDefaults applies the integrations
func (c *Config) ApplyDefaults(globals integrations_v2.Globals) error {
	c.Common.ApplyDefaults(globals.SubsystemOpts.Metrics.Autoscrape)
	return nil
}

// Identifier returns a string that identifies the integration.
func (c *Config) Identifier(globals integrations_v2.Globals) (string, error) {
	if c.Common.InstanceKey != nil {
		return *c.Common.InstanceKey, nil
	}

	u, err := dburl.Parse(c.ConnectionString)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s:%s", u.Hostname(), u.Port()), nil
}

func init() {
	integrations_v2.Register(&Config{}, integrations_v2.TypeMultiplex)
}

// NewIntegration returns the OracleDB Exporter Integration
func (c *Config) NewIntegration(logger log.Logger, globals integrations_v2.Globals) (integrations_v2.Integration, error) {
	handler, err := createHandler(logger, c)
	if err != nil {
		return nil, fmt.Errorf("failed to create http handler, %w", err)
	}

	return metricsutils.NewMetricsHandlerIntegration(logger, c, c.Common, globals, handler)
}

func createHandler(logger log.Logger, c *Config) (http.HandlerFunc, error) {
	if err := c.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	oeExporter, err := oe.NewExporter(logger, &oe.Config{
		DSN:          c.ConnectionString,
		MaxIdleConns: c.MaxIdleConns,
		MaxOpenConns: c.MaxOpenConns,
		// no custom metrics for this integration
		CustomMetrics:  "",
		QueryTimeout:   c.QueryTimeout,
		ScrapeInterval: c.MetricsScrapeInterval,
	})
	if err != nil {
		return nil, err
	}

	registry := prometheus.NewRegistry()
	registry.MustRegister(oeExporter)
	h := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
	return h.ServeHTTP, nil
}
