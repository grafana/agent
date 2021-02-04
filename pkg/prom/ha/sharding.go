package ha

import (
	"context"
	"fmt"
	"hash/fnv"
	"net/http"
	"sync"
	"time"

	"github.com/cortexproject/cortex/pkg/ring"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/grafana/agent/pkg/agentproto"
	"github.com/grafana/agent/pkg/prom/instance"
)

// Reshard initiates an entire reshard of the current HA scraping service instance.
// All configs will be reloaded from the KV store and the scraping service instance
// will see what should be managed locally.
//
// Satisfies agentproto.ScrapingServiceServer.
func (s *Server) Reshard(ctx context.Context, _ *agentproto.ReshardRequest) (_ *empty.Empty, err error) {
	s.configManagerMut.Lock()
	defer s.configManagerMut.Unlock()

	start := time.Now()
	defer func() {
		success := "1"
		if err != nil {
			success = "0"
		}
		reshardDuration.WithLabelValues(success).Observe(time.Since(start).Seconds())
	}()

	var (
		// current configs the agent is tracking
		currentConfigs = s.im.ListConfigs()

		// configs found in the KV store. currentConfigs - discoveredConfigs is the
		// list of configs that was removed from the KV store since the last reshard.
		discoveredConfigs = map[string]struct{}{}
	)

	configCh, err := s.AllConfigs(ctx)
	if err != nil {
		level.Error(s.logger).Log("msg", "failed getting config list when resharding", "err", err)
		return nil, err
	}
	for ch := range configCh {
		discoveredConfigs[ch.Name] = struct{}{}
		if err := s.im.ApplyConfig(ch); err != nil {
			level.Error(s.logger).Log("msg", "failed to apply config when resharding", "err", err)
		}
	}

	// Find the set of configs that weren't present in the KV store response but are
	// being tracked locally. Any config that wasn't still in the KV store must be
	// removed from our tracked list.
	for runningConfig := range currentConfigs {
		_, keyInStore := discoveredConfigs[runningConfig]
		if keyInStore {
			continue
		}

		level.Info(s.logger).Log("msg", "deleting config removed from store", "name", runningConfig)
		err := s.im.DeleteConfig(runningConfig)
		if err != nil {
			level.Error(s.logger).Log("msg", "failed to delete stale config", "err", err)
		}
	}

	return &empty.Empty{}, nil
}

// AllConfigs gets all configs known to the KV store.
func (s *Server) AllConfigs(ctx context.Context) (<-chan instance.Config, error) {
	keys, err := s.kv.List(ctx, "")
	if err != nil {
		return nil, err
	}

	ch := make(chan instance.Config)

	var wg sync.WaitGroup
	wg.Add(len(keys))
	go func() {
		wg.Wait()
		close(ch)
	}()

	for _, key := range keys {
		go func(key string) {
			defer wg.Done()

			// TODO(rfratto): retries might be useful here
			v, err := s.kv.Get(ctx, key)
			if err != nil {
				level.Error(s.logger).Log("msg", "failed to get config with key", "key", key, "err", err)
				return
			} else if v == nil {
				level.Warn(s.logger).Log("skipping key that was deleted after list was called", "key", key)
				return
			}

			cfg := v.(*instance.Config)
			ch <- *cfg
		}(key)
	}
	return ch, nil
}

// ReadRing is a subset of the Cortex ring.ReadRing interface with only the
// functionality used by the HA server.
type ReadRing interface {
	http.Handler

	Get(key uint32, op ring.Operation, bufDescs []ring.IngesterDesc, bufHosts, bufZones []string) (ring.ReplicationSet, error)
	GetAllHealthy(op ring.Operation) (ring.ReplicationSet, error)
}

// ShardingInstanceManager wraps around an existing instance.Manager and uses a
// hash ring to determine if a config should be applied. If an applied
// config used to be owned by the local address but no longer does, it
// will be deleted on the next apply.
type ShardingInstanceManager struct {
	log   log.Logger
	inner instance.Manager
	ring  ReadRing
	addr  string

	keyToHash map[string]uint32
}

