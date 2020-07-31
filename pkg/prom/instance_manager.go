package prom

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/grafana/agent/pkg/prom/instance"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/prometheus/scrape"
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

	// DefaultInstanceManagerConfig is the default config for instance managers,
	// derived from the default Agent config.
	DefaultInstanceManagerConfig = InstanceManagerConfig{
		InstanceRestartBackoff: DefaultConfig.InstanceRestartBackoff,
	}
)

type InstanceManagerConfig struct {
	InstanceRestartBackoff time.Duration
}

// InstanceManager manages a set of Instances, calling a factory function to
// create a new Instance whenever one should be created. Instances will be
// kept running by the InstanceManager and restarted if they are stopped.
type InstanceManager struct {
	cfg    InstanceManagerConfig
	logger log.Logger

	// Take care when locking mut: if you hold onto a lock of mut while calling
	// Stop on one of the processes below, you will deadlock.
	mut       sync.Mutex
	processes map[string]*instanceManagerProcess

	launch   InstanceFactory
	validate ConfigValidator
}

// instanceManagerProcess represents a goroutine managing an instance.Config. In
// practice, this will be an *instance.Instance. cancel requests that the goroutine
// should shut down. done will be closed after the goroutine exits.
type instanceManagerProcess struct {
	cfg    instance.Config
	inst   Instance
	cancel context.CancelFunc
	done   chan bool
}

// Stop stops the process and waits for it to exit.
func (p instanceManagerProcess) Stop() {
	p.cancel()
	<-p.done
}

// NewInstanceManager creates a new InstanceManager. The function f will be invoked
// any time a new instance.Config is tracked. The context provided to the function
// will be cancelled when that instance.Config is no longer being tracked.
//
// The InstanceLauncher will be called in a goroutine and is expected to run forever
// until the associate instance stops or the context is canceled.
//
// The ConfigValidator will be called before launching an instance. If the config
// is not valid, the config will not be launched.
func NewInstanceManager(cfg InstanceManagerConfig, logger log.Logger, launch InstanceFactory, validate ConfigValidator) *InstanceManager {
	return &InstanceManager{
		cfg:       cfg,
		logger:    logger,
		processes: make(map[string]*instanceManagerProcess),
		launch:    launch,
		validate:  validate,
	}
}

// An InstanceFactory should return an unstarted instance given some config.
type InstanceFactory func(c instance.Config) (Instance, error)

// A ConfigValidator should validate an instance.Config and return an error if
// a problem was found.
type ConfigValidator func(c *instance.Config) error

// Instance represents a running process that performance Prometheus
// functionality. It is implemented by instance.Instance but is defined as an
// interface here for the sake of testing.
type Instance interface {
	Run(ctx context.Context) error
	TargetsActive() map[string][]*scrape.Target
}

// ListInstances returns the current active instances managed by the InstanceManager.
func (im *InstanceManager) ListInstances() map[string]Instance {
	im.mut.Lock()
	defer im.mut.Unlock()

	insts := make(map[string]Instance, len(im.processes))
	for name, process := range im.processes {
		insts[name] = process.inst
	}
	return insts
}

// ListConfigs lists the current active configs managed by the InstanceManager.
func (im *InstanceManager) ListConfigs() map[string]instance.Config {
	im.mut.Lock()
	defer im.mut.Unlock()

	cfgs := make(map[string]instance.Config, len(im.processes))
	for name, process := range im.processes {
		cfgs[name] = process.cfg
	}
	return cfgs
}

// ApplyConfig takes an instance.Config and either adds a new tracked config
// or updates an existing track config. The value for Name in c is used to
// uniquely identify the instance.Config and determine whether it is new
// or existing.
func (im *InstanceManager) ApplyConfig(c instance.Config) error {
	if im.validate != nil {
		err := im.validate(&c)
		if err != nil {
			return fmt.Errorf("failed to validate instance %s: %w", c.Name, err)
		}
	}

	im.mut.Lock()
	defer im.mut.Unlock()

	// If the config already exists, we need to "restart" it. We do this by
	// stopping the old process and spawning a new one with the updated config.
	if proc, ok := im.processes[c.Name]; ok {
		proc.Stop()
	}

	// Spawn a new process for the new config.
	err := im.spawnProcess(c)
	if err != nil {
		return err
	}
	currentActiveConfigs.Inc()
	return nil
}

func (im *InstanceManager) spawnProcess(c instance.Config) error {
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan bool)

	inst, err := im.launch(c)
	if err != nil {
		return err
	}

	proc := &instanceManagerProcess{
		cancel: cancel,
		done:   done,
		cfg:    c,
		inst:   inst,
	}

	im.processes[c.Name] = proc

	go func() {
		im.runProcess(ctx, c.Name, proc.inst)
		close(done)

		im.mut.Lock()
		// Now that the process is stopped, we can remove it from our tracked
		// list. It will stop showing up in the result of ListConfigs.
		//
		// However, it's possible that a new config has been applied and overwrote
		// the initial value in our map. We should only delete the process from
		// the map if it hasn't changed from what we initially set it to.
		if storedProc, exist := im.processes[c.Name]; exist && storedProc == proc {
			delete(im.processes, c.Name)
		}
		im.mut.Unlock()

		currentActiveConfigs.Dec()
	}()

	return nil
}

// runProcess runs an instance and keeps it alive until the context is canceled.
func (im *InstanceManager) runProcess(ctx context.Context, name string, inst Instance) {
	for {
		err := inst.Run(ctx)
		if err != nil && err != context.Canceled {
			instanceAbnormalExits.WithLabelValues(name).Inc()
			level.Error(im.logger).Log("msg", "instance stopped abnormally, restarting after backoff period", "err", err, "backoff", im.cfg.InstanceRestartBackoff, "instance", name)
			time.Sleep(im.cfg.InstanceRestartBackoff)
		} else {
			level.Info(im.logger).Log("msg", "stopped instance", "instance", name)
			break
		}
	}
}

// DeleteConfig removes an instance.Config by its name. Returns an error if
// the instance.Config is not currently being tracked.
func (im *InstanceManager) DeleteConfig(name string) error {
	im.mut.Lock()
	proc, ok := im.processes[name]
	if !ok {
		return errors.New("config does not exist")
	}
	im.mut.Unlock()

	// spawnProcess is responsible for removing the process from the
	// map after it stops so we don't need to delete anything from
	// im.processes here.
	proc.Stop()
	return nil
}

// Stop stops the InstanceManager and stops all active processes for configs.
func (im *InstanceManager) Stop() {
	var wg sync.WaitGroup

	im.mut.Lock()
	wg.Add(len(im.processes))
	for _, proc := range im.processes {
		go func(proc *instanceManagerProcess) {
			proc.Stop()
			wg.Done()
		}(proc)
	}
	im.mut.Unlock()

	wg.Wait()
}
