package ha

import (
	"context"
	"hash/fnv"
	"sync"
	"time"

	"github.com/cortexproject/cortex/pkg/ring"
	"github.com/go-kit/kit/log/level"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/grafana/agent/pkg/agentproto"
	"github.com/grafana/agent/pkg/prometheus/instance"
	"gopkg.in/yaml.v2"
)

// HashedConfig is a config with a hash associated with its content. Used
// for sharding configs to an instance of Server.
type HashedConfig struct {
	instance.Config
	hash uint32
}

// HashedConfig converts an instance.Config into a HashedConfig.
func HashConfig(c instance.Config) (HashedConfig, error) {
	bb, err := yaml.Marshal(c)
	if err != nil {
		return HashedConfig{}, err
	}
	h := fnv.New32()
	_, _ = h.Write(bb)
	v := h.Sum32()

	return HashedConfig{Config: c, hash: v}, nil
}

// Hash returns the hash of the HashedConfig.
func (c HashedConfig) Hash() uint32 { return c.hash }

// Reshard initiates an entire reshard of the current HA scraping service instance.
// All configs will be reloaded from the KV store and the scraping service instance
// will see what should be managed locally.
//
// Satisfies agentproto.ScrapingServiceServer.
func (s *Server) Reshard(ctx context.Context, _ *agentproto.ReshardRequest) (_ *empty.Empty, err error) {
	s.shardMut.Lock()
	defer s.shardMut.Unlock()

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
		key := ch.Name
		discoveredConfigs[key] = struct{}{}
		s.processKey(key, &ch)
	}

	// Find the set of configs that weren't present in the KV store response but are
	// being tracked locally. Any config that wasn't still in the KV store must be
	// removed from our tracked list.
	for runningConfig := range currentConfigs {
		_, keyInStore := discoveredConfigs[runningConfig]
		if !keyInStore {
			s.untrackConfig(runningConfig)
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

// OwnsConfig returns true if a given HashedConfig matches any of the
// tokens belonging to the current server.
func (s *Server) OwnsConfig(c HashedConfig) (bool, error) {
	rs, err := s.ring.Get(c.Hash(), ring.Write, nil)
	if err != nil {
		return false, err
	}
	for _, r := range rs.Ingesters {
		if r.Addr == s.lc.Addr {
			return true, nil
		}
	}
	return false, nil
}

// processKey handles a config name and value from the store. One of the
// following will happen:
//
// 1. If the config is nil, the key will be treated as deleted and removed.
//
// 2. If the config is owned by the current Server, it will be tracked.
//
// 3. If the config is not owned by the current Server, it will be removed
//    from the tracked list if it is already tracked.
func (s *Server) processKey(key string, v interface{}) {
	if v == nil {
		level.Info(s.logger).Log("msg", "detected deleted config", "name", key)
		s.untrackConfig(key)
		return
	}

	cfg, err := HashConfig(*v.(*instance.Config))
	if err != nil {
		level.Error(s.logger).Log("msg", "failed to hash config", "err", err)
	}

	owned, err := s.OwnsConfig(cfg)
	if err != nil {
		level.Error(s.logger).Log("msg", "failed to check if a config is owned. if the config is owned by this ingester, it won't be used until at least the next reshard period.", "err", err)
	}

	if owned {
		s.trackConfig(cfg)
	} else {
		s.untrackConfig(key)
	}
}

// untrackConfig will remove a track config if it's currently being tracked.
// untrackConfig is a no-op if the config is not already tracked.
func (s *Server) untrackConfig(key string) {
	_, owned := s.keyToHash[key]
	if !owned {
		return
	}

	level.Info(s.logger).Log("msg", "untracking config", "name", key)
	if err := s.cm.DeleteConfig(key); err != nil {
		level.Error(s.logger).Log("msg", "failed to remove config", "name", key)
	}
	delete(s.keyToHash, key)
}

// trackConfig tracks a config. If the config is already tracked and the hash
// hasn't changed since the last time it was tracked, nothing happens here.
// Otherwise, trackConfig tracks the hash of the config and passes it through
// to the Server's ConfigManager.
//
// trackConfig does not check if the Server owns the HashedConfig; this is the
// responsibility of the caller.
func (s *Server) trackConfig(c HashedConfig) {
	// If we're already tracking this config and its hash hasn't changed, we
	// don't need to do anything here.
	if s.keyToHash[c.Name] == c.Hash() {
		return
	}

	// This is a new or updated config: we need to pass it to our ConfigManager
	// and track it locally.
	level.Info(s.logger).Log("msg", "tracking new config", "name", c.Name)
	s.keyToHash[c.Name] = c.Hash()
	s.cm.ApplyConfig(c.Config)
}
