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

	"github.com/drone/envsubst/v2"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/grafana/agent/pkg/config/features"
	"github.com/grafana/agent/pkg/logs"
	"github.com/grafana/agent/pkg/metrics"
	"github.com/grafana/agent/pkg/server"
	"github.com/grafana/agent/pkg/traces"
	"github.com/grafana/agent/pkg/util"
	"github.com/grafana/dskit/kv/consul"
	"github.com/prometheus/common/config"
	"github.com/prometheus/common/version"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
)

var (
	featRemoteConfigs    = features.Feature("remote-configs")
	featIntegrationsNext = features.Feature("integrations-next")
	featDynamicConfig    = features.Feature("dynamic-config")
	featExtraMetrics     = features.Feature("extra-scrape-metrics")

	allFeatures = []features.Feature{
		featRemoteConfigs,
		featIntegrationsNext,
		featDynamicConfig,
		featExtraMetrics,
	}
)

var (
	fileTypeYAML    = "yaml"
	fileTypeDynamic = "dynamic"

	fileTypes = []string{fileTypeYAML, fileTypeDynamic}
)

// DefaultConfig holds default settings for all the subsystems.
var DefaultConfig = Config{
	// All subsystems with a DefaultConfig should be listed here.
	Server:                server.DefaultConfig,
	Metrics:               metrics.DefaultConfig,
	Integrations:          DefaultVersionedIntegrations,
	EnableConfigEndpoints: false,
	EnableUsageReport:     true,
}

// Config contains underlying configurations for the agent
type Config struct {
	Server       server.Config         `yaml:"server,omitempty"`
	Metrics      metrics.Config        `yaml:"metrics,omitempty"`
	Integrations VersionedIntegrations `yaml:"integrations,omitempty"`
	Traces       traces.Config         `yaml:"traces,omitempty"`
	Logs         *logs.Config          `yaml:"logs,omitempty"`

	// Deprecated fields user has used. Generated during UnmarshalYAML.
	Deprecations []string `yaml:"-"`

	// Remote config options
	BasicAuthUser     string `yaml:"-"`
	BasicAuthPassFile string `yaml:"-"`

	// Toggle for config endpoint(s)
	EnableConfigEndpoints bool `yaml:"-"`

	// Report enabled features options
	EnableUsageReport bool     `yaml:"-"`
	EnabledFeatures   []string `yaml:"-"`
}

// UnmarshalYAML implements yaml.Unmarshaler.
func (c *Config) UnmarshalYAML(unmarshal func(interface{}) error) error {
	// Apply defaults to the config from our struct and any defaults inherited
	// from flags before unmarshaling.
	*c = DefaultConfig
	util.DefaultConfigFromFlags(c)

	type baseConfig Config

	type config struct {
		baseConfig `yaml:",inline"`

		// Deprecated field names:
		Prometheus *metrics.Config `yaml:"prometheus,omitempty"`
		Loki       *logs.Config    `yaml:"loki,omitempty"`
		Tempo      *traces.Config  `yaml:"tempo,omitempty"`
	}

	var fc config
	fc.baseConfig = baseConfig(*c)

	if err := unmarshal(&fc); err != nil {
		return err
	}

	// Migrate old fields to the new name
	if fc.Prometheus != nil && fc.Metrics.Unmarshaled && fc.Prometheus.Unmarshaled {
		return fmt.Errorf("at most one of prometheus and metrics should be specified")
	} else if fc.Prometheus != nil && fc.Prometheus.Unmarshaled {
		fc.Deprecations = append(fc.Deprecations, "`prometheus` has been deprecated in favor of `metrics`")
		fc.Metrics = *fc.Prometheus
		fc.Prometheus = nil
	}

	if fc.Logs != nil && fc.Loki != nil {
		return fmt.Errorf("at most one of loki and logs should be specified")
	} else if fc.Logs == nil && fc.Loki != nil {
		fc.Deprecations = append(fc.Deprecations, "`loki` has been deprecated in favor of `logs`")
		fc.Logs = fc.Loki
		fc.Loki = nil
	}

	if fc.Tempo != nil && fc.Traces.Unmarshaled {
		return fmt.Errorf("at most one of tempo and traces should be specified")
	} else if fc.Tempo != nil && fc.Tempo.Unmarshaled {
		fc.Deprecations = append(fc.Deprecations, "`tempo` has been deprecated in favor of `traces`")
		fc.Traces = *fc.Tempo
		fc.Tempo = nil
	}

	*c = Config(fc.baseConfig)
	return nil
}

