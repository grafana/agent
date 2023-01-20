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
	ConnectionString: os.Getenv("DATA_SOURCE_NAME"),
	MaxOpenConns:     10,
	MaxIdleConns:     0,
	QueryTimeout:     5,
	ScrapeInterval:   0,
	Common:           common.MetricsConfig{},
}

var (
	errNoConnectionString = errors.New("no connection string was provided")
)

// Config is the configuration for the oracledb v2 integration
type Config struct {
	ConnectionString  string               `yaml:"connection_string"`
	MaxIdleConns      int                  `yaml:"max_idle_connections"`
	MaxOpenConns      int                  `yaml:"max_open_connections"`
	ScrapeInterval    time.Duration        `yaml:"scrape_interval"`
	CustomMetricsPath string               `yaml:"custom_metrics_path"`
	QueryTimeout      int                  `yaml:"query_timeout"`
	Common            common.MetricsConfig `yaml:",inline"`
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
		return errors.New("no connection string was provided")
	}
	if _, err := dburl.Parse(c.ConnectionString); err != nil {
		return fmt.Errorf("unable to parse connection string: %w", err)
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
		return nil, fmt.Errorf("failed to create http handler")
	}

	return metricsutils.NewMetricsHandlerIntegration(logger, c, c.Common, globals, handler)
}

func createHandler(logger log.Logger, c *Config) (http.HandlerFunc, error) {
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

	registry := prometheus.NewRegistry()
	registry.MustRegister(oeExporter)
	h := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
	return h.ServeHTTP, nil
}
