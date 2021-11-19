package config

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"unicode"

	"github.com/cortexproject/cortex/pkg/ring/kv/consul"
	"github.com/drone/envsubst/v2"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/grafana/agent/pkg/integrations"
	"github.com/grafana/agent/pkg/logs"
	"github.com/grafana/agent/pkg/metrics"
	"github.com/grafana/agent/pkg/traces"
	"github.com/grafana/agent/pkg/util"
	"github.com/pkg/errors"
	"github.com/prometheus/common/version"
	"github.com/weaveworks/common/server"
	"gopkg.in/yaml.v2"
)

// DefaultConfig holds default settings for all the subsystems.
var DefaultConfig = Config{
	// All subsystems with a DefaultConfig should be listed here.
	Metrics:      metrics.DefaultConfig,
	Integrations: integrations.DefaultManagerConfig,
}

// Config contains underlying configurations for the agent
type Config struct {
	Server       server.Config              `yaml:"server,omitempty"`
	Metrics      metrics.Config             `yaml:"metrics,omitempty"`
	Integrations integrations.ManagerConfig `yaml:"integrations,omitempty"`
	Traces       traces.Config              `yaml:"traces,omitempty"`
	Logs         *logs.Config               `yaml:"logs,omitempty"`

	// We support a secondary server just for the /-/reload endpoint, since
	// invoking /-/reload against the primary server can cause the server
	// to restart.
	ReloadAddress string `yaml:"-"`
	ReloadPort    int    `yaml:"-"`

	// Deprecated fields user has used. Generated during UnmarshalYAML.
	Deprecations []string `yaml:"-"`
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
func (c *Config) MarshalYAML() (interface{}, error) {
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
	if err := enc.Encode((*config)(c)); err != nil {
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

// ApplyDefaults sets default values in the config
func (c *Config) ApplyDefaults() error {
	if err := c.Metrics.ApplyDefaults(); err != nil {
		return err
	}

	if err := c.Integrations.ApplyDefaults(&c.Metrics); err != nil {
		return err
	}

	c.Metrics.ServiceConfig.Lifecycler.ListenPort = c.Server.GRPCListenPort
	c.Integrations.ListenPort = c.Server.HTTPListenPort
	c.Integrations.ListenHost = c.Server.HTTPListenAddress

	c.Integrations.ServerUsingTLS = c.Server.HTTPTLSConfig.TLSKeyPath != "" && c.Server.HTTPTLSConfig.TLSCertPath != ""

	if len(c.Integrations.PrometheusRemoteWrite) == 0 {
		c.Integrations.PrometheusRemoteWrite = c.Metrics.Global.RemoteWrite
	}

	c.Integrations.PrometheusGlobalConfig = c.Metrics.Global.Prometheus

	// since the Traces config might rely on an existing Loki config
	// this check is made here to look for cross config issues before we attempt to load
	if err := c.Traces.Validate(c.Logs); err != nil {
		return err
	}

	return nil
}

// RegisterFlags registers flags in underlying configs
func (c *Config) RegisterFlags(f *flag.FlagSet) {
	c.Server.MetricsNamespace = "agent"
	c.Server.RegisterInstrumentation = true
	c.Metrics.RegisterFlags(f)
	c.Server.RegisterFlags(f)

	f.StringVar(&c.ReloadAddress, "reload-addr", "127.0.0.1", "address to expose a secondary server for /-/reload on.")
	f.IntVar(&c.ReloadPort, "reload-port", 0, "port to expose a secondary server for /-/reload on. 0 disables secondary server.")
}

// LoadFile reads a file and passes the contents to Load
func LoadFile(filename string, expandEnvVars bool, c *Config) error {
	buf, err := ioutil.ReadFile(filename)
	if err != nil {
		return errors.Wrap(err, "error reading config file")
	}
	return LoadBytes(buf, expandEnvVars, c)
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
	return load(fs, args, LoadFile)
}

// load allows for tests to inject a function for retrieving the config file that
// doesn't require having a literal file on disk.
func load(fs *flag.FlagSet, args []string, loader func(string, bool, *Config) error) (*Config, error) {
	var (
		cfg = DefaultConfig

		printVersion    bool
		file            string
		configExpandEnv bool
	)

	fs.StringVar(&file, "config.file", "", "configuration file to load")
	fs.BoolVar(&printVersion, "version", false, "Print this build's version information")
	fs.BoolVar(&configExpandEnv, "config.expand-env", false, "Expands ${var} in config according to the values of the environment variables.")
	cfg.RegisterFlags(fs)

	if err := fs.Parse(args); err != nil {
		return nil, fmt.Errorf("error parsing flags: %w", err)
	}

	if printVersion {
		fmt.Println(version.Print("agent"))
		os.Exit(0)
	}

	if file == "" {
		return nil, fmt.Errorf("-config.file flag required")
	} else if err := loader(file, configExpandEnv, &cfg); err != nil {
		return nil, fmt.Errorf("error loading config file %s: %w", file, err)
	}

	// Parse the flags again to override any YAML values with command line flag
	// values
	if err := fs.Parse(args); err != nil {
		return nil, fmt.Errorf("error parsing flags: %w", err)
	}

	// Finally, apply defaults to config that wasn't specified by file or flag
	if err := cfg.ApplyDefaults(); err != nil {
		return nil, fmt.Errorf("error in config file: %w", err)
	}

	return &cfg, nil
}
