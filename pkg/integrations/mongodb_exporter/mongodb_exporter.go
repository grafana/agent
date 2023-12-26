package mongodb_exporter //nolint:golint

import (
	"errors"
	"fmt"
	"net/url"
	"os"

	"github.com/go-kit/log"
	"github.com/percona/mongodb_exporter/exporter"
	config_util "github.com/prometheus/common/config"

	"github.com/grafana/agent/pkg/integrations"
	integrations_v2 "github.com/grafana/agent/pkg/integrations/v2"
	"github.com/grafana/agent/pkg/integrations/v2/metricsutils"
)

var DefaultConfig = Config{
	DirectConnect: true,
	EnableDBStats: false,
	EnableDiagnosticData: false,
	EnableReplicasetStatus: false,
	EnableTopMetrics: false,
	EnableIndexStats: false,
	EnableCollStats: false,
}

var collectAll = true

// Config controls mongodb_exporter
type Config struct {
	// MongoDB connection URI. example:mongodb://user:pass@127.0.0.1:27017/admin?ssl=true"
	URI                      config_util.Secret `yaml:"mongodb_uri"`
	DirectConnect            bool               `yaml:"direct_connect,omitempty"`
	DiscoveringMode          bool               `yaml:"discovering_mode,omitempty"`
	EnableDBStats            bool               `yaml:"enable_db_stats,omitempty"`
	EnableDiagnosticData     bool               `yaml:"enable_diagnostic_data,omitempty"`
	EnableReplicasetStatus   bool               `yaml:"enable_replicaset_status,omitempty"`
	EnableTopMetrics         bool               `yaml:"enable_top_metrics,omitempty"`
	EnableIndexStats         bool               `yaml:"enable_index_stats,omitempty"`
	EnableCollStats          bool               `yaml:"enable_coll_stats,omitempty"`
	TLSBasicAuthConfigPath   string             `yaml:"tls_basic_auth_config_path,omitempty"`
}

// UnmarshalYAML implements yaml.Unmarshaler for Config
func (c *Config) UnmarshalYAML(unmarshal func(interface{}) error) error {
	*c = DefaultConfig
	type plain Config
	return unmarshal((*plain)(c))
}

// Name returns the name of the integration that this config represents.
func (c *Config) Name() string {
	return "mongodb_exporter"
}

// InstanceKey returns the address:port of the mongodb server being queried.
func (c *Config) InstanceKey(_ string) (string, error) {
	u, err := url.Parse(string(c.URI))
	if err != nil {
		return "", fmt.Errorf("could not parse mongodb_uri: %w", errors.Unwrap(err))
	}
	return u.Host, nil
}

// NewIntegration creates a new mongodb_exporter
func (c *Config) NewIntegration(logger log.Logger) (integrations.Integration, error) {
	return New(logger, c)
}

func init() {
	integrations.RegisterIntegration(&Config{})
	integrations_v2.RegisterLegacy(&Config{}, integrations_v2.TypeMultiplex, metricsutils.NewNamedShim("mongodb"))
}

// New creates a new mongodb_exporter integration.
func New(logger log.Logger, c *Config) (integrations.Integration, error) {
	logrusLogger := integrations.NewLogger(logger)

	if c.TLSBasicAuthConfigPath != "" {
		if _, err := os.Stat(c.TLSBasicAuthConfigPath); err != nil {
			return nil, fmt.Errorf("tls config file path is invalid: %s. error: %w", c.TLSBasicAuthConfigPath, errors.Unwrap(err))
		}
	}

	if c.EnableDBStats || c.EnableDiagnosticData || 
	   c.EnableReplicasetStatus || c.EnableCollStats || 
	   c.EnableTopMetrics || c.EnableIndexStats  {
		  collectAll = false	
	} 

	if collectAll {
		c.EnableDBStats = true
		c.EnableDiagnosticData = true
		c.EnableReplicasetStatus = true
		c.EnableTopMetrics = true
		c.EnableIndexStats = true
		c.EnableCollStats = true
	}

	if c.EnableIndexStats || c.EnableCollStats {
		c.DiscoveringMode = true
	}

	exp := exporter.New(&exporter.Opts{
		URI:                    string(c.URI),
		Logger:                 logrusLogger,
		DisableDefaultRegistry: true,

		// NOTE(rfratto): CompatibleMode configures the exporter to use old metric
		// names from mongodb_exporter <v0.20.0. Many existing dashboards rely on
		// the old names, so we hard-code it to true now. We may wish to make this
		// configurable in the future.
		CompatibleMode:           true,
		DirectConnect:            c.DirectConnect,
		DiscoveringMode:          c.DiscoveringMode,

		CollectAll:               collectAll,
		EnableDBStats:            c.EnableDBStats,
		EnableDiagnosticData:     c.EnableDiagnosticData,
		EnableReplicasetStatus:   c.EnableReplicasetStatus,
		EnableTopMetrics:         c.EnableTopMetrics,
		EnableIndexStats:         c.EnableIndexStats,
		EnableCollStats:          c.EnableCollStats,
		
		TLSConfigPath:   c.TLSBasicAuthConfigPath,
	})

	return integrations.NewHandlerIntegration(c.Name(), exp.Handler()), nil
}
