// Package prometheus implements a Prometheus-lite client for service discovery,
// scraping metrics into a WAL, and remote_write. Clients are broken into a
// set of instances, each of which contain their own set of configs.
package prometheus

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"sync"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/gorilla/mux"
	"github.com/grafana/agent/pkg/prometheus/ha"
	"github.com/grafana/agent/pkg/prometheus/instance"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/prometheus/config"
)

var (
	instanceAbnormalExits = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "agent_prometheus_instance_abnormal_exits_total",
		Help: "Total number of times a Prometheus instance exited unexpectedly, causing it to be restarted.",
	}, []string{"instance_name"})

	currentActiveConfigs = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "agent_prometheus_active_configs",
		Help: "Current number of active configs being used by the agent.",
	})
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
	ServiceConfig          ha.Config           `yaml:"scraping_service"`
	Configs                []instance.Config   `yaml:"configs,omitempty"`
	InstanceRestartBackoff time.Duration       `yaml:"instance_restart_backoff,omitempty"`
}

func (c *Config) ApplyDefaults() {
	for i := range c.Configs {
		c.Configs[i].ApplyDefaults(&c.Global)
	}
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

	if c.ServiceConfig.Enabled && len(c.Configs) > 0 {
		return errors.New("cannot use configs when scraping_service mode is enabled")
	}

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

// Agent is an agent for collecting Prometheus metrics. It acts as a
// Prometheus-lite; only running the service discovery, remote_write,
// and WAL components of Prometheus. It is broken down into a series
// of Instances, each of which perform metric collection.
type Agent struct {
	cfg    Config
	logger log.Logger

	cm *ConfigManager

	instanceFactory instanceFactory

	ha *ha.Server
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

	a.cm = NewConfigManager(a.spawnInstance)
	for _, c := range cfg.Configs {
		a.cm.ApplyConfig(c)
	}

	if cfg.ServiceConfig.Enabled {
		var err error
		a.ha, err = ha.New(cfg.ServiceConfig, a.logger)
		if err != nil {
			return nil, err
		}
	}

	return a, nil
}

// spawnInstance takes an instance.Config and launches an instance, restarting
// it if it stops unexpectedly. The instance will be stopped whenever ctx
// is canceled. This function will not return until the launched instance
// has fully shut down.
func (a *Agent) spawnInstance(ctx context.Context, c instance.Config) {
	var (
		mut  sync.Mutex
		inst inst
		err  error
	)

	// Done is used to make sure the goroutine below doesn't leak.
	done := make(chan bool)
	defer close(done)

	go func() {
		select {
		case <-ctx.Done():
		case <-done:
		}

		mut.Lock()
		defer mut.Unlock()
		if inst != nil {
			inst.Stop()
		}
	}()

	for {
		mut.Lock()
		inst, err = a.instanceFactory(a.cfg.Global, c, a.cfg.WALDir, a.logger)
		if err != nil {
			level.Error(a.logger).Log("msg", "failed to create instance", "err", err)
			return
		}
		mut.Unlock()

		err = inst.Wait()
		if err == nil || err != instance.ErrInstanceStoppedNormally {
			instanceAbnormalExits.WithLabelValues(c.Name).Inc()
			level.Error(a.logger).Log("msg", "instance stopped abnormally, restarting after backoff period", "err", err, "backoff", a.cfg.InstanceRestartBackoff, "instance", c.Name)
			time.Sleep(a.cfg.InstanceRestartBackoff)
		} else {
			level.Info(a.logger).Log("msg", "stopped instance", "instance", c.Name)
			break
		}
	}
}

// WireAPI adds API routes to the provided mux router.
func (a *Agent) WireAPI(r *mux.Router) {
	if a.cfg.ServiceConfig.Enabled {
		a.ha.WireAPI(r)
	}
}

// Stop stops the agent and all its instances.
func (a *Agent) Stop() {
	a.cm.Stop()
}

// ConfigManager manages a set of instance.Configs, calling a function whenever
// a Config should be "started."
type ConfigManager struct {
	// Take care when locking mut: if you hold onto a lock of mut while calling
	// Stop on one of the processes below, you will deadlock.
	mut       sync.Mutex
	processes map[string]configManagerProcess

	newProcess func(ctx context.Context, c instance.Config)
}

type configManagerProcess struct {
	cfg    instance.Config
	cancel context.CancelFunc
	done   chan bool
}

// Stop stops the process and waits for it to exit.
func (p configManagerProcess) Stop() {
	p.cancel()
	<-p.done
}

// NewConfigManager creates a new ConfigManager. The function f will be invoked
// any time a new instance.Config is tracked. The context provided to the function
// will be cancelled when that instance.Config is no longer being tracked.
func NewConfigManager(f func(ctx context.Context, c instance.Config)) *ConfigManager {
	return &ConfigManager{
		processes:  make(map[string]configManagerProcess),
		newProcess: f,
	}
}

// ListConfigs lists the current active configs managed by the ConfigManager.
func (cm *ConfigManager) ListConfigs() map[string]instance.Config {
	cm.mut.Lock()
	defer cm.mut.Unlock()

	cfgs := make(map[string]instance.Config, len(cm.processes))
	for name, process := range cm.processes {
		cfgs[name] = process.cfg
	}
	return cfgs
}

// ApplyConfig takes an instance.Config and either adds a new tracked config
// or updates an existing track config. The value for Name in c is used to
// uniquely identify the instance.Config and determine whether it is new
// or existing.
func (cm *ConfigManager) ApplyConfig(c instance.Config) {
	cm.mut.Lock()
	defer cm.mut.Unlock()

	// Is there an existing process for the instance.Config? If so, stop it.
	if proc, ok := cm.processes[c.Name]; ok {
		proc.Stop()
	}

	// Spawn a new process for the new config.
	cm.spawnProcess(c)
	currentActiveConfigs.Inc()
}

func (cm *ConfigManager) spawnProcess(c instance.Config) {
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan bool)

	cm.processes[c.Name] = configManagerProcess{
		cancel: cancel,
		done:   done,
	}

	go func() {
		cm.newProcess(ctx, c)

		// Delete the process from the tracked map
		cm.mut.Lock()
		delete(cm.processes, c.Name)
		close(done)
		cm.mut.Unlock()
		currentActiveConfigs.Dec()
	}()
}

