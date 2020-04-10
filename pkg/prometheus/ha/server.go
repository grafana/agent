// Package ha implements a high availability clustering mode for the agent. It
// is also referred to as the "scraping service" mode, as this is a fairly
// accurate description of what it does: a series of configs are stored in a
// KV store and a cluster of agents pulls configs from the store and shards
// them amongst the cluster, thereby distributing scraping load.
package ha

import (
	"context"
	"errors"
	"flag"
	"hash/fnv"
	"sync"

	"github.com/cortexproject/cortex/pkg/ring"
	"github.com/cortexproject/cortex/pkg/ring/kv"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/grafana/agent/pkg/prometheus/instance"
	"gopkg.in/yaml.v2"
)

// ConfigManager is an interface to manipulating a set of running
// instance.Configs. It is satisfied by the ConfigManager struct in
// pkg/prometheus, but is provided as an interface here for testing and
// avoiding import cycles.
type ConfigManager interface {
	// ListConfigs gets the list of currently known configs.
	ListConfigs() map[string]instance.Config

	// ApplyConfig adds or updates a config.
	ApplyConfig(c instance.Config)

	// DeleteConfig deletes a config by name, uniquely keyed by the
	// Name field in instance.Config.
	DeleteConfig(name string) error
}

// Config describes how to instantiate a scraping service Server instance.
type Config struct {
	Enabled    bool                  `yaml:"enabled"`
	KVStore    kv.Config             `yaml:"kvstore"`
	Lifecycler ring.LifecyclerConfig `yaml:"lifecycler"`
}

// RegisterFlagsWithPrefix adds the flags required to config this to the given
// FlagSet with a specified prefix.
func (c *Config) RegisterFlagsWithPrefix(prefix string, f *flag.FlagSet) {
	f.BoolVar(&c.Enabled, prefix+"enabled", false, "enables the scraping service mode")
	c.KVStore.RegisterFlagsWithPrefix(prefix+"config-store.", "configurations/", f)
	c.Lifecycler.RegisterFlagsWithPrefix(prefix, f)
}

// Server implements the HA scraping service.
type Server struct {
	cfg    Config
	logger log.Logger

	cm ConfigManager

	// Stored hashes for a key
	hashMut   sync.Mutex
	keyToHash map[string]uint32

	// Interface to storing and retreiving config objects
	kv kv.Client

	// Management for being a cluster member
	lc *ring.Lifecycler

	// View into members of the cluster
	ring *ring.Ring

	cancel context.CancelFunc
	exited chan bool
}

// New creates a new HA scraping service instance.
func New(cfg Config, logger log.Logger, cm ConfigManager) (*Server, error) {
	if cfg.Lifecycler.RingConfig.ReplicationFactor != 1 {
		return nil, errors.New("replication_factor must be 1")
	}

	s := &Server{
		cfg:    cfg,
		logger: log.With(logger, "component", "ha"),

		cm: cm,

		keyToHash: make(map[string]uint32),
	}

	var err error
	s.kv, err = kv.NewClient(cfg.KVStore, GetCodec())
	if err != nil {
		return nil, err
	}

	// TODO(rfratto): switching to a BasicLifecycler would be nice here, it'd allow
	// the joining/leaving process to include waiting for resharding.
	s.lc, err = ring.NewLifecycler(cfg.Lifecycler, nil, "agent", "agent", true)
	if err != nil {
		return nil, err
	}
	s.lc.StartAsync(context.Background())

	s.ring, err = ring.New(cfg.Lifecycler.RingConfig, "agent_viewer", "agent")
	if err != nil {
		return nil, err
	}
	s.ring.StartAsync(context.Background())

	/*
		TODO(rfratto): scraping service node lifecycle

		join process:
			- [x] join cluster
			- [ ] connect to all agents in cluster that is not the local client
			- [ ] tell them to reshard
			- [ ] locally reshard

		loop process:
			- [x] when a key is added, track it if it hashes to self
			- [x] when a key is removed, remove it if it hashes to self
			- [ ] on some interval, manually reshard. helps fight against drift
				from UNHEALTHY nodes or KV not notifying of changes from WatchPrefix

		leaving process (flushTransferer):
			- [ ] connect to all agents in cluster that is not the local client
			- [ ] tell them to reshard
	*/

	ctx, cancel := context.WithCancel(context.Background())
	s.cancel = cancel

	go s.loop(ctx)
	return s, nil
}

