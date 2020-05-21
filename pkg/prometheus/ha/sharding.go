package ha

import (
	"context"
	"hash/fnv"
	"net/http"
	"sync"
	"time"

	"github.com/cortexproject/cortex/pkg/ring"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/grafana/agent/pkg/agentproto"
	"github.com/grafana/agent/pkg/prometheus/instance"
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
		currentConfigs = s.cm.ListConfigs()

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
		s.cm.ApplyConfig(ch)
	}

	// Find the set of configs that weren't present in the KV store response but are
	// being tracked locally. Any config that wasn't still in the KV store must be
	// removed from our tracked list.
	for runningConfig := range currentConfigs {
		_, keyInStore := discoveredConfigs[runningConfig]
		if keyInStore {
			continue
		}

		err := s.cm.DeleteConfig(runningConfig)
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
				level.Error(s.logger).Log("failed to get key for resharding", "key", key)
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

	Get(key uint32, op ring.Operation, buf []ring.IngesterDesc) (ring.ReplicationSet, error)
	GetAll() (ring.ReplicationSet, error)
}

// ShardingConfigManager wraps around an existing ConfigManager and uses a
// hash ring to determine if a config should be applied. If an applied
// config used to be owned by the local address but no longer does, it
// will be deleted on the next apply.
type ShardingConfigManager struct {
	log   log.Logger
	inner ConfigManager
	ring  ReadRing
	addr  string

	keyToHash map[string]uint32
}

// NewShardingConfigManager creates a new ShardingConfigManager. The wrap
// argument holds the underlying config manager, while ring and addr are
// used together to do hash ring lookups: for a given hash, a config is
// owned by the ShardingConfigManager if the address of a node from a
// lookup matches the addr argument passed to NewShardingConfigManager.
func NewShardingConfigManager(logger log.Logger, wrap ConfigManager, ring ReadRing, addr string) ShardingConfigManager {
	return ShardingConfigManager{
		log:       logger,
		inner:     wrap,
		ring:      ring,
		addr:      addr,
		keyToHash: make(map[string]uint32),
	}
}

// ListConfigs implements ConfigManager.ListConfigs.
func (m ShardingConfigManager) ListConfigs() map[string]instance.Config {
	// Direct pass through; no sharding needed here.
	return m.inner.ListConfigs()
}

// ApplyConfig implements ConfigManager.ApplyConfig.
func (m ShardingConfigManager) ApplyConfig(c instance.Config) {
	hash, err := configHash(&c)
	if err != nil {
		level.Error(m.log).Log("msg", "failed to hash config", "err", err)
		return
	}

	owned, err := m.owns(hash)
	if err != nil {
		level.Error(m.log).Log("msg", "failed to check if a config is owned, skipping config until next reshard", "err", err)
		return
	}

	if owned {
		// If the config is unchanged, do nothing.
		if m.keyToHash[c.Name] == hash {
			return
		}

		level.Info(m.log).Log("msg", "detected new or changed config", "name", c.Name)
		m.keyToHash[c.Name] = hash
		m.inner.ApplyConfig(c)
	} else {
		// If we don't own the config, it's possible that we owned it before
		// and need to delete it now.
		err := m.DeleteConfig(c.Name)
		if err != nil {
			level.Error(m.log).Log("msg", "failed to delete stale config", "err", err)
		}
	}
}

// DeleteConfig implements ConfigManager.DeleteConfig.
func (m ShardingConfigManager) DeleteConfig(name string) error {
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

// owns checks if the ShardingConfigManager is responsible for
// a given hash.
func (m ShardingConfigManager) owns(hash uint32) (bool, error) {
	rs, err := m.ring.Get(hash, ring.Write, nil)
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

func configHash(c *instance.Config) (uint32, error) {
	val, err := instance.MarshalConfig(c, false)
	if err != nil {
		return 0, err
	}
	h := fnv.New32()
	_, _ = h.Write([]byte(val))
	return h.Sum32(), nil
}
