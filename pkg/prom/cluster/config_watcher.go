package cluster

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/google/go-cmp/cmp"
	"github.com/grafana/agent/pkg/prom/instance"
	"github.com/grafana/agent/pkg/prom/instance/configstore"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	reshardDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name: "agent_prometheus_scraping_service_reshard_duration",
		Help: "How long it took for resharding to run.",
	}, []string{"success"})
)

// configWatcher connects to a configstore and will apply configs to an
// instance.Manager.
type configWatcher struct {
	log log.Logger

	cfgMut  sync.Mutex
	cfg     Config
	stopped bool
	stop    context.CancelFunc

	store    configstore.Store
	im       instance.Manager
	owns     OwnershipFunc
	validate ValidationFunc

	instanceMut sync.Mutex
	instances   map[string]struct{}
}

// OwnershipFunc should determine if a given keep is owned by the caller.
type OwnershipFunc = func(key string) (bool, error)

// ValidationFunc should validate a config.
type ValidationFunc = func(*instance.Config) error

// newConfigWatcher watches store for changes and checks for each config against
// owns. It will also poll the configstore at a configurable interval.
func newConfigWatcher(log log.Logger, cfg Config, store configstore.Store, im instance.Manager, owns OwnershipFunc, validate ValidationFunc) (*configWatcher, error) {
	ctx, cancel := context.WithCancel(context.Background())

	w := &configWatcher{
		log: log,

		stop: cancel,

		store:    store,
		im:       im,
		owns:     owns,
		validate: validate,

		instances: make(map[string]struct{}),
	}
	if err := w.ApplyConfig(cfg); err != nil {
		return nil, err
	}
	go w.run(ctx)
	return w, nil
}

func (w *configWatcher) ApplyConfig(cfg Config) error {
	w.cfgMut.Lock()
	defer w.cfgMut.Unlock()

	if cmp.Equal(w.cfg, cfg) {
		return nil
	}

	if w.stopped {
		return fmt.Errorf("configWatcher already stopped")
	}

	w.cfg = cfg
	return nil
}

func (w *configWatcher) run(ctx context.Context) {
	for {
		w.cfgMut.Lock()
		nextPoll := w.cfg.ReshardInterval
		w.cfgMut.Unlock()

		select {
		case <-ctx.Done():
			return
		case <-time.After(nextPoll):
			// Errors are logged by w.Refresh, ignore the error here.
			_ = w.Refresh(ctx)
		case ev := <-w.store.Watch():
			if err := w.handleEvent(ev); err != nil {
				level.Error(w.log).Log("msg", "failed to handle changend or deleted config", "key", ev.Key, "err", err)
			}
		}
	}
}

// Refresh reloads all configs from the configstore. Deleted configs will be
// removed.
func (w *configWatcher) Refresh(ctx context.Context) (err error) {
	start := time.Now()
	defer func() {
		success := "1"
		if err != nil {
			success = "0"
		}
		reshardDuration.WithLabelValues(success).Observe(time.Since(start).Seconds())
	}()

	configs, err := w.store.All(ctx, func(key string) bool {
		owns, err := w.owns(key)
		if err != nil {
			level.Error(w.log).Log("msg", "failed to check for ownership, instance will be deleted if it is running", "key", key, "err", err)
			return false
		}
		return owns
	})
	if err != nil {
		return fmt.Errorf("failed to get configs from store: %w", err)
	}

	var (
		keys       = make(map[string]struct{})
		firstError error
	)

Outer:
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case cfg, ok := <-configs:
			// w.store.All will close configs when all of them have been read.
			if !ok {
				break Outer
			}

			if err := w.handleEvent(configstore.WatchEvent{Key: cfg.Name, Config: &cfg}); err != nil {
				level.Error(w.log).Log("msg", "failed to process changed config", "key", cfg.Name, "err", err)
				if firstError == nil {
					firstError = err
				}
			}

			keys[cfg.Name] = struct{}{}
		}
	}

	// Any config we used to be running that disappeared from this most recent
	// iteration should be deleted. We hold the lock just for the duration of
	// populating deleted because handleEvent also grabs a hold on the lock.
	var deleted []string
	w.instanceMut.Lock()
	for key := range w.instances {
		if _, exist := keys[key]; exist {
			continue
		}
		deleted = append(deleted, key)
	}
	w.instanceMut.Unlock()

	// Send a deleted event for any key that has gone away.
	for _, key := range deleted {
		if err := w.handleEvent(configstore.WatchEvent{Key: key, Config: nil}); err != nil {
			level.Error(w.log).Log("msg", "failed to process changed config", "key", key, "err", err)
		}
	}

	return firstError
}

func (w *configWatcher) handleEvent(ev configstore.WatchEvent) error {
	w.instanceMut.Lock()
	defer w.instanceMut.Unlock()

	owned, err := w.owns(ev.Key)
	if err != nil {
		level.Error(w.log).Log("msg", "failed to see if config is owned. instance will be deleted if it is running", "err", err)
	}

	var (
		_, isRunning = w.instances[ev.Key]
		isDeleted    = ev.Config == nil
	)

	switch {
	// Two deletion scenarios:
	// 1. A config we're running got moved to a new owner.
	// 2. A config we're running got deleted
	case (isRunning && !owned) || (isDeleted && isRunning):
		err := w.im.DeleteConfig(ev.Key)
		delete(w.instances, ev.Key)
		if err != nil {
			return fmt.Errorf("failed to delete: %w", err)
		}

	case !isDeleted && owned:
		if err := w.validate(ev.Config); err != nil {
			return fmt.Errorf(
				"failed to validate config. %[1]s cannot run until the global settings are adjusted or the config is adjusted to operate within the global constraints. error: %[2]w",
				ev.Key, err,
			)
		}

		if err := w.im.ApplyConfig(*ev.Config); err != nil {
			return fmt.Errorf("failed to apply config: %w", err)
		}
		w.instances[ev.Key] = struct{}{}
	}

	return nil
}

// Stop stops the configWatcher. Cannot be called more than once.
func (w *configWatcher) Stop() error {
	w.cfgMut.Lock()
	defer w.cfgMut.Unlock()

	if w.stopped {
		return fmt.Errorf("already stopped")
	}
	w.stopped = true

	w.stop()

	w.instanceMut.Lock()
	defer w.instanceMut.Unlock()

	for key := range w.instances {
		if err := w.im.DeleteConfig(key); err != nil {
			level.Warn(w.log).Log("msg", "failed deleting config on shutdown", "key", key, "err", err)
		}
	}
	w.instances = make(map[string]struct{})

	return nil
}
