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

	"github.com/cortexproject/cortex/pkg/ring/kv"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/prometheus/config"
)

var (
	instanceAbnormalExits = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "agent_prometheus_instance_abnormal_exits_total",
		Help: "Total number of times a Prometheus instance exited unexpectedly, causing it to be restarted.",
	}, []string{"instance_name"})
)

var (
	DefaultConfig = Config{
		Global:                 config.DefaultGlobalConfig,
		InstanceRestartBackoff: 5 * time.Second,
	}
)

// Config defines the configuration for the entire set of Prometheus client
// instances, along with a global configuration.
type Config struct {
	Global                 config.GlobalConfig `yaml:"global"`
	WALDir                 string              `yaml:"wal_directory"`
	ServiceConfig          ServiceConfig       `yaml:"service"`
	Configs                []InstanceConfig    `yaml:"configs,omitempty"`
	InstanceRestartBackoff time.Duration       `yaml:"instance_restart_backoff,omitempty"`
}

func (c *Config) UnmarshalYAML(unmarshal func(interface{}) error) error {
	*c = DefaultConfig

	// We want to set c to the defaults and then overwrite it with the input.
	// To make unmarshal fill the plain data struct rather than calling UnmarshalYAML
	// again, we have to hide it using a type indirection.
	type plain Config
	if err := unmarshal((*plain)(c)); err != nil {
		return err
	}

	for i := range c.Configs {
		c.Configs[i].ApplyDefaults(&c.Global)
	}

	return nil
}

// RegisterFlags defines flags corresponding to the Config.
func (c *Config) RegisterFlags(f *flag.FlagSet) {
	f.StringVar(&c.WALDir, "prometheus.wal-directory", "", "base directory to store the WAL in")
	f.DurationVar(&c.InstanceRestartBackoff, "prometheus.instance-restart-backoff", DefaultConfig.InstanceRestartBackoff, "how long to wait before restarting a failed Prometheus instance")

	c.ServiceConfig.RegisterFlagsWithPrefix("prometheus.service.", f)
}

// Validate checks if the Config has all required fields filled out.
func (c *Config) Validate() error {
	if c.WALDir == "" {
		return errors.New("no wal_directory configured")
	}

	usedNames := map[string]struct{}{}

	for i, cfg := range c.Configs {
		if _, ok := usedNames[cfg.Name]; ok {
			return fmt.Errorf(
				"prometheus instance names must be unique. found multiple instances with name %s",
				cfg.Name,
			)
		}
		usedNames[cfg.Name] = struct{}{}

		if err := cfg.Validate(); err != nil {
			return fmt.Errorf("error validating instance %d: %s", i, err)
		}
	}

	return nil
}

// ServiceConfig describes the configuration for the scraping service.
type ServiceConfig struct {
	Enabled bool      `yaml:"enabled"`
	KVStore kv.Config `yaml:"kvstore"`
}

// RegisterFlagsWithPrefix adds the flags required to config this to the given
// FlagSet with a specified prefix.
func (c *ServiceConfig) RegisterFlagsWithPrefix(prefix string, f *flag.FlagSet) {
	c.KVStore.RegisterFlagsWithPrefix(prefix, "configurations/", f)
	f.BoolVar(&c.Enabled, prefix+"enabled", false, "enables the scraping service mode")
}

// Agent is an agent for collecting Prometheus metrics. It acts as a
// Prometheus-lite; only running the service discovery, remote_write,
// and WAL components of Prometheus. It is broken down into a series
// of Instances, each of which perform metric collection.
type Agent struct {
	cfg    Config
	logger log.Logger

	instanceMtx sync.Mutex
	instances   []instance

	instanceFactory instanceFactory

	kv kv.Client
}

// New creates and starts a new Agent.
func New(cfg Config, logger log.Logger) (*Agent, error) {
	return newAgent(cfg, logger, defaultInstanceFactory)
}

func newAgent(cfg Config, logger log.Logger, fact instanceFactory) (*Agent, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	a := &Agent{
		cfg:             cfg,
		logger:          log.With(logger, "agent", "prometheus"),
		instanceFactory: fact,
	}

	for _, c := range cfg.Configs {
		inst, err := fact(cfg.Global, c, cfg.WALDir, a.logger)
		if err != nil {
			return nil, err
		}
		a.instances = append(a.instances, inst)
	}

	if cfg.ServiceConfig.Enabled {
		var err error
		a.kv, err = kv.NewClient(cfg.ServiceConfig.KVStore, GetCodec())
		if err != nil {
			return nil, err
		}
	}

	go a.run()
	return a, nil
}

func (a *Agent) run() {
	// This function watches all instances for abnormal shutdowns and restarts them
	// whenever that's detected. This function only exits when all instances
	// shutdown normally, which can only happen when Stop is called on the agent.
	a.forAllInstances(func(i int, _ instance) {
		for {
			inst := a.instances[i]
			err := inst.Wait()

			if err == nil || err != errInstanceStoppedNormally {
				instanceAbnormalExits.WithLabelValues(inst.Config().Name).Inc()
				level.Error(a.logger).Log("msg", "instance stopped abnormally, restarting after backoff period", "err", err, "backoff", a.cfg.InstanceRestartBackoff)
				time.Sleep(a.cfg.InstanceRestartBackoff)
			} else {
				level.Info(a.logger).Log("msg", "agent stopped normally")
				return
			}

			// Try to recreate the instance.
			cfg := inst.Config()
			inst, err = a.instanceFactory(a.cfg.Global, cfg, a.cfg.WALDir, a.logger)
			if err != nil {
				level.Error(a.logger).Log("msg", "failed to recreate instance", "err", err)
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
	a.forAllInstances(func(idx int, inst instance) {
		inst.Stop()
	})
}

// forAllInstances runs f in parallel for all provided instances. Only returns when
// all f exit.
func (a *Agent) forAllInstances(f func(idx int, inst instance)) {
	var wg sync.WaitGroup
	wg.Add(len(a.instances))

	a.instanceMtx.Lock()
	for idx, inst := range a.instances {
		go func(idx int, inst instance) {
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

// instance is an interface implemented by Instance, and used by tests
// to isolate agent from instance functionality.
type instance interface {
	Wait() error
	Config() InstanceConfig
	Stop()
}

type instanceFactory = func(global config.GlobalConfig, cfg InstanceConfig, walDir string, logger log.Logger) (instance, error)

func defaultInstanceFactory(global config.GlobalConfig, cfg InstanceConfig, walDir string, logger log.Logger) (instance, error) {
	return NewInstance(global, cfg, walDir, logger)
}