func (s *Server) loop(ctx context.Context) {
	kvWatchExit := make(chan bool)

	go func() {
		defer close(kvWatchExit)

		level.Info(s.logger).Log("msg", "watching for changes to configs")

		s.kv.WatchPrefix(ctx, "", func(key string, v interface{}) bool {
			if ctx.Err() != nil {
				return false
			}

			if v == nil {
				s.keyRemoved(key)
				return true
			}

			cfg, err := HashConfig(*v.(*instance.Config))
			if err != nil {
				level.Error(s.logger).Log("msg", "failed to hash config", "err", err)
			}

			owned, err := s.OwnsConfig(cfg)
			if err != nil {
				level.Error(s.logger).Log("msg", "failed to check if a config is owned. if the config is owned by this ingester, it won't be used until at least the next poll period.", "err", err)
			} else if owned {
				s.configUpdated(cfg)
			}

			return true
		})

		level.Info(s.logger).Log("msg", "stopped watching for changes to configs")
	}()

	<-kvWatchExit
	defer close(s.exited)
}

// keyRemoved is called whenever any config key is deleted, even if it is not
// owned by the running Server. keyRemoved will check if the key is tracked
// by the running Server and remove it from the attached ConfigManager.
func (s *Server) keyRemoved(key string) {
	s.hashMut.Lock()
	defer s.hashMut.Unlock()

	_, owned := s.keyToHash[key]
	if !owned {
		return
	}

	level.Info(s.logger).Log("msg", "removing config that was deleted", "name", key)
	if err := s.cm.DeleteConfig(key); err != nil {
		level.Error(s.logger).Log("msg", "failed to stop config", "name", key)
	}
	delete(s.keyToHash, key)
}

// configUpdated is called whenever any config key is added or updated, even
// if it is not owned by the running Server. configUpdated will check if the
// config should be tracked by the running Server and add it to the attached
// ConfigManager.
func (s *Server) configUpdated(c HashedConfig) {
	s.hashMut.Lock()
	defer s.hashMut.Unlock()

	if ok, err := s.OwnsConfig(c); err != nil {
		level.Error(s.logger).Log("msg", "failed to check ownership for created/changed config", "name", c.Name)
	} else if !ok {
		// If we used to track this config, we don't anymore and we need to remove it.
		_, owned := s.keyToHash[c.Name]
		if !owned {
			return
		}

		level.Info(s.logger).Log("msg", "removing config that no longer belongs to agent", "name", c.Name)
		if err := s.cm.DeleteConfig(c.Name); err != nil {
			level.Error(s.logger).Log("msg", "failed to stop config", "name", c.Name)
		}

		delete(s.keyToHash, c.Name)
		return
	}

	// If we're already tracking this config and its hash hasn't changed, we
	// don't need to do anything here.
	if s.keyToHash[c.Name] == c.Hash() {
		return
	}

	// This is a new or updated config: we need to pass it to our configmanager
	// and track it locally.
	level.Info(s.logger).Log("msg", "tracking new config", "name", c.Name)
	s.keyToHash[c.Name] = c.Hash()
	s.cm.ApplyConfig(c.Config)
}

// Stop stops the HA server and its dependencies.
func (s *Server) Stop() error {
	// Close the loop and wait for it to stop.
	s.cancel()
	<-s.exited

	s.ring.StopAsync()
	ringErr := s.ring.AwaitTerminated(context.Background())

	s.lc.StopAsync()
	lcErr := s.lc.AwaitTerminated(context.Background())

	// TODO(rfratto): combine errors?
	if ringErr != nil {
		return ringErr
	}

	return lcErr
}

// AllConfigs gets all configs known to the KV store.
func (s *Server) AllConfigs(ctx context.Context) (<-chan instance.Config, error) {
	keys, err := s.kv.List(ctx, "")
	if err != nil {
		return nil, err
	}

	ch := make(chan instance.Config)
	for _, key := range keys {
		go func(key string) {
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

// HashedConfig is a config with a hash associated with its content.
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
