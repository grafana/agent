package config

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"testing"
	"unicode"

	"github.com/weaveworks/common/logging"

	"github.com/grafana/agent/pkg/config/interfaces"

	"github.com/prometheus/exporter-toolkit/web"

	promCfg "github.com/prometheus/prometheus/config"

	"github.com/weaveworks/common/server"

	"github.com/drone/envsubst/v2"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/grafana/agent/pkg/config/features"
	"github.com/grafana/agent/pkg/logs"
	"github.com/grafana/agent/pkg/metrics"
	"github.com/grafana/agent/pkg/traces"
	"github.com/prometheus/common/config"
	"github.com/prometheus/common/version"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

var (
	featRemoteConfigs    = features.Feature("remote-configs")
	featIntegrationsNext = features.Feature("integrations-next")
	featDynamicConfig    = features.Feature("dynamic-config")

	allFeatures = []features.Feature{
		featRemoteConfigs,
		featIntegrationsNext,
		featDynamicConfig,
	}
)

// DefaultConfig holds default settings for all the subsystems.
var DefaultConfig = Config{
	// All subsystems with a DefaultConfig should be listed here.
	Metrics:               metrics.DefaultConfig,
	Integrations:          DefaultVersionedIntegrations,
	EnableConfigEndpoints: false,
}

// Config contains underlying configurations for the agent
type Config struct {
	Server       server.Config         `yaml:"server,omitempty"`
	Metrics      metrics.Config        `yaml:"metrics,omitempty"`
	Integrations VersionedIntegrations `yaml:"integrations,omitempty"`
	Traces       traces.Config         `yaml:"traces,omitempty"`
	Logs         *logs.Config          `yaml:"logs,omitempty"`

	// We support a secondary server just for the /-/reload endpoint, since
	// invoking /-/reload against the primary server can cause the server
	// to restart.
	ReloadAddressVal string `yaml:"-"`
	ReloadPortVal    int    `yaml:"-"`

	// Deprecated fields user has used. Generated during UnmarshalYAML.
	Deprecations []string `yaml:"-"`

	// Remote config options
	BasicAuthUser     string `yaml:"-"`
	BasicAuthPassFile string `yaml:"-"`

	// Toggle for config endpoint(s)
	EnableConfigEndpoints bool `yaml:"-" ,default:"false"`

	node *yaml.Node
}

func (c *Config) ReloadPort() int {
	return c.ReloadPortVal
}

func (c *Config) ReloadAddress() string {
	return c.ReloadAddressVal
}

func (c *Config) ServerConfig() interfaces.ServerConfig {
	sw := &ServerWrapper{}
	sw.Config = &c.Server
	return sw
}

func (c *Config) MetricsConfig() interfaces.MetricsConfig {
	return nil
}

func (c *Config) LogsConfig() interfaces.LogsConfig {
	//TODO implement me
	panic("implement me")
}

func (c *Config) WALDir() string {
	return c.Metrics.WALDir
}

func (c *Config) GlobalRemoteWrite() []*promCfg.RemoteWriteConfig {
	return c.Metrics.Global.RemoteWrite
}

func (c *Config) GlobalConfig() *promCfg.GlobalConfig {
	return &c.Metrics.Global.Prometheus
}

func (c *Config) HTTPListenPort() int {
	return c.Server.HTTPListenPort
}

func (c *Config) HTTPListenAddress() string {
	return c.Server.HTTPListenAddress
}

func (c *Config) HTTPTLSConfig() web.TLSStruct {
	return c.Server.HTTPTLSConfig
}

type ServerWrapper struct {
	*server.Config
}

func (s *ServerWrapper) LogLevel() logging.Level {
	return s.Config.LogLevel
}

func (s *ServerWrapper) LogFormat() logging.Format {
	return s.Config.LogFormat
}

func (s *ServerWrapper) Log() logging.Interface {
	return s.Config.Log
}

func (s *ServerWrapper) HTTPListenPort() int {
	return s.Config.HTTPListenPort
}

func (s *ServerWrapper) HTTPListenAddress() string {
	return s.Config.HTTPListenAddress
}

func (s *ServerWrapper) HTTPTLSConfig() web.TLSStruct {
	return s.Config.HTTPTLSConfig
}

// LogDeprecations will log use of any deprecated fields to l as warn-level
// messages.
func (c *Config) LogDeprecations(l log.Logger) {
	for _, d := range c.Deprecations {
		level.Warn(l).Log("msg", fmt.Sprintf("DEPRECATION NOTICE: %s", d))
	}
}

