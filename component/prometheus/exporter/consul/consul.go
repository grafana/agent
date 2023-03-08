package consul

import (
	"time"

	"github.com/grafana/agent/component"
	"github.com/grafana/agent/component/prometheus/exporter"
	"github.com/grafana/agent/pkg/integrations"
	"github.com/grafana/agent/pkg/integrations/consul_exporter"
)

func init() {
	component.Register(component.Registration{
		Name:    "prometheus.exporter.consul",
		Args:    Config{},
		Exports: exporter.Exports{},
		Build:   exporter.New(createExporter, "consul"),
	})
}

func createExporter(opts component.Options, args component.Arguments) (integrations.Integration, error) {
	cfg := args.(Config)
	return cfg.Convert().NewIntegration(opts.Logger)
}

// DefaultConfig holds the default settings for the consul_exporter exporter.
var DefaultConfig = Config{
	Server:        "http://localhost:8500",
	Timeout:       500 * time.Millisecond,
	AllowStale:    true,
	KVFilter:      ".*",
	HealthSummary: true,
}

// Config controls the consul_exporter exporter.
type Config struct {
	Server             string        `river:"server,attr,optional"`
	CAFile             string        `river:"ca_file,attr,optional"`
	CertFile           string        `river:"cert_file,attr,optional"`
	KeyFile            string        `river:"key_file,attr,optional"`
	ServerName         string        `river:"server_name,attr,optional"`
	Timeout            time.Duration `river:"timeout,attr,optional"`
	InsecureSkipVerify bool          `river:"insecure_skip_verify,attr,optional"`
	RequestLimit       int           `river:"concurrent_request_limit,attr,optional"`
	AllowStale         bool          `river:"allow_stale,attr,optional"`
	RequireConsistent  bool          `river:"require_consistent,attr,optional"`

	KVPrefix      string `river:"kv_prefix,attr,optional"`
	KVFilter      string `river:"kv_filter,attr,optional"`
	HealthSummary bool   `river:"generate_health_summary,attr,optional"`
}

// UnmarshalRiver implements River unmarshalling for Config.
func (c *Config) UnmarshalRiver(f func(interface{}) error) error {
	*c = DefaultConfig

	type cfg Config
	return f((*cfg)(c))
}

func (c *Config) Convert() *consul_exporter.Config {
	return &consul_exporter.Config{
		Server:             c.Server,
		CAFile:             c.CAFile,
		CertFile:           c.CertFile,
		KeyFile:            c.KeyFile,
		ServerName:         c.ServerName,
		Timeout:            c.Timeout,
		InsecureSkipVerify: c.InsecureSkipVerify,
		RequestLimit:       c.RequestLimit,
		AllowStale:         c.AllowStale,
		RequireConsistent:  c.RequireConsistent,
		KVPrefix:           c.KVPrefix,
		KVFilter:           c.KVFilter,
		HealthSummary:      c.HealthSummary,
	}
}
