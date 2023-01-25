package oracledbexporter

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/go-kit/log"
	"github.com/grafana/agent/pkg/integrations/v2/common"
	"github.com/grafana/agent/pkg/integrations/v2/metricsutils"
	oe "github.com/iamseth/oracledb_exporter/collector"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	// required driver for integration
	_ "github.com/sijms/go-ora/v2"

	oe_v1 "github.com/grafana/agent/pkg/integrations/oracledb_exporter"
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

	u, err := url.Parse(c.ConnectionString)
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
	if err := oe_v1.ValidateConnString(c.ConnectionString); err != nil {
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
