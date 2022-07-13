// Package redis_exporter embeds https://github.com/oliver006/redis_exporter
package redis_exporter //nolint:golint

import (
	"errors"
	"fmt"
	"io/ioutil"
	"time"

	"github.com/grafana/agent/pkg/integrations"
	integrations_v2 "github.com/grafana/agent/pkg/integrations/v2"
	"github.com/grafana/agent/pkg/integrations/v2/metricsutils"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	re "github.com/oliver006/redis_exporter/exporter"
	config_util "github.com/prometheus/common/config"
)

var _ config_util.Secret = Config{}.RedisPassword

// DefaultConfig holds non-zero default options for the Config when it is
// unmarshaled from YAML.
var DefaultConfig = Config{
	Namespace:               "redis",
	ConfigCommand:           "CONFIG",
	ConnectionTimeout:       15 * time.Second,
	SetClientName:           true,
	CheckKeyGroupsBatchSize: 10000,
	MaxDistinctKeyGroups:    100,
}

// Config controls the redis_exporter integration.
type Config struct {
	IncludeExporterMetrics bool `yaml:"include_exporter_metrics"`

	// exporter-specific config.
	//
	// The exporter binary config differs to this, but these
	// are the only fields that are relevant to the exporter struct.
	RedisAddr               string             `yaml:"redis_addr,omitempty"`
	RedisUser               string             `yaml:"redis_user,omitempty"`
	RedisPassword           config_util.Secret `yaml:"redis_password,omitempty"`
	RedisPasswordFile       string             `yaml:"redis_password_file,omitempty"`
	Namespace               string             `yaml:"namespace,omitempty"`
	ConfigCommand           string             `yaml:"config_command,omitempty"`
	CheckKeys               string             `yaml:"check_keys,omitempty"`
	CheckKeyGroups          string             `yaml:"check_key_groups,omitempty"`
	CheckKeyGroupsBatchSize int64              `yaml:"check_key_groups_batch_size,omitempty"`
	MaxDistinctKeyGroups    int64              `yaml:"max_distinct_key_groups,omitempty"`
	CheckSingleKeys         string             `yaml:"check_single_keys,omitempty"`
	CheckStreams            string             `yaml:"check_streams,omitempty"`
	CheckSingleStreams      string             `yaml:"check_single_streams,omitempty"`
	CountKeys               string             `yaml:"count_keys,omitempty"`
	ScriptPath              string             `yaml:"script_path,omitempty"`
	ConnectionTimeout       time.Duration      `yaml:"connection_timeout,omitempty"`
	TLSClientKeyFile        string             `yaml:"tls_client_key_file,omitempty"`
	TLSClientCertFile       string             `yaml:"tls_client_cert_file,omitempty"`
	TLSCaCertFile           string             `yaml:"tls_ca_cert_file,omitempty"`
	SetClientName           bool               `yaml:"set_client_name,omitempty"`
	IsTile38                bool               `yaml:"is_tile38,omitempty"`
	ExportClientList        bool               `yaml:"export_client_list,omitempty"`
	ExportClientPort        bool               `yaml:"export_client_port,omitempty"`
	RedisMetricsOnly        bool               `yaml:"redis_metrics_only,omitempty"`
	PingOnConnect           bool               `yaml:"ping_on_connect,omitempty"`
	InclSystemMetrics       bool               `yaml:"incl_system_metrics,omitempty"`
	SkipTLSVerification     bool               `yaml:"skip_tls_verification,omitempty"`
}