// DeleteConfig removes an instance.Config by its name. Returns an error if
// the instance.Config is not currently being tracked.
func (cm *ConfigManager) DeleteConfig(name string) error {
	// Does it exist?
	cm.mut.Lock()
	proc, ok := cm.processes[name]
	if !ok {
		return errors.New("config does not exist")
	}
	cm.mut.Unlock()

	// spawnProcess is responsible for removing the process from the
	// map after it stops so we don't need to delete anything from
	// cm.processses here.
	proc.Stop()
	return nil
}

// Stop stops the ConfigManager and stops all active processes for configs.
func (cm *ConfigManager) Stop() {
	var wg sync.WaitGroup

	cm.mut.Lock()
	wg.Add(len(cm.processes))
	for _, proc := range cm.processes {
		go func(proc configManagerProcess) {
			proc.Stop()
			wg.Done()
		}(proc)
	}
	cm.mut.Unlock()

	wg.Wait()
}

// inst is an interface implemented by Instance, and used by tests
// to isolate agent from instance functionality.
type inst interface {
	Wait() error
	Config() instance.Config
	Stop()
}

type instanceFactory = func(global config.GlobalConfig, cfg instance.Config, walDir string, logger log.Logger) (inst, error)

func defaultInstanceFactory(global config.GlobalConfig, cfg instance.Config, walDir string, logger log.Logger) (inst, error) {
	return instance.New(global, cfg, walDir, logger)
}
