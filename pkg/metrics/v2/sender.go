package metrics

import (
	"fmt"
	"path/filepath"
	"sync"
	"time"

	"github.com/go-kit/log"
	"github.com/grafana/agent/pkg/metrics/wal"
	"github.com/grafana/agent/pkg/util"
	"github.com/prometheus/client_golang/prometheus"
	prom_config "github.com/prometheus/prometheus/config"
	"github.com/prometheus/prometheus/scrape"
	"github.com/prometheus/prometheus/storage"
	"github.com/prometheus/prometheus/storage/remote"
	"github.com/prometheus/statsd_exporter/pkg/level"
)

// storageSet abstracts over a series of WALs by name. Returns an error if the
// provided name isn't known.
type storageSet interface {
	Appendable(name string) (storage.Appendable, error)
}

// senderManager manages a set of senders. It implements storageSet.
type senderManager struct {
	log log.Logger
	reg prometheus.Registerer
	o   Options

	mut             sync.RWMutex
	senderInstances map[string]*sender
}

// newSenderManager creates a new senderManager. No senders are configured until ApplyConfig is called.
func newSenderManager(l log.Logger, reg prometheus.Registerer, o Options) *senderManager {
	return &senderManager{
		log: l,
		reg: reg,
		o:   o,

		senderInstances: make(map[string]*sender),
	}
}

// Appendable returns an appenable storage by instance name. Returns an error
// if the named storage doesn't exist.
func (sm *senderManager) Appendable(name string) (storage.Appendable, error) {
	sm.mut.RLock()
	defer sm.mut.RUnlock()

	sender, ok := sm.senderInstances[name]
	if !ok {
		return nil, fmt.Errorf("sender %q not found", name)
	}
	return sender.wal, nil
}

// ApplyConfig synchronizes the senders with cfg.
func (sm *senderManager) ApplyConfig(cfg *Config) error {
	sm.mut.Lock()
	defer sm.mut.Unlock()

	var firstError error
	saveError := func(e error) {
		if firstError == nil {
			firstError = e
		}
	}

	currentConfigs := make(map[string]struct{}, len(cfg.Configs))

	for _, ic := range cfg.Configs {
		currentConfigs[ic.Name] = struct{}{}

		sender, ok := sm.senderInstances[ic.Name]
		if !ok {
			l := log.With(sm.log, "component", "metrics.sender", "instance", ic.Name)
			reg := prometheus.WrapRegistererWith(prometheus.Labels{"instance": ic.Name}, sm.reg)

			var err error
			sender, err = newSender(l, reg, filepath.Join(sm.o.WALDir, ic.Name), sm.o.RemoteFlushDeadline)
			if err != nil {
				level.Error(sm.log).Log("msg", "failed creating a sender for metrics instance", "instance", ic.Name, "err", err)
				saveError(err)
				continue
			}
			sm.senderInstances[ic.Name] = sender
		}

		if err := sender.ApplyConfig(cfg.Global.Prometheus, ic.RemoteWrite); err != nil {
			level.Error(sm.log).Log("msg", "failed apply remote_write configs for metrics instance", "instance", ic.Name, "err", err)
			saveError(err)
		}
	}

	// Remove any senders that have gone away between reloads.
	for instance, sender := range sm.senderInstances {
		_, exist := currentConfigs[instance]
		if !exist {
			level.Info(sm.log).Log("msg", "shutting down stale metrics sender", "instance", instance)
			if err := sender.Close(); err != nil {
				level.Error(sm.log).Log("msg", "failed to shut down stale metrics instance", "instance", instance)
				saveError(err)
			}
			delete(sm.senderInstances, instance)
		}
	}

	return firstError
}

// Stop stops the senderManager and all runnning senders.
func (sm *senderManager) Stop() error {
	sm.mut.Lock()
	defer sm.mut.Unlock()

	var firstError error
	saveError := func(e error) {
		if firstError == nil {
			firstError = e
		}
	}

	for inst, sender := range sm.senderInstances {
		if err := sender.Close(); err != nil {
			level.Warn(sm.log).Log("msg", "failed when shutting down metrics sender", "instance", inst, "err", err)
			saveError(err)
		}
	}

	return firstError
}

// sender combines an individual WAL and remote_write.
type sender struct {
	log log.Logger

	reg *util.Unregisterer

	wal     *wal.Storage
	rs      *remote.Storage
	storage storage.Storage
}

// newSender creates a new sender. ApplyConfig must be invoked to configure
// locations to send metrics to. After ApplyConfig is called, metrics written
// to dir will be synced over remote_write in the background.
func newSender(l log.Logger, reg prometheus.Registerer, dir string, flushDeadline time.Duration) (*sender, error) {
	ureg := util.WrapWithUnregisterer(reg)

	w, err := wal.NewStorage(l, ureg, dir)
	if err != nil {
		return nil, err
	}
	rs := remote.NewStorage(l, ureg, w.StartTime, w.Directory(), flushDeadline, noOpScrapeManager{})
	storage := storage.NewFanout(l, w, rs)

	return &sender{
		log:     l,
		wal:     w,
		rs:      rs,
		storage: storage,
		reg:     ureg,
	}, nil
}

// ApplyConfig updates the set of remote endpoints which are receiving metrics.
func (s *sender) ApplyConfig(global prom_config.GlobalConfig, rw []*prom_config.RemoteWriteConfig) error {
	return s.rs.ApplyConfig(&prom_config.Config{
		GlobalConfig:       global,
		RemoteWriteConfigs: rw,
	})
}

// Close stops the sender. Any registered metrics will be unregistered.
func (s *sender) Close() error {
	_ = s.reg.UnregisterAll()
	return s.storage.Close()
}

type noOpScrapeManager struct{}

func (noOpScrapeManager) Get() (*scrape.Manager, error) { return nil, nil }
