package oracledbexporter

import (
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

	integrations_v2 "github.com/grafana/agent/pkg/integrations/v2"
)

// DefaultConfig is the default config for the oracledb v2 integration
var DefaultConfig = Config{
	DSN:            os.Getenv("DATA_SOURCE_NAME"),
	ScrapeInterval: 0,
}

// Config is the configuration for the oracledb v2 integration
type Config struct {
	DSN                    string               `yaml:"dsn,omitempty"`
	SID                    string               `yaml:"sid,omitempty"`
	MaxIdleConns           int                  `yaml:"max_idle_connections,omitempty"`
	MaxOpenConns           int                  `yaml:"max_open_connections,omitempty"`
	ScrapeInterval         time.Duration        `yaml:"scrape_interval,omitempty"`
	DefaultFileMetricsPath string               `yaml:"default_file_metrics_path,omitempty"`
	CustomMetricsPath      string               `yaml:"custom_metrics_path,omitempty"`
	QueryTimeout           string               `yaml:"query_timeout,omitempty"`
	Common                 common.MetricsConfig `yaml:",inline"`
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
	return c.DSN, nil
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

	registry := prometheus.NewRegistry()
	registry.MustRegister(oeExporter)
	h := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
	return h.ServeHTTP, nil
}
