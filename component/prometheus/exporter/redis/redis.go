package redis

import (
	"strings"
	"time"

	"github.com/grafana/agent/component"
	"github.com/grafana/agent/component/prometheus/exporter"
	"github.com/grafana/agent/pkg/flow/rivertypes"
	"github.com/grafana/agent/pkg/integrations"
	"github.com/grafana/agent/pkg/integrations/redis_exporter"
	config_util "github.com/prometheus/common/config"
)

func init() {
	component.Register(component.Registration{
		Name:    "prometheus.exporter.redis",
		Args:    Config{},
		Exports: exporter.Exports{},
		Build:   exporter.New(createExporter, "redis"),
	})
}

func createExporter(opts component.Options, args component.Arguments) (integrations.Integration, error) {
	cfg := args.(Config)
	return cfg.Convert().NewIntegration(opts.Logger)
}

// DefaultConfig holds non-zero default options for the Config when it is
// unmarshaled from river.
var DefaultConfig = Config{
	IncludeExporterMetrics:  false,
	Namespace:               "redis",
	ConfigCommand:           "CONFIG",
	ConnectionTimeout:       15 * time.Second,
	SetClientName:           true,
	CheckKeyGroupsBatchSize: 10000,
	MaxDistinctKeyGroups:    100,
}

type Config struct {
	IncludeExporterMetrics bool `river:"include_exporter_metrics,attr,optional"`

	// exporter-specific config.
	//
	// The exporter binary config differs to this, but these
	// are the only fields that are relevant to the exporter struct.
	RedisAddr               string            `river:"redis_addr,attr,optional"`
	RedisUser               string            `river:"redis_user,attr,optional"`
	RedisPassword           rivertypes.Secret `river:"redis_password,attr,optional"`
	RedisPasswordFile       string            `river:"redis_password_file,attr,optional"`
	RedisPasswordMapFile    string            `river:"redis_password_map_file,attr,optional"`
	Namespace               string            `river:"namespace,attr,optional"`
	ConfigCommand           string            `river:"config_command,attr,optional"`
	CheckKeys               []string          `river:"check_keys,attr,optional"`
	CheckKeyGroups          []string          `river:"check_key_groups,attr,optional"`
	CheckKeyGroupsBatchSize int64             `river:"check_key_groups_batch_size,attr,optional"`
	MaxDistinctKeyGroups    int64             `river:"max_distinct_key_groups,attr,optional"`
	CheckSingleKeys         []string          `river:"check_single_keys,attr,optional"`
	CheckStreams            []string          `river:"check_streams,attr,optional"`
	CheckSingleStreams      []string          `river:"check_single_streams,attr,optional"`
	CountKeys               []string          `river:"count_keys,attr,optional"`
	ScriptPaths             []string          `river:"script_paths,attr,optional"`
	ConnectionTimeout       time.Duration     `river:"connection_timeout,attr,optional"`
	TLSClientKeyFile        string            `river:"tls_client_key_file,attr,optional"`
	TLSClientCertFile       string            `river:"tls_client_cert_file,attr,optional"`
	TLSCaCertFile           string            `river:"tls_ca_cert_file,attr,optional"`
	SetClientName           bool              `river:"set_client_name,attr,optional"`
	IsTile38                bool              `river:"is_tile38,attr,optional"`
	IsCluster               bool              `river:"is_cluster,attr,optional"`
	ExportClientList        bool              `river:"export_client_list,attr,optional"`
	ExportClientPort        bool              `river:"export_client_port,attr,optional"`
	RedisMetricsOnly        bool              `river:"redis_metrics_only,attr,optional"`
	PingOnConnect           bool              `river:"ping_on_connect,attr,optional"`
	InclSystemMetrics       bool              `river:"incl_system_metrics,attr,optional"`
	InclConfigMetrics       bool              `river:"incl_config_metrics,attr,optional"`
	RedactConfigMetrics     bool              `river:"redact_config_metrics,attr,optional"`
	SkipTLSVerification     bool              `river:"skip_tls_verification,attr,optional"`
}

// UnmarshalRiver implements River unmarshalling for Config.
func (c *Config) UnmarshalRiver(f func(interface{}) error) error {
	*c = DefaultConfig

	type cfg Config
	return f((*cfg)(c))
}

func (c *Config) Convert() *redis_exporter.Config {
	return &redis_exporter.Config{
		IncludeExporterMetrics:  c.IncludeExporterMetrics,
		RedisAddr:               c.RedisAddr,
		RedisUser:               c.RedisUser,
		RedisPassword:           config_util.Secret(c.RedisPassword),
		RedisPasswordFile:       c.RedisPasswordFile,
		RedisPasswordMapFile:    c.RedisPasswordMapFile,
		Namespace:               c.Namespace,
		ConfigCommand:           c.ConfigCommand,
		CheckKeys:               strings.Join(c.CheckKeys, ","),
		CheckKeyGroups:          strings.Join(c.CheckKeyGroups, ","),
		CheckKeyGroupsBatchSize: c.CheckKeyGroupsBatchSize,
		MaxDistinctKeyGroups:    c.MaxDistinctKeyGroups,
		CheckSingleKeys:         strings.Join(c.CheckSingleKeys, ","),
		CheckStreams:            strings.Join(c.CheckStreams, ","),
		CheckSingleStreams:      strings.Join(c.CheckSingleStreams, ","),
		CountKeys:               strings.Join(c.CountKeys, ","),
		ScriptPath:              strings.Join(c.ScriptPaths, ","),
		ConnectionTimeout:       c.ConnectionTimeout,
		TLSClientKeyFile:        c.TLSClientKeyFile,
		TLSClientCertFile:       c.TLSClientCertFile,
		TLSCaCertFile:           c.TLSCaCertFile,
		SetClientName:           c.SetClientName,
		IsTile38:                c.IsTile38,
		IsCluster:               c.IsCluster,
		ExportClientList:        c.ExportClientList,
		ExportClientPort:        c.ExportClientPort,
		RedisMetricsOnly:        c.RedisMetricsOnly,
		PingOnConnect:           c.PingOnConnect,
		InclSystemMetrics:       c.InclSystemMetrics,
		InclConfigMetrics:       c.InclConfigMetrics,
		RedactConfigMetrics:     c.RedactConfigMetrics,
		SkipTLSVerification:     c.SkipTLSVerification,
	}
}
