// Package prom implements a Prometheus-lite client for service discovery,
// scraping metrics into a WAL, and remote_write. Clients are broken into a
// set of instances, each of which contain their own set of configs.
package prom

import (
	"errors"
	"flag"
	"fmt"
	"sync"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/google/go-cmp/cmp"
	"github.com/grafana/agent/pkg/prom/ha"
	"github.com/grafana/agent/pkg/prom/ha/client"
	"github.com/grafana/agent/pkg/prom/instance"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/prometheus/config"
	"google.golang.org/grpc"
)

var (
	DefaultConfig = Config{
		Global:                 config.DefaultGlobalConfig,
		InstanceRestartBackoff: instance.DefaultBasicManagerConfig.InstanceRestartBackoff,
		WALCleanupAge:          DefaultCleanupAge,
		WALCleanupPeriod:       DefaultCleanupPeriod,
		ServiceConfig:          ha.DefaultConfig,
		ServiceClientConfig:    client.DefaultConfig,
		InstanceMode:           instance.DefaultMode,
	}
)

// Config defines the configuration for the entire set of Prometheus client
// instances, along with a global configuration.
type Config struct {
	// Whether the Prometheus subsystem should be enabled.
	Enabled bool `yaml:"-"`

	Global                 config.GlobalConfig `yaml:"global"`
	WALDir                 string              `yaml:"wal_directory"`
	WALCleanupAge          time.Duration       `yaml:"wal_cleanup_age"`
	WALCleanupPeriod       time.Duration       `yaml:"wal_cleanup_period"`
	ServiceConfig          ha.Config           `yaml:"scraping_service"`
	ServiceClientConfig    client.Config       `yaml:"scraping_service_client"`
	Configs                []instance.Config   `yaml:"configs,omitempty"`
	InstanceRestartBackoff time.Duration       `yaml:"instance_restart_backoff,omitempty"`
	InstanceMode           instance.Mode       `yaml:"instance_mode"`
}

// UnmarshalYAML implements yaml.Unmarshaler.
func (c *Config) UnmarshalYAML(unmarshal func(interface{}) error) error {
	*c = DefaultConfig

	// If the Config is unmarshaled, it's present in the config and should be
	// enabled.
	c.Enabled = true

	type plain Config
	return unmarshal((*plain)(c))
}

// ApplyDefaults applies default values to the Config and validates it.
func (c *Config) ApplyDefaults() error {
	if c.WALDir == "" {
		return errors.New("no wal_directory configured")
	}

	if c.ServiceConfig.Enabled && len(c.Configs) > 0 {
		return errors.New("cannot use configs when scraping_service mode is enabled")
	}

	usedNames := map[string]struct{}{}

	for i := range c.Configs {
		name := c.Configs[i].Name
		if err := c.Configs[i].ApplyDefaults(&c.Global); err != nil {
			// Try to show a helpful name in the error
			if name == "" {
				name = fmt.Sprintf("at index %d", i)
			}

			return fmt.Errorf("error validating instance %s: %w", name, err)
		}

		if _, ok := usedNames[name]; ok {
			return fmt.Errorf(
				"prometheus instance names must be unique. found multiple instances with name %s",
				name,
			)
		}
		usedNames[name] = struct{}{}
	}
	return nil
}

// RegisterFlags defines flags corresponding to the Config.
func (c *Config) RegisterFlags(f *flag.FlagSet) {
	f.StringVar(&c.WALDir, "prometheus.wal-directory", "", "base directory to store the WAL in")
	f.DurationVar(&c.WALCleanupAge, "prometheus.wal-cleanup-age", DefaultConfig.WALCleanupAge, "remove abandoned (unused) WALs older than this")
	f.DurationVar(&c.WALCleanupPeriod, "prometheus.wal-cleanup-period", DefaultConfig.WALCleanupPeriod, "how often to check for abandoned WALs")
	f.DurationVar(&c.InstanceRestartBackoff, "prometheus.instance-restart-backoff", DefaultConfig.InstanceRestartBackoff, "how long to wait before restarting a failed Prometheus instance")

	c.ServiceConfig.RegisterFlagsWithPrefix("prometheus.service.", f)
	c.ServiceClientConfig.RegisterFlags(f)
}

