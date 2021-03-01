// Package prom implements a Prometheus-lite client for service discovery,
// scraping metrics into a WAL, and remote_write. Clients are broken into a
// set of instances, each of which contain their own set of configs.
package prom

import (
	"errors"
	"flag"
	"fmt"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
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

	Global                 config.GlobalConfig           `yaml:"global"`
	WALDir                 string                        `yaml:"wal_directory"`
	WALCleanupAge          time.Duration                 `yaml:"wal_cleanup_age"`
	WALCleanupPeriod       time.Duration                 `yaml:"wal_cleanup_period"`
	ServiceConfig          ha.Config                     `yaml:"scraping_service"`
	ServiceClientConfig    client.Config                 `yaml:"scraping_service_client"`
	Configs                []instance.Config             `yaml:"configs,omitempty"`
	InstanceRestartBackoff time.Duration                 `yaml:"instance_restart_backoff,omitempty"`
	InstanceMode           instance.Mode                 `yaml:"instance_mode"`
	RemoteWrite            []*instance.RemoteWriteConfig `yaml:"remote_write,omitempty"`
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
		if err := c.Configs[i].ApplyDefaults(&c.Global, c.RemoteWrite); err != nil {
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

	ha *ha.Server
}

// New creates and starts a new Agent.
func New(reg prometheus.Registerer, cfg Config, logger log.Logger) (*Agent, error) {
	return newAgent(reg, cfg, logger, defaultInstanceFactory)
}

func newAgent(reg prometheus.Registerer, cfg Config, logger log.Logger, fact instanceFactory) (*Agent, error) {
	a := &Agent{
		cfg:             cfg,
		logger:          log.With(logger, "agent", "prometheus"),
		instanceFactory: fact,
		reg:             reg,
	}

	a.bm = instance.NewBasicManager(instance.BasicManagerConfig{
		InstanceRestartBackoff: cfg.InstanceRestartBackoff,
	}, a.logger, a.newInstance, a.validateInstance)

	var err error
	a.mm, err = instance.NewModalManager(a.reg, a.logger, a.bm, cfg.InstanceMode)
	if err != nil {
		return nil, fmt.Errorf("failed to create modal instance manager: %w", err)
	}

	// Periodically attempt to clean up WALs from instances that aren't being run by
	// this agent anymore.
	a.cleaner = NewWALCleaner(
		a.logger,
		a.mm,
		cfg.WALDir,
		cfg.WALCleanupAge,
		cfg.WALCleanupPeriod,
	)

	allConfigsValid := true
	for _, c := range cfg.Configs {
		if err := a.mm.ApplyConfig(c); err != nil {
			level.Error(logger).Log("msg", "failed to apply config", "name", c.Name, "err", err)
			allConfigsValid = false
		}
	}
	if !allConfigsValid {
		return nil, fmt.Errorf("one or more configs was found to be invalid")
	}

	if cfg.ServiceConfig.Enabled {
		var err error
		a.ha, err = ha.New(reg, cfg.ServiceConfig, &cfg.Global, cfg.ServiceClientConfig, a.logger, a.mm, cfg.RemoteWrite)
		if err != nil {
			return nil, err
		}
	}

	return a, nil
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
	return c.ApplyDefaults(&a.cfg.Global, c.RemoteWrite)
}

func (a *Agent) WireGRPC(s *grpc.Server) {
	if a.cfg.ServiceConfig.Enabled {
		a.ha.WireGRPC(s)
	}
}

func (a *Agent) Config() Config                    { return a.cfg }
func (a *Agent) InstanceManager() instance.Manager { return a.mm }

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
