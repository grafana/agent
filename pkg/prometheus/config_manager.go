package prometheus

import (
	"context"
	"errors"
	"sync"

	"github.com/grafana/agent/pkg/prometheus/instance"
)

// ConfigManager manages a set of instance.Configs, calling a function whenever
// a Config should be "started."
type ConfigManager struct {
	// Take care when locking mut: if you hold onto a lock of mut while calling
	// Stop on one of the processes below, you will deadlock.
	mut       sync.Mutex
	processes map[string]configManagerProcess

	newProcess func(ctx context.Context, c instance.Config)
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

// NewConfigManager creates a new ConfigManager. The function f will be invoked
// any time a new instance.Config is tracked. The context provided to the function
// will be cancelled when that instance.Config is no longer being tracked.
//
// f is spawned in a goroutine and is associated with a configManagerProcess.
// It is valid for f to run forever until the provided context is cancelled. Once
// f exits, the config associated with it is automatically removed from the active
// list.
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

	// If the config already exists, we need to "restart" it. We do this by
	// stopping the old process and spawning a new one with the updated config.
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

		// After the process stops, we can remove it from our tracked
		// list. It will then stop showing up in the result of
		// ListConfigs.
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
