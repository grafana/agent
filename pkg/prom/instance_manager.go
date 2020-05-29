package prom

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/grafana/agent/pkg/prom/instance"
)

// InstanceManager manages a set of instance.Configs, calling a function whenever
// a Config should be "started." It is detacted from the concept of actual instances
// to allow for mocking. The New function in this package creates an Agent that
// utilizes InstanceManager for actually launching real instances.
type InstanceManager struct {
	// Take care when locking mut: if you hold onto a lock of mut while calling
	// Stop on one of the processes below, you will deadlock.
	mut       sync.Mutex
	processes map[string]*configManagerProcess

	launch   InstanceLauncher
	validate ConfigValidator
}

// configManagerProcess represents a goroutine managing an instance.Config. In
// practice, this will be an *instance.Instance. cancel requests that the goroutine
// should shut down. done will be closed after the goroutine exits.
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

// NewInstanceManager creates a new InstanceManager. The function f will be invoked
// any time a new instance.Config is tracked. The context provided to the function
// will be cancelled when that instance.Config is no longer being tracked.
//
// The InstanceLauncher will be called in a goroutine and is expected to run forever
// until the associate instance stops or the context is canceled.
//
// The ConfigValidator will be called before launching an instance. If the config
// is not valid, the config will not be launched.
func NewInstanceManager(launch InstanceLauncher, validate ConfigValidator) *InstanceManager {
	return &InstanceManager{
		processes: make(map[string]*configManagerProcess),
		launch:    launch,
		validate:  validate,
	}
}

type InstanceLauncher func(ctx context.Context, c instance.Config)
type ConfigValidator func(c *instance.Config) error

// ListConfigs lists the current active configs managed by the InstanceManager.
func (cm *InstanceManager) ListConfigs() map[string]instance.Config {
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
func (cm *InstanceManager) ApplyConfig(c instance.Config) error {
	if cm.validate != nil {
		err := cm.validate(&c)
		if err != nil {
			return fmt.Errorf("failed to validate instance %s: %w", c.Name, err)
		}
	}

	cm.mut.Lock()
	defer cm.mut.Unlock()

	// If the config already exists, we need to "restart" it. We do this by
	// stopping the old process and spawning a new one with the updated config.
	if proc, ok := cm.processes[c.Name]; ok {
		proc.Stop()
	}

	// Spawn a new process for the new config.
	cm.spawnProcess(c)
	currentActiveConfigs.Inc()
	return nil
}

func (cm *InstanceManager) spawnProcess(c instance.Config) {
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan bool)

	proc := &configManagerProcess{
		cancel: cancel,
		done:   done,
	}

	cm.processes[c.Name] = proc

	go func() {
		cm.launch(ctx, c)
		close(done)

		cm.mut.Lock()
		// Now that the process is stopped, we can remove it from our tracked
		// list. It will stop showing up in the result of ListConfigs.
		//
		// However, it's possible that a new config has been applied and overwrote
		// the initial value in our map. We should only delete the process from
		// the map if it hasn't changed from what we initially set it to.
		if storedProc, exist := cm.processes[c.Name]; exist && storedProc == proc {
			delete(cm.processes, c.Name)
		}
		cm.mut.Unlock()

		currentActiveConfigs.Dec()
	}()
}

// DeleteConfig removes an instance.Config by its name. Returns an error if
// the instance.Config is not currently being tracked.
func (cm *InstanceManager) DeleteConfig(name string) error {
	cm.mut.Lock()
	proc, ok := cm.processes[name]
	if !ok {
		return errors.New("config does not exist")
	}
	cm.mut.Unlock()

	// spawnProcess is responsible for removing the process from the
	// map after it stops so we don't need to delete anything from
	// cm.processes here.
	proc.Stop()
	return nil
}

// Stop stops the InstanceManager and stops all active processes for configs.
func (cm *InstanceManager) Stop() {
	var wg sync.WaitGroup

	cm.mut.Lock()
	wg.Add(len(cm.processes))
	for _, proc := range cm.processes {
		go func(proc *configManagerProcess) {
			proc.Stop()
			wg.Done()
		}(proc)
	}
	cm.mut.Unlock()

	wg.Wait()
}
