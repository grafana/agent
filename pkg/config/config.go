package config

import (
	"flag"
	"io/ioutil"

	prom "github.com/grafana/agent/pkg/prometheus"
	"github.com/pkg/errors"
	"github.com/weaveworks/common/server"
	"gopkg.in/yaml.v2"
)

// Config contains underlying configurations for the agent
type Config struct {
	Server     server.Config `yaml:"server"`
	Prometheus prom.Config   `yaml:"prometheus,omitempty"`
}

func (c *Config) ApplyDefaults() {
	c.Prometheus.ApplyDefaults()
}

// RegisterFlags registers flags in underlying configs
func (c *Config) RegisterFlags(f *flag.FlagSet) {
	c.Server.MetricsNamespace = "agent"
	c.Server.RegisterInstrumentation = true
	c.Prometheus.RegisterFlags(f)
	c.Server.RegisterFlags(f)
}

func LoadFile(filename string, c *Config) error {
	buf, err := ioutil.ReadFile(filename)
	if err != nil {
		return errors.Wrap(err, "error reading config file")
	}

	return Load(buf, c)
}

// Load loads a config and applies defaults
func Load(buf []byte, c *Config) error {
	err := yaml.UnmarshalStrict(buf, c)
	if err != nil {
		return err
	}

	c.ApplyDefaults()

	return nil
}