// Agent is an agent for collecting Prometheus metrics. It acts as a
// Prometheus-lite; only running the service discovery, remote_write, and WAL
// components of Prometheus. It is broken down into a series of Instances, each
// of which perform metric collection.
type Agent struct {
	mut    sync.Mutex
	cfg    Config
	logger log.Logger
	reg    prometheus.Registerer

	// Store both the basic manager and the modal manager so we can update their
	// settings indepedently. Only the ModalManager should be used for mutating
	// configs.
	bm      *instance.BasicManager
	mm      *instance.ModalManager
	cleaner *WALCleaner

	instanceFactory instanceFactory

	ha    *ha.Server
	haAPI *ha.API
}

// New creates and starts a new Agent.
func New(reg prometheus.Registerer, cfg Config, logger log.Logger) (*Agent, error) {
	return newAgent(reg, cfg, logger, defaultInstanceFactory)
}

func newAgent(reg prometheus.Registerer, cfg Config, logger log.Logger, fact instanceFactory) (*Agent, error) {
	a := &Agent{
		logger:          log.With(logger, "agent", "prometheus"),
		instanceFactory: fact,
		reg:             reg,
	}

	a.haAPI = ha.NewAPI(a.logger, nil)

	a.bm = instance.NewBasicManager(instance.BasicManagerConfig{
		InstanceRestartBackoff: cfg.InstanceRestartBackoff,
	}, a.logger, a.newInstance, a.validateInstance)

	var err error
	a.mm, err = instance.NewModalManager(a.reg, a.logger, a.bm, cfg.InstanceMode)
	if err != nil {
		return nil, fmt.Errorf("failed to create modal instance manager: %w", err)
	}

	if err := a.ApplyConfig(cfg); err != nil {
		return nil, fmt.Errorf("failed to apply config: %w", err)
	}
	return a, nil
}

// ApplyConfig will mutate the state of the Agent to match the new Config.
func (a *Agent) ApplyConfig(c Config) error {
	a.mut.Lock()
	defer a.mut.Unlock()

	if cmp.Equal(c, a.cfg) {
		// No config change
		return nil
	}

	// Update instance managers with new settings. If the InstanceMode changed,
	// the apply will be slower - all instances need to be recreated.
	a.bm.UpdateManagerConfig(instance.BasicManagerConfig{
		InstanceRestartBackoff: c.InstanceRestartBackoff,
	})
	if err := a.mm.SetMode(c.InstanceMode); err != nil {
		return fmt.Errorf("failed to update instance mode: %w", err)
	}

	//
	// Apply new instances
	//
	var (
		newConfigs      = make(map[string]struct{}, len(c.Configs))
		allConfigsValid = true
	)
	for _, c := range c.Configs {
		if err := a.mm.ApplyConfig(c); err != nil {
			level.Error(a.logger).Log("msg", "failed to apply config", "name", c.Name, "err", err)
			allConfigsValid = false
		}
		newConfigs[c.Name] = struct{}{}
	}
	if !allConfigsValid {
		return fmt.Errorf("one or more configs was found to be invalid")
	}

	// Iterate over the old configs and delete them if they're not in the
	// newConfigs map
	for _, oldConfig := range a.cfg.Configs {
		if _, found := newConfigs[oldConfig.Name]; found {
			continue
		}
		if err := a.mm.DeleteConfig(oldConfig.Name); err != nil {
			return fmt.Errorf("failed to remove deleted config %s: %w", oldConfig.Name, err)
		}
	}

	//
	// Update HA mode to match new state. If it's disabled and currently running,
	// stop it.
	//
	if haNeedsUpdate(a.cfg, c) {
		if a.ha != nil {
			if err := a.ha.Stop(); err != nil {
				return fmt.Errorf("failed to stop scraping service for config update: %w", err)
			}
			a.ha = nil
		}

		if c.ServiceConfig.Enabled {
			var err error
			a.ha, err = ha.New(a.reg, c.ServiceConfig, &c.Global, c.ServiceClientConfig, a.logger, a.mm)
			if err != nil {
				return fmt.Errorf("failed to start scraping service: %w", err)
			}
		}

		a.haAPI.SetServer(a.ha)
	}

	//
	// Update WAL cleaner
	//
	if cleanerNeedsUpdate(a.cfg, c) {
		if a.cleaner != nil {
			a.cleaner.Stop()
		}
		a.cleaner = NewWALCleaner(
			a.logger,
			a.mm,
			c.WALDir,
			c.WALCleanupAge,
			c.WALCleanupPeriod,
		)
	}

	a.cfg = c
	return nil
}

