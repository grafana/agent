package config

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/grafana/agent/pkg/integrations"
	"github.com/grafana/agent/pkg/loki"
	"github.com/grafana/agent/pkg/prom"
	"github.com/grafana/agent/pkg/tempo"
	"github.com/pkg/errors"
	"github.com/prometheus/common/version"
	"github.com/weaveworks/common/server"
	"gopkg.in/yaml.v2"
)

// Config contains underlying configurations for the agent
type Config struct {
	Server       server.Config       `yaml:"server"`
	Prometheus   prom.Config         `yaml:"prometheus,omitempty"`
	Loki         loki.Config         `yaml:"loki,omitempty"`
	Integrations integrations.Config `yaml:"integrations"`
	Tempo        tempo.Config        `yaml:"tempo,omitempty"`
}

// ApplyDefaults sets default values in the config
func (c *Config) ApplyDefaults() error {
	// The integration subsystem depends on Prometheus; so if it's enabled, force Prometheus
	// to be enabled.
	//
	// TODO(rfratto): when Loki integrations are added, this line will no longer work; each
	// integration will then have to be associated with a subsystem.
	if c.Integrations.Enabled && !c.Prometheus.Enabled {
		fmt.Println("NOTE: enabling Prometheus subsystem as Integrations are enabled")
		c.Prometheus.Enabled = true
	}

	if c.Prometheus.Enabled {
		if err := c.Prometheus.ApplyDefaults(); err != nil {
			return err
		}

		// The default port exposed to the lifecycler should be the gRPC listen
		// port since the agents will use gRPC for notifying other agents of
		// resharding.
		c.Prometheus.ServiceConfig.Lifecycler.ListenPort = c.Server.GRPCListenPort
	}

	if c.Integrations.Enabled {
		c.Integrations.ListenPort = &c.Server.HTTPListenPort

		// Apply the defaults for integrations *after* manual defaults are applied
		// (like the c.Integrations.ListenPort) above.
		if err := c.Integrations.ApplyDefaults(); err != nil {
			return err
		}
	}

	return nil
}

// RegisterFlags registers flags in underlying configs
func (c *Config) RegisterFlags(f *flag.FlagSet) {
	c.Server.MetricsNamespace = "agent"
	c.Server.RegisterInstrumentation = true
	c.Prometheus.RegisterFlags(f)
	c.Server.RegisterFlags(f)
	c.Loki.RegisterFlags(f)
}

// LoadFile reads a file and passes the contents to Load
func LoadFile(filename string, c *Config) error {
	buf, err := ioutil.ReadFile(filename)
	if err != nil {
		return errors.Wrap(err, "error reading config file")
	}

	return LoadBytes(buf, c)
}

// LoadBytes unmarshals a config from a buffer. Defaults are not
// applied to the file and must be done manually if LoadBytes
// is called directly.
func LoadBytes(buf []byte, c *Config) error {
	return yaml.UnmarshalStrict(buf, c)
}

// Load loads a config file from a flagset. Flags will be registered
// to the flagset before parsing them with the values specified by
// args.
func Load(fs *flag.FlagSet, args []string) (*Config, error) {
	return load(fs, args, LoadFile)
}

// load allows for tests to inject a function for retreiving the config file that
// doesn't require having a literal file on disk.
func load(fs *flag.FlagSet, args []string, loader func(string, *Config) error) (*Config, error) {
	var (
		printVersion bool
		cfg          Config
		file         string
	)

	fs.StringVar(&file, "config.file", "", "configuration file to load")
	fs.BoolVar(&printVersion, "version", false, "Print this build's version information")
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
	} else if err := loader(file, &cfg); err != nil {
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
