// Package consul_exporter embeds https://github.com/prometheus/consul_exporter
package consul_exporter //nolint:golint

import (
	"time"

	"github.com/go-kit/kit/log"
	"github.com/grafana/agent/pkg/integrations/common"
	"github.com/grafana/agent/pkg/integrations/config"
	consul_api "github.com/hashicorp/consul/api"
	"github.com/prometheus/consul_exporter/pkg/exporter"
)

var DefaultConfig = Config{
	Server:        "http://localhost:8500",
	Timeout:       500 * time.Millisecond,
	AllowStale:    true,
	KVFilter:      ".*",
	HealthSummary: true,
}

// Config controls the consul_exporter integration.
type Config struct {
	// Enabled enables the integration.
	Enabled bool `yaml:"enabled"`

	CommonConfig config.Common `yaml:",inline"`

	Server             string        `yaml:"server"`
	CAFile             string        `yaml:"ca_file"`
	CertFile           string        `yaml:"cert_file"`
	KeyFile            string        `yaml:"key_file"`
	ServerName         string        `yaml:"server_name"`
	Timeout            time.Duration `yaml:"timeout"`
	InsecureSkipVerify bool          `yaml:"insecure_skip_verify"`
	RequestLimit       int           `yaml:"concurrent_request_limit"`
	AllowStale         bool          `yaml:"allow_stale"`
	RequireConsistent  bool          `yaml:"require_consistent"`

	KVPrefix      string `yaml:"kv_prefix"`
	KVFilter      string `yaml:"kv_filter"`
	HealthSummary bool   `yaml:"generate_health_summary"`
}

// UnmarshalYAML implements yaml.Unmarshaler for Config.
func (c *Config) UnmarshalYAML(unmarshal func(interface{}) error) error {
	*c = DefaultConfig

	type plain Config
	return unmarshal((*plain)(c))
}

// New creates a new consul_exporter integration. The integration scrapes
// metrics from a consul process.
func New(log log.Logger, c Config) (common.Integration, error) {
	var (
		consulOpts = exporter.ConsulOpts{
			CAFile:       c.CAFile,
			CertFile:     c.CertFile,
			Insecure:     c.InsecureSkipVerify,
			KeyFile:      c.KeyFile,
			RequestLimit: c.RequestLimit,
			ServerName:   c.ServerName,
			Timeout:      c.Timeout,
			URI:          c.Server,
		}
		queryOptions = consul_api.QueryOptions{
			AllowStale:        c.AllowStale,
			RequireConsistent: c.RequireConsistent,
		}
	)

	e, err := exporter.New(consulOpts, queryOptions, c.KVPrefix, c.KVFilter, c.HealthSummary, log)
	if err != nil {
		return nil, err
	}

	return common.NewCollectorIntegration(
		"consul_exporter",
		c.CommonConfig,
		e,
		false,
	), nil
}
