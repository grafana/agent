package config

import (
	"flag"
	"io/ioutil"

	"github.com/grafana/agent/pkg/integrations"
	"github.com/grafana/agent/pkg/prom"
	"github.com/pkg/errors"
	"github.com/weaveworks/common/server"
	"gopkg.in/yaml.v2"
)

// Config contains underlying configurations for the agent
type Config struct {
	Server       server.Config       `yaml:"server"`
	Prometheus   prom.Config         `yaml:"prometheus,omitempty"`
	Integrations integrations.Config `yaml:"integrations"`
}

// ApplyDefaults sets default values in the config
func (c *Config) ApplyDefaults() error {
	if err := c.Prometheus.ApplyDefaults(); err != nil {
		return err
	}

	// The default port exposed to the lifecycler should be the gRPC listen
	// port since the agents will use gRPC for notifying other agents of
	// resharding.
	c.Prometheus.ServiceConfig.Lifecycler.ListenPort = &c.Server.GRPCListenPort
	c.Integrations.ListenPort = &c.Server.HTTPListenPort
	return nil
}

// RegisterFlags registers flags in underlying configs
func (c *Config) RegisterFlags(f *flag.FlagSet) {
	c.Server.MetricsNamespace = "agent"
	c.Server.RegisterInstrumentation = true
	c.Prometheus.RegisterFlags(f)
	c.Server.RegisterFlags(f)
}

// LoadFile reads a file and passes the contents to Load
func LoadFile(filename string, c *Config) error {
	buf, err := ioutil.ReadFile(filename)
	if err != nil {
		return errors.Wrap(err, "error reading config file")
	}

	return Load(buf, c)
}

// Load loads a config, but doesn't apply defaults. Defaults
// should be deferred to a separate process to allow flags
// to override values unmarshaled here.
func Load(buf []byte, c *Config) error {
	return yaml.UnmarshalStrict(buf, c)
}