// Validate validates the config, flags, and sets default values.
func (c *Config) Validate(fs *flag.FlagSet) error {
	if err := c.Metrics.ApplyDefaults(); err != nil {
		return err
	}

	c.Metrics.ServiceConfig.Lifecycler.ListenPort = c.Server.GRPCListenPort

	if err := c.Integrations.ApplyDefaults(c.ServerConfig(), c.MetricsConfig()); err != nil {
		return err
	}

	// since the Traces config might rely on an existing Loki config
	// this check is made here to look for cross config issues before we attempt to load
	if err := c.Traces.Validate(c.Logs); err != nil {
		return err
	}

	c.Metrics.ServiceConfig.APIEnableGetConfiguration = c.EnableConfigEndpoints

	// Don't validate flags if there's no FlagSet. Used for testing.
	if fs == nil {
		return nil
	}
	deps := []features.Dependency{
		{Flag: "config.url.basic-auth-user", Feature: featRemoteConfigs},
		{Flag: "config.url.basic-auth-password-file", Feature: featRemoteConfigs},
	}
	return features.Validate(fs, deps)
}

// RegisterFlags registers flags in underlying configs
func (c *Config) RegisterFlags(f *flag.FlagSet) {
	c.Server.RegisterInstrumentation = true
	c.Metrics.RegisterFlags(f)
	c.Server.RegisterFlags(f)

	f.StringVar(&c.ReloadAddressVal, "reload-addr", "127.0.0.1", "address to expose a secondary server for /-/reload on.")
	f.IntVar(&c.ReloadPortVal, "reload-port", 0, "port to expose a secondary server for /-/reload on. 0 disables secondary server.")

	f.StringVar(&c.BasicAuthUser, "config.url.basic-auth-user", "",
		"basic auth username for fetching remote config. (requires remote-configs experiment to be enabled")
	f.StringVar(&c.BasicAuthPassFile, "config.url.basic-auth-password-file", "",
		"path to file containing basic auth password for fetching remote config. (requires remote-configs experiment to be enabled")

	f.BoolVar(&c.EnableConfigEndpoints, "config.enable-read-api", false, "Enables the /-/config and /agent/api/v1/configs/{name} APIs. Be aware that secrets could be exposed by enabling these endpoints!")
}

// LoadFile reads a file and passes the contents to Load
func LoadFile(filename string, expandEnvVars bool, c *Config) error {
	buf, err := ioutil.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("error reading config file %w", err)
	}
	return LoadBytes(buf, expandEnvVars, c)
}

// LoadRemote reads a config from url
func LoadRemote(url string, expandEnvVars bool, c *Config) error {
	remoteOpts := &remoteOpts{}
	if c.BasicAuthUser != "" && c.BasicAuthPassFile != "" {
		remoteOpts.HTTPClientConfig = &config.HTTPClientConfig{
			BasicAuth: &config.BasicAuth{
				Username:     c.BasicAuthUser,
				PasswordFile: c.BasicAuthPassFile,
			},
		}
	}

	if remoteOpts.HTTPClientConfig != nil {
		dir, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current working directory: %w", err)
		}
		remoteOpts.HTTPClientConfig.SetDirectory(dir)
	}

	rc, err := newRemoteConfig(url, remoteOpts)
	if err != nil {
		return fmt.Errorf("error reading remote config: %w", err)
	}
	// fall back to file if no scheme is passed
	if rc == nil {
		return LoadFile(url, expandEnvVars, c)
	}
	bb, err := rc.retrieve()
	if err != nil {
		return fmt.Errorf("error retrieving remote config: %w", err)
	}
	return LoadBytes(bb, expandEnvVars, c)
}

// LoadDynamicConfiguration is used to load configuration from a variety of sources using
// dynamic loader, this is a templated approach
func LoadDynamicConfiguration(url string, expandvar bool, c *Config) error {
	if expandvar {
		return errors.New("expand var is not supported when using dynamic configuration, use gomplate env instead")
	}
	cmf, err := NewDynamicLoader()
	if err != nil {
		return err
	}
	err = cmf.LoadConfigByPath(url)
	if err != nil {
		return err
	}

	err = cmf.ProcessConfigs(c)
	if err != nil {
		return fmt.Errorf("error processing config templates %w", err)
	}
	return nil
}

// LoadBytes unmarshals a config from a buffer. Defaults are not
// applied to the file and must be done manually if LoadBytes
// is called directly.
func LoadBytes(buf []byte, expandEnvVars bool, c *Config) error {
	// (Optionally) expand with environment variables
	if expandEnvVars {
		s, err := envsubst.Eval(string(buf), getenv)
		if err != nil {
			return fmt.Errorf("unable to substitute config with environment variables: %w", err)
		}
		buf = []byte(s)
	}
	bb := bytes.Buffer{}
	bb.Write(buf)
	dec := yaml.NewDecoder(&bb)
	dec.KnownFields(true)
	return dec.Decode(c)
}

