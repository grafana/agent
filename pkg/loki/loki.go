// Package loki implements Loki logs support for the Grafana Cloud Agent.
package loki

import (
	"fmt"
	"os"
	"sync"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/google/go-cmp/cmp"
	"github.com/grafana/agent/pkg/util"
	"github.com/grafana/loki/pkg/promtail"
	"github.com/grafana/loki/pkg/promtail/client"
	"github.com/grafana/loki/pkg/promtail/config"
	"github.com/grafana/loki/pkg/promtail/server"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/version"
)

func init() {
	client.UserAgent = fmt.Sprintf("GrafanaCloudAgent/%s", version.Version)
}

type Loki struct {
	mut sync.Mutex

	reg       prometheus.Registerer
	l         log.Logger
	instances map[string]*Instance
}

// New creates and starts Loki log collection.
func New(reg prometheus.Registerer, c Config, l log.Logger) (*Loki, error) {
	l = log.With(l, "component", "loki")

	loki := &Loki{
		instances: make(map[string]*Instance),
		reg:       reg,
		l:         log.With(l, "component", "loki"),
	}
	if err := loki.ApplyConfig(c); err != nil {
		return nil, err
	}
	return loki, nil
}

// ApplyConfig updates Loki with a new Config.
func (l *Loki) ApplyConfig(c Config) error {
	l.mut.Lock()
	defer l.mut.Unlock()

	if c.PositionsDirectory != "" {
		err := os.MkdirAll(c.PositionsDirectory, 0700)
		if err != nil {
			level.Warn(l.l).Log("msg", "failed to create the positions directory. logs may be unable to save their position", "path", c.PositionsDirectory, "err", err)
		}
	}

	newInstances := make(map[string]*Instance, len(c.Configs))

	for _, ic := range c.Configs {
		// If an old instance existed, update it and move it to the new map.
		if old, ok := l.instances[ic.Name]; ok {
			err := old.ApplyConfig(ic)
			if err != nil {
				return err
			}

			delete(l.instances, ic.Name)
			newInstances[ic.Name] = old
			continue
		}

		inst, err := NewInstance(l.reg, ic, l.l)
		if err != nil {
			return fmt.Errorf("unable to apply config for %s: %w", ic.Name, err)
		}
		newInstances[ic.Name] = inst
	}

	// Any remaining promtail in l.instances has been removed from the new
	// config. Stop them before replacing the map.
	for _, i := range l.instances {
		i.Stop()
	}
	l.instances = newInstances

	return nil
}

func (l *Loki) Stop() {
	l.mut.Lock()
	defer l.mut.Unlock()

	for _, i := range l.instances {
		i.Stop()
	}
}

// Instance is an individual Loki instance.
type Instance struct {
	mut sync.Mutex

	cfg *InstanceConfig
	log log.Logger
	reg *util.Unregisterer

	promtail *promtail.Promtail
}

// NewInstance creates and starts a Loki instance.
func NewInstance(reg prometheus.Registerer, c *InstanceConfig, l log.Logger) (*Instance, error) {
	instReg := prometheus.WrapRegistererWith(prometheus.Labels{"loki_config": c.Name}, reg)

	inst := Instance{
		reg: util.WrapWithUnregisterer(instReg),
		log: log.With(l, "loki_config", c.Name),
	}
	if err := inst.ApplyConfig(c); err != nil {
		return nil, err
	}
	return &inst, nil
}

// ApplyConfig will apply a new InstanceConfig. If the config hasn't changed,
// then nothing will happen, otherwise the old Promtail will be stopped and
// then replaced with a new one.
func (i *Instance) ApplyConfig(c *InstanceConfig) error {
	i.mut.Lock()
	defer i.mut.Unlock()

	// No-op if the configs haven't changed.
	if cmp.Equal(c, i.cfg) {
		level.Debug(i.log).Log("msg", "instance config hasn't changed, not recreating Promtail")
		return nil
	}
	i.cfg = c

	if i.promtail != nil {
		i.promtail.Shutdown()
		i.promtail = nil
	}

	// Unregister all existing metrics before trying to create a new instance.
	if !i.reg.UnregisterAll() {
		// If UnregisterAll fails, we need to abort, otherwise the new promtail
		// would try to re-register an existing metric and might panic.
		return fmt.Errorf("failed to unregister all metrics from previous promtail. THIS IS A BUG!")
	}

	if len(c.ClientConfigs) == 0 {
		level.Debug(i.log).Log("msg", "skipping creation of a promtail because no client_configs are present")
		return nil
	}

	p, err := promtail.New(config.Config{
		ServerConfig:    server.Config{Disable: true},
		ClientConfigs:   c.ClientConfigs,
		PositionsConfig: c.PositionsConfig,
		ScrapeConfig:    c.ScrapeConfig,
		TargetConfig:    c.TargetConfig,
	}, false, promtail.WithLogger(i.log), promtail.WithRegisterer(i.reg))
	if err != nil {
		return fmt.Errorf("unable to create Loki logging instance: %w", err)
	}

	i.promtail = p
	return nil
}

func (i *Instance) Stop() {
	i.mut.Lock()
	defer i.mut.Unlock()

	if i.promtail != nil {
		i.promtail.Shutdown()
		i.promtail = nil
	}
}
