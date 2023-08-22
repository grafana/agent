// Package prometheus implements a Prometheus-lite client for service discovery,
// scraping metrics into a WAL, and remote_write. Clients are broken into a
// set of instances, each of which contain their own set of configs.
package prometheus

import (
	"github.com/go-kit/kit/log"
	"github.com/prometheus/prometheus/config"
)

// Config defines the configuration for the entire set of Prometheus client
// instances, along with a global configuration.
type Config struct {
	Global  config.GlobalConfig `yaml:"global"`
	Configs []InstanceConfig    `yaml:"configs,omitempty"`
}

// ApplyDefaults applies default configurations to the configuration to all
// values that have not been changed to their non-zero value.
func (c *Config) ApplyDefaults() {
	if zeroGlobalConfig(c.Global) {
		c.Global = config.DefaultGlobalConfig
	}

	for i := range c.Configs {
		c.Configs[i].ApplyDefaults(&c.Global)
	}
}

// Validate checks if the Config has all required fields filled out. This
// should only be called after ApplyDefaults.
func (c *Config) Validate() error {
	// TODO(rfratto): validation
	return nil
}

// zeroGlobalConfig checks if a GlobalConfig is unchanged from
// all zero values. Copied from Prometheus.
func zeroGlobalConfig(c config.GlobalConfig) bool {
	return c.ExternalLabels == nil &&
		c.ScrapeInterval == 0 &&
		c.ScrapeTimeout == 0 &&
		c.EvaluationInterval == 0
}

// Agent is an agent for collecting Prometheus metrics. It acts as a
// Prometheus-lite; only running the service discovery, remote_write,
// and WAL components of Prometheus. It is broken down into a series
// of Instances, each of which perform metric collection.
type Agent struct {
	cfg       Config
	logger    log.Logger
	instances []*instance
}

// New creates and starts a new Agent.
func New(cfg Config, logger log.Logger) *Agent {
	// TODO(rfratto): validate config here or have the invoker be responsible?
	a := &Agent{
		cfg:    cfg,
		logger: log.With(logger, "agent", "prometheus"),
	}

	for _, c := range cfg.Configs {
		a.instances = append(a.instances, newInstance(cfg.Global, c, a.logger))
	}

	return a
}

// Stop stops the agent.
func (a *Agent) Stop() {
	for _, i := range a.instances {
		i.Stop()
	}
}