// GetExporterOptions returns relevant Config properties as a redis_exporter
// Options struct. The redis_exporter Options struct has no yaml tags, so
// we marshal the yaml into Config and then create the re.Options from that.
func (c Config) GetExporterOptions() re.Options {
	return re.Options{
		User:                  c.RedisUser,
		Password:              string(c.RedisPassword),
		Namespace:             c.Namespace,
		ConfigCommandName:     c.ConfigCommand,
		CheckKeys:             c.CheckKeys,
		CheckKeysBatchSize:    c.CheckKeyGroupsBatchSize,
		CheckKeyGroups:        c.CheckKeyGroups,
		CheckSingleKeys:       c.CheckSingleKeys,
		CheckStreams:          c.CheckStreams,
		CheckSingleStreams:    c.CheckSingleStreams,
		CountKeys:             c.CountKeys,
		InclSystemMetrics:     c.InclSystemMetrics,
		SkipTLSVerification:   c.SkipTLSVerification,
		SetClientName:         c.SetClientName,
		IsTile38:              c.IsTile38,
		ExportClientList:      c.ExportClientList,
		ExportClientsInclPort: c.ExportClientPort,
		ConnectionTimeouts:    c.ConnectionTimeout,
		RedisMetricsOnly:      c.RedisMetricsOnly,
		PingOnConnect:         c.PingOnConnect,
	}
}

// UnmarshalYAML implements yaml.Unmarshaler for Config
func (c *Config) UnmarshalYAML(unmarshal func(interface{}) error) error {
	*c = DefaultConfig

	type plain Config
	return unmarshal((*plain)(c))
}

// Name returns the name of the integration this config is for.
func (c *Config) Name() string {
	return "redis_exporter"
}

// InstanceKey returns the addr of the redis server.
func (c *Config) InstanceKey(agentKey string) (string, error) {
	return c.RedisAddr, nil
}

// NewIntegration converts the config into an integration instance.
func (c *Config) NewIntegration(l log.Logger) (integrations.Integration, error) {
	return New(l, c)
}

func init() {
	integrations.RegisterIntegration(&Config{})
	integrations_v2.RegisterLegacy(&Config{}, integrations_v2.TypeMultiplex, metricsutils.NewNamedShim("redis"))
}

// New creates a new redis_exporter integration. The integration queries
// a redis instance's INFO and exposes the results as metrics.
func New(log log.Logger, c *Config) (integrations.Integration, error) {
	level.Debug(log).Log("msg", "initializing redis_exporter", "config", c)

	exporterConfig := c.GetExporterOptions()

	if c.RedisAddr == "" {
		return nil, errors.New("cannot create redis_exporter; redis_exporter.redis_addr is not defined")
	}

	if c.ScriptPath != "" {
		ls, err := ioutil.ReadFile(c.ScriptPath)
		if err != nil {
			return nil, fmt.Errorf("Error loading script file %s: %w", c.ScriptPath, err)
		}
		exporterConfig.LuaScript = ls
	}

	//new version of the exporter takes the file paths directly, for hot-reloading support (https://github.com/oliver006/redis_exporter/pull/526)

	if (c.TLSClientKeyFile != "") != (c.TLSClientCertFile != "") {
		return nil, errors.New("TLS client key file and cert file should both be present")
	} else if c.TLSClientKeyFile != "" && c.TLSClientCertFile != "" {
		exporterConfig.ClientKeyFile = c.TLSClientKeyFile
		exporterConfig.ClientCertFile = c.TLSClientCertFile
	}

	if c.TLSCaCertFile != "" {
		exporterConfig.CaCertFile = c.TLSCaCertFile
	}

	// optional password file to take precedence over password property
	if c.RedisPasswordFile != "" {
		password, err := ioutil.ReadFile(c.RedisPasswordFile)
		if err != nil {
			return nil, fmt.Errorf("Error loading password file %s: %w", c.RedisPasswordFile, err)
		}
		exporterConfig.Password = string(password)
	}

	exporter, err := re.NewRedisExporter(
		c.RedisAddr,
		exporterConfig,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create redis exporter: %w", err)
	}

	return integrations.NewCollectorIntegration(
		c.Name(),
		integrations.WithCollectors(exporter),
		integrations.WithExporterMetricsIncluded(c.IncludeExporterMetrics),
	), nil
}