// getenv is a wrapper around os.Getenv that ignores patterns that are numeric
// regex capture groups (ie "${1}").
func getenv(name string) string {
	numericName := true

	for _, r := range name {
		if !unicode.IsDigit(r) {
			numericName = false
			break
		}
	}

	if numericName {
		// We need to add ${} back in since envsubst removes it.
		return fmt.Sprintf("${%s}", name)
	}
	return os.Getenv(name)
}

// Load loads a config file from a flagset. Flags will be registered
// to the flagset before parsing them with the values specified by
// args.
func Load(fs *flag.FlagSet, args []string) (*Config, error) {
	return load(fs, args, func(url string, expand bool, c *Config) error {
		if features.Enabled(fs, featRemoteConfigs) {
			return LoadRemote(url, expand, c)
		}
		if features.Enabled(fs, featDynamicConfig) && !features.Enabled(fs, featIntegrationsNext) {
			return fmt.Errorf("integrations-next must be enabled for dynamic configuration to work")
		} else if features.Enabled(fs, featDynamicConfig) {
			return LoadDynamicConfiguration(url, expand, c)
		}
		return LoadFile(url, expand, c)
	})
}

// load allows for tests to inject a function for retrieving the config file that
// doesn't require having a literal file on disk.
func load(fs *flag.FlagSet, args []string, loader func(string, bool, *Config) error) (*Config, error) {
	var (
		cfg = DefaultConfig

		printVersion      bool
		file              string
		dynamicConfigPath string
		configExpandEnv   bool
	)

	fs.StringVar(&file, "config.file", "", "configuration file to load")
	fs.StringVar(&dynamicConfigPath, "config.dynamic-config-path", "", "dynamic configuration path that points to a single configuration file supports file:// or s3:// protocols. Must be enabled by -enable-features=dynamic-config,integrations-next")

	fs.BoolVar(&printVersion, "version", false, "Print this build's version information")
	fs.BoolVar(&configExpandEnv, "config.expand-env", false, "Expands ${var} in config according to the values of the environment variables.")
	cfg.RegisterFlags(fs)
	features.Register(fs, allFeatures)

	if err := fs.Parse(args); err != nil {
		return nil, fmt.Errorf("error parsing flags: %w", err)
	}

	if printVersion {
		fmt.Println(version.Print("agent"))
		os.Exit(0)
	}

	if features.Enabled(fs, featDynamicConfig) {
		if dynamicConfigPath == "" {
			return nil, fmt.Errorf("-config.dynamic-config-path flag required when using dynamic configuration")
		} else if err := loader(dynamicConfigPath, configExpandEnv, &cfg); err != nil {
			return nil, fmt.Errorf("error loading dynamic configuration file %s: %w", dynamicConfigPath, err)
		}
	} else if file == "" {
		return nil, fmt.Errorf("-config.file flag required")
	} else if err := loader(file, configExpandEnv, &cfg); err != nil {
		return nil, fmt.Errorf("error loading config file %s: %w", file, err)
	}

	// Parse the flags again to override any YAML values with command line flag
	// values.
	if err := fs.Parse(args); err != nil {
		return nil, fmt.Errorf("error parsing flags: %w", err)
	}

	// Complete unmarshaling integrations using the version from the flag. This
	// MUST be called before ApplyDefaults.
	version := interfaces.IntegrationsVersion1
	if features.Enabled(fs, featIntegrationsNext) {
		version = interfaces.IntegrationsVersion2
	}

	if err := cfg.Integrations.setVersion(version); err != nil {
		return nil, fmt.Errorf("error loading config file %s: %w", file, err)
	}

	// Finally, apply defaults to config that wasn't specified by file or flag
	if err := cfg.Validate(fs); err != nil {
		return nil, fmt.Errorf("error in config file: %w", err)
	}
	return &cfg, nil
}

// CheckSecret is a helper function to ensure the original value is overwritten with <secret>
func CheckSecret(t *testing.T, rawCfg string, originalValue string) {
	var cfg = &Config{}
	err := LoadBytes([]byte(rawCfg), false, cfg)
	require.NoError(t, err)
	bb, err := yaml.Marshal(cfg)
	require.NoError(t, err)
	scrubbedCfg := string(bb)
	require.True(t, strings.Contains(scrubbedCfg, "<secret>"))
	require.False(t, strings.Contains(scrubbedCfg, originalValue))
}