func haNeedsUpdate(oldCfg, newCfg Config) bool {
	compare := []struct{ prev, next interface{} }{
		{prev: oldCfg.ServiceConfig, next: newCfg.ServiceConfig},
		{prev: oldCfg.Global, next: newCfg.Global},
		{prev: oldCfg.ServiceClientConfig, next: newCfg.ServiceClientConfig},
	}
	for _, c := range compare {
		if !cmp.Equal(c.prev, c.next) {
			return true
		}
	}
	return false
}

func cleanerNeedsUpdate(oldCfg, newCfg Config) bool {
	compare := []struct{ prev, next interface{} }{
		{prev: oldCfg.WALDir, next: newCfg.WALDir},
		{prev: oldCfg.WALCleanupAge, next: newCfg.WALCleanupAge},
		{prev: oldCfg.WALCleanupPeriod, next: newCfg.WALCleanupPeriod},
	}
	for _, c := range compare {
		if !cmp.Equal(c.prev, c.next) {
			return true
		}
	}
	return false
}

// newInstance creates a new Instance given a config.
func (a *Agent) newInstance(c instance.Config) (instance.ManagedInstance, error) {
	// Controls the label
	instanceLabel := "instance_name"
	if a.cfg.InstanceMode == instance.ModeShared {
		instanceLabel = "instance_group_name"
	}

	reg := prometheus.WrapRegistererWith(prometheus.Labels{
		instanceLabel: c.Name,
	}, a.reg)

	return a.instanceFactory(reg, a.cfg.Global, c, a.cfg.WALDir, a.logger)
}

func (a *Agent) validateInstance(c *instance.Config) error {
	return c.ApplyDefaults(&a.cfg.Global)
}

func (a *Agent) WireGRPC(s *grpc.Server) {
	a.haAPI.WireGRPC(s)
}

func (a *Agent) Config() Config {
	a.mut.Lock()
	defer a.mut.Unlock()

	return a.cfg
}

func (a *Agent) InstanceManager() instance.Manager {
	return a.mm
}

// Stop stops the agent and all its instances.
func (a *Agent) Stop() {
	if a.ha != nil {
		if err := a.ha.Stop(); err != nil {
			level.Error(a.logger).Log("msg", "failed to stop scraping service server", "err", err)
		}
	}
	a.cleaner.Stop()

	// Only need to stop the ModalManager, which will passthrough everything to the
	// BasicManager.
	a.mm.Stop()
}

type instanceFactory = func(reg prometheus.Registerer, global config.GlobalConfig, cfg instance.Config, walDir string, logger log.Logger) (instance.ManagedInstance, error)

func defaultInstanceFactory(reg prometheus.Registerer, global config.GlobalConfig, cfg instance.Config, walDir string, logger log.Logger) (instance.ManagedInstance, error) {
	return instance.New(reg, global, cfg, walDir, logger)
}
