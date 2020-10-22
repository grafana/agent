package redis_exporter //nolint:golint

import (
	"time"

	re "github.com/oliver006/redis_exporter/lib/exporter"

	"github.com/grafana/agent/pkg/integrations/config"
)

var (
	// DefaultConfig holds non-zero default options for the Config when it is
	// unmarshaled from YAML.
	DefaultConfig = Config{
		Namespace:         "redis",
		ConfigCommand:     "CONFIG",
		ConnectionTimeout: (15 * time.Second),
		SetClientName:     true,
	}
)

// Config controls the redis_exporter integration. The exporter accepts more
// config properties than this, but these are the only fields with non-default
// values that we need to define right now.
type Config struct {
	Enabled      bool          `yaml:"enabled"`
	CommonConfig config.Common `yaml:",inline"`

	IncludeExporterMetrics bool `yaml:"include_exporter_metrics"`

	// exporter-specific config
	RedisAddr           string        `yaml:"redis_addr"`
	RedisUser           string        `yaml:"redis_user"`
	RedisPassword       string        `yaml:"redis_password"`
	RedisPasswordFile   string        `yaml:"redis_password_file"`
	Namespace           string        `yaml:"namespace"`
	ConfigCommand       string        `yaml:"config_command"`
	CheckKeys           string        `yaml:"check_keys"`
	CheckSingleKeys     string        `yaml:"check_single_keys"`
	CheckStreams        string        `yaml:"check_streams"`
	CheckSingleStreams  string        `yaml:"check_single_streams"`
	CountKeys           string        `yaml:"count_keys"`
	ScriptPath          string        `yaml:"script_path"`
	ConnectionTimeout   time.Duration `yaml:"connection_timeout"`
	TLSClientKeyFile    string        `yaml:"tls_client_key_file"`
	TLSClientCertFile   string        `yaml:"tls_client_cert_file"`
	TLSCaCertFile       string        `yaml:"tls_ca_cert_file"`
	SetClientName       bool          `yaml:"set_client_name"`
	IsTile38            bool          `yaml:"is_tile38"`
	ExportClientList    bool          `yaml:"export_client_list"`
	RedisMetricsOnly    bool          `yaml:"redis_metrics_only"`
	PingOnConnect       bool          `yaml:"ping_on_connect"`
	InclSystemMetrics   bool          `yaml:"incl_system_metrics"`
	SkipTLSVerification bool          `yaml:"skip_tls_verification"`
}

// GetExporterOptions returns relevant Config properties as a redis_exporter
// Options struct. The redis_exporter Options struct has no yaml tags, so
// we marshal the yaml into Config and then create the re.Options from that.
func (c Config) GetExporterOptions() re.Options {

	return re.Options{
		User:                c.RedisUser,
		Password:            c.RedisPassword,
		Namespace:           c.Namespace,
		ConfigCommandName:   c.ConfigCommand,
		CheckKeys:           c.CheckKeys,
		CheckSingleKeys:     c.CheckSingleKeys,
		CheckStreams:        c.CheckStreams,
		CheckSingleStreams:  c.CheckSingleStreams,
		CountKeys:           c.CountKeys,
		InclSystemMetrics:   c.InclSystemMetrics,
		SkipTLSVerification: c.SkipTLSVerification,
		SetClientName:       c.SetClientName,
		IsTile38:            c.IsTile38,
		ExportClientList:    c.ExportClientList,
		ConnectionTimeouts:  c.ConnectionTimeout,
		RedisMetricsOnly:    c.RedisMetricsOnly,
		PingOnConnect:       c.PingOnConnect,
	}
}

// UnmarshalYAML implements yaml.Unmarshaler for Config
func (c *Config) UnmarshalYAML(unmarshal func(interface{}) error) error {
	*c = DefaultConfig

	type plain Config
	return unmarshal((*plain)(c))
}
