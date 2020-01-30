// Package prometheus implements a Prometheus-lite client for service discovery,
// scraping metrics into a WAL, and remote_write. Clients are broken into a
// set of instances, each of which contain their own set of configs.
package prometheus

import (
	"errors"
	"flag"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/prometheus/config"
)

// Config defines the configuration for the entire set of Prometheus client
// instances, along with a global configuration.
type Config struct {
	Global  config.GlobalConfig `yaml:"global"`
	WALDir  string              `yaml:"wal_directory"`
	Configs []InstanceConfig    `yaml:"configs,omitempty"`
}

func (c *Config) UnmarshalYAML(unmarshal func(interface{}) error) error {
	c.Global = config.DefaultGlobalConfig

	// We want to set c to the defaults and then overwrite it with the input.
	// To make unmarshal fill the plain data struct rather than calling UnmarshalYAML
	// again, we have to hide it using a type indirection.
	type plain Config
	if err := unmarshal((*plain)(c)); err != nil {
		return err
	}

	if zeroGlobalConfig(c.Global) {
		c.Global = config.DefaultGlobalConfig
	}

	// TODO(rfratto): move rest of ApplyDefaults to UnmarshalYAML
	return nil
}

// RegisterFlags defines flags corresponding to the Config.
func (c *Config) RegisterFlags(f *flag.FlagSet) {
	f.StringVar(&c.WALDir, "prometheus.wal-directory", "", "base directory to store the WAL in")
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
	if c.WALDir == "" {
		return errors.New("no wal_directory configured")
	}

	for i, cfg := range c.Configs {
		if err := cfg.Validate(); err != nil {
			return fmt.Errorf("error validating instance %d: %s", i, err)
		}
	}

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
	cfg    Config
	logger log.Logger

	instanceMtx sync.Mutex
	instances   []*instance
}

// New creates and starts a new Agent.
func New(cfg Config, logger log.Logger) (*Agent, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	a := &Agent{
		cfg:    cfg,
		logger: log.With(logger, "agent", "prometheus"),
	}

	for _, c := range cfg.Configs {
		inst, err := newInstance(cfg.Global, c, cfg.WALDir, a.logger)
		if err != nil {
			return nil, err
		}
		a.instances = append(a.instances, inst)
	}

	go a.run()
	return a, nil
}

func (a *Agent) run() {
	// This function watches all instances for abnormal shutdowns and restarts them
	// whenever that's detected. This function only exits when all instances
	// shutdown normally, which can only happen when Stop is called on the agent.
	a.forAllInstances(func(i int, _ *instance) {
		for {
			inst := a.instances[i]
			<-inst.exited

			if err := inst.Err(); err != nil && err != errInstanceStoppedNormally {
				// TODO(rfratto): metric for abnormal instance exits
				level.Error(a.logger).Log("msg", "instance stopped abnormally. restarting in 5 sec...", "err", err)
				time.Sleep(time.Second * 5)
			} else {
				level.Info(a.logger).Log("msg", "agent stopped normally")
				return
			}

			// Try to recreate the instance.
			inst, err := newInstance(a.cfg.Global, inst.cfg, a.cfg.WALDir, a.logger)
			if err != nil {
				level.Error(a.logger).Log("msg", "failed to recreate instance", "err", err)
				// TODO(rfratto): should this be a panic? if we let the function return here
				// then that's an entire agent instance that's lost and won't recover until
				// the entire process is restarted.
				return
			}

			a.instanceMtx.Lock()
			a.instances[i] = inst
			a.instanceMtx.Unlock()
		}
	})
}

// Stop stops the agent and all its instances.
func (a *Agent) Stop() {
	a.forAllInstances(func(idx int, inst *instance) {
		inst.Stop()
	})
}

// forAllInstances runs f in parallel for all provided instances. Only returns when
// all f exit.
func (a *Agent) forAllInstances(f func(idx int, inst *instance)) {
	var wg sync.WaitGroup
	wg.Add(len(a.instances))

	a.instanceMtx.Lock()
	for idx, inst := range a.instances {
		go func(idx int, inst *instance) {
			f(idx, inst)
			wg.Done()
		}(idx, inst)
	}
	a.instanceMtx.Unlock()

	wg.Wait()
}

// MetricValueCollector wraps around a Gatherer and provides utilities for
// pulling metric values from a given metric name and label matchers.
//
// This is used by the agent instances to find the most recent timestamp
// successfully remote_written to for pruposes of safely truncating the WAL.
//
// MetricValueCollector is only intended for use with Gauges and Counters.
type MetricValueCollector struct {
	g     prometheus.Gatherer
	match string
}

// NewMetricValueCollector creates a new MetricValueCollector.
func NewMetricValueCollector(g prometheus.Gatherer, match string) *MetricValueCollector {
	return &MetricValueCollector{
		g:     g,
		match: match,
	}
}

// GetValues looks through all the tracked metrics and returns all values
// for metrics that match some key value pair.
func (vc *MetricValueCollector) GetValues(label string, labelValues ...string) ([]float64, error) {
	vals := []float64{}

	families, err := vc.g.Gather()
	if err != nil {
		return nil, err
	}

	for _, family := range families {
		if !strings.Contains(family.GetName(), vc.match) {
			continue
		}

		for _, m := range family.GetMetric() {
			matches := false
			for _, l := range m.GetLabel() {
				if l.GetName() != label {
					continue
				}

				v := l.GetValue()
				for _, match := range labelValues {
					if match == v {
						matches = true
						break
					}
				}
				break
			}
			if !matches {
				continue
			}

			var value float64
			if m.Gauge != nil {
				value = m.Gauge.GetValue()
			} else if m.Counter != nil {
				value = m.Counter.GetValue()
			} else if m.Untyped != nil {
				value = m.Untyped.GetValue()
			} else {
				return nil, errors.New("tracking unexpected metric type")
			}

			vals = append(vals, value)
		}
	}

	return vals, nil
}