// NewShardingInstanceManager creates a new ShardingInstanceManager that wraps
// around an underlying instance.Manager. ring and addr are used together to do
// hash ring lookups; for a given applied config, it is owned by the instance of
// ShardingInstanceManager if looking up its hash in the ring results in the
// address specified by addr.
func NewShardingInstanceManager(logger log.Logger, wrap instance.Manager, ring ReadRing, addr string) ShardingInstanceManager {
	return ShardingInstanceManager{
		log:       logger,
		inner:     wrap,
		ring:      ring,
		addr:      addr,
		keyToHash: make(map[string]uint32),
	}
}

// ListInstances returns the list of instances that have been applied through
// the ShardingInstanceManager. It will return a subset of the overall
// set of configs passed to the instance.Manager as a whole.
//
// Returning the subset of configs that only the ShardingInstanceManager
// applied itself allows for the underlying instance.Manager to manage
// its own set of configs that will not be affected by the scraping
// service resharding and deleting configs that aren't found in the KV
// store.
func (m ShardingInstanceManager) ListInstances() map[string]instance.ManagedInstance {
	inner := m.inner.ListInstances()
	sharded := make(map[string]instance.ManagedInstance, len(inner))

	for k, v := range inner {
		if _, isSharded := m.keyToHash[k]; isSharded {
			sharded[k] = v
		}
	}

	return sharded
}

// ListConfigs returns the list of configs that have been applied through
// the ShardingInstanceManager. It will return a subset of the overall
// set of configs passed to the instance.Manager as a whole.
//
// Returning the subset of configs that only the ShardingInstanceManager
// applied itself allows for the underlying instance.Manager to manage
// its own set of configs that will not be affected by the scraping
// service resharding and deleting configs that aren't found in the KV
// store.
func (m ShardingInstanceManager) ListConfigs() map[string]instance.Config {
	inner := m.inner.ListConfigs()
	sharded := make(map[string]instance.Config, len(inner))

	for k, v := range inner {
		if _, isSharded := m.keyToHash[k]; isSharded {
			sharded[k] = v
		}
	}

	return sharded
}

// ApplyConfig implements instance.Manager.ApplyConfig.
func (m ShardingInstanceManager) ApplyConfig(c instance.Config) error {
	keyHash := configKeyHash(&c)
	owned, err := m.owns(keyHash)
	if err != nil {
		level.Error(m.log).Log("msg", "failed to check if a config is owned, skipping config until next reshard", "err", err)
		return nil
	}

	if owned {
		hash, err := configHash(&c)
		if err != nil {
			return fmt.Errorf("failed to hash config: %w", err)
		}

		// If the config is unchanged, do nothing.
		if m.keyToHash[c.Name] == hash {
			return nil
		}

		level.Info(m.log).Log("msg", "detected new or changed config", "name", c.Name)
		m.keyToHash[c.Name] = hash
		return m.inner.ApplyConfig(c)
	}

	// If we don't own the config, it's possible that we owned it before
	// and need to delete it now.
	return m.DeleteConfig(c.Name)
}

// DeleteConfig implements instance.Manager.DeleteConfig.
func (m ShardingInstanceManager) DeleteConfig(name string) error {
	// Doesn't exist, ignore.
	if _, exist := m.keyToHash[name]; !exist {
		return nil
	}

	level.Info(m.log).Log("msg", "removing config", "name", name)
	err := m.inner.DeleteConfig(name)
	if err == nil {
		delete(m.keyToHash, name)
	}
	return err
}

// Stop implements instance.Manager.Stop.
func (m ShardingInstanceManager) Stop() { m.inner.Stop() }

// owns checks if the ShardingInstanceManager is responsible for
// a given hash.
func (m ShardingInstanceManager) owns(hash uint32) (bool, error) {
	rs, err := m.ring.Get(hash, ring.Write, nil, nil, nil)
	if err != nil {
		return false, err
	}
	for _, r := range rs.Ingesters {
		if r.Addr == m.addr {
			return true, nil
		}
	}
	return false, nil
}

// configHash returns the hash of the entirety of an instance config.
func configHash(c *instance.Config) (uint32, error) {
	val, err := instance.MarshalConfig(c, false)
	if err != nil {
		return 0, err
	}
	h := fnv.New32()
	_, _ = h.Write(val)
	return h.Sum32(), nil
}

// configKeyHash gets a hash for a config that is used to determine ownership.
// It is based on primary keys of the instance config rather than the entire
// config.
func configKeyHash(c *instance.Config) uint32 {
	h := fnv.New32()
	_, _ = h.Write([]byte(c.Name))
	return h.Sum32()
}