// MarshalYAML implements yaml.Marshaler.
func (c Config) MarshalYAML() (interface{}, error) {
	var buf bytes.Buffer

	enc := yaml.NewEncoder(&buf)
	enc.SetHook(func(in interface{}) (ok bool, out interface{}, err error) {
		// Obscure the password fields for known types that do not obscure passwords.
		switch v := in.(type) {
		case consul.Config:
			v.ACLToken = "<secret>"
			return true, v, nil
		default:
			return false, nil, nil
		}
	})

	type config Config
	if err := enc.Encode((config)(c)); err != nil {
		return nil, err
	}

	// Use a yaml.MapSlice rather than a map[string]interface{} so
	// order of keys is retained compared to just calling MarshalConfig.
	var m yaml.MapSlice
	if err := yaml.Unmarshal(buf.Bytes(), &m); err != nil {
		return nil, err
	}
	return m, nil
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

	c.Metrics.ServiceConfig.Lifecycler.ListenPort = c.Server.Flags.GRPC.ListenPort

	if err := c.Integrations.ApplyDefaults(&c.Server, &c.Metrics); err != nil {
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
	c.Metrics.RegisterFlags(f)
	c.Server.RegisterFlags(f)

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
	// Unmarshal yaml config
	return yaml.UnmarshalStrict(buf, c)
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
	return load(fs, args, func(path, fileType string, expandArgs bool, c *Config) error {
		switch fileType {
		case fileTypeYAML:
			if features.Enabled(fs, featRemoteConfigs) {
				return LoadRemote(path, expandArgs, c)
			}
			return LoadFile(path, expandArgs, c)
		case fileTypeDynamic:
			if !features.Enabled(fs, featDynamicConfig) {
				return fmt.Errorf("feature %q must be enabled to use file type %s", featDynamicConfig, fileTypeDynamic)
			} else if !features.Enabled(fs, featIntegrationsNext) {
				return fmt.Errorf("feature %q must be enabled to use file type %s", featIntegrationsNext, fileTypeDynamic)
			} else if features.Enabled(fs, featRemoteConfigs) {
				return fmt.Errorf("feature %q can not be enabled with file type %s", featRemoteConfigs, fileTypeDynamic)
			} else if expandArgs {
				return fmt.Errorf("-config.expand-env can not be used with file type %s", fileTypeDynamic)
			}
			return LoadDynamicConfiguration(path, expandArgs, c)
		default:
			return fmt.Errorf("unknown file type %q. accepted values: %s", fileType, strings.Join(fileTypes, ", "))
		}
	})
}

type loaderFunc func(path string, fileType string, expandArgs bool, target *Config) error

// load allows for tests to inject a function for retrieving the config file that
// doesn't require having a literal file on disk.
func load(fs *flag.FlagSet, args []string, loader loaderFunc) (*Config, error) {
	var (
		cfg = DefaultConfig

		printVersion     bool
		file             string
		fileType         string
		configExpandEnv  bool
		disableReporting bool
	)

	fs.StringVar(&file, "config.file", "", "configuration file to load")
	fs.StringVar(&fileType, "config.file.type", "yaml", fmt.Sprintf("Type of file pointed to by -config.file flag. Supported values: %s. %s requires dynamic-config and integrations-next features to be enabled.", strings.Join(fileTypes, ", "), fileTypeDynamic))
	fs.BoolVar(&printVersion, "version", false, "Print this build's version information.")
	fs.BoolVar(&configExpandEnv, "config.expand-env", false, "Expands ${var} in config according to the values of the environment variables.")
	fs.BoolVar(&disableReporting, "disable-reporting", false, "Disable reporting of enabled feature flags to Grafana.")
	cfg.RegisterFlags(fs)

	features.Register(fs, allFeatures)

	if err := fs.Parse(args); err != nil {
		return nil, fmt.Errorf("error parsing flags: %w", err)
	}

	if printVersion {
		fmt.Println(version.Print("agent"))
		os.Exit(0)
	}

	if file == "" {
		return nil, fmt.Errorf("-config.file flag required")
	} else if err := loader(file, fileType, configExpandEnv, &cfg); err != nil {
		return nil, fmt.Errorf("error loading config file %s: %w", file, err)
	}

	// Parse the flags again to override any YAML values with command line flag
	// values.
	if err := fs.Parse(args); err != nil {
		return nil, fmt.Errorf("error parsing flags: %w", err)
	}

	// Complete unmarshaling integrations using the version from the flag. This
	// MUST be called before ApplyDefaults.
	version := integrationsVersion1
	if features.Enabled(fs, featIntegrationsNext) {
		version = integrationsVersion2
	}

	if err := cfg.Integrations.setVersion(version); err != nil {
		return nil, fmt.Errorf("error loading config file %s: %w", file, err)
	}

	if features.Enabled(fs, featExtraMetrics) {
		cfg.Metrics.Global.ExtraMetrics = true
	}

	if disableReporting {
		cfg.EnableUsageReport = false
	} else {
		cfg.EnabledFeatures = features.GetAllEnabled(fs)
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
