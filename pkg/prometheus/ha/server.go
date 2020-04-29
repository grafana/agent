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
	"fmt"
	"hash/fnv"
	"sync"
	"time"

	"github.com/cortexproject/cortex/pkg/ring"
	"github.com/cortexproject/cortex/pkg/ring/kv"
	"github.com/cortexproject/cortex/pkg/util"
	"github.com/cortexproject/cortex/pkg/util/services"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/grafana/agent/pkg/agentproto"
	"github.com/grafana/agent/pkg/prometheus/ha/client"
	"github.com/grafana/agent/pkg/prometheus/instance"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"google.golang.org/grpc"
	"gopkg.in/yaml.v2"
)

var (
	reshardDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name: "agent_prometheus_scraping_service_reshard_duration",
		Help: "How long it took for resharding to run.",
	}, []string{"success"})
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
	Enabled         bool                  `yaml:"enabled"`
	ReshardInterval time.Duration         `yaml:"reshard_interval"`
	KVStore         kv.Config             `yaml:"kvstore"`
	Lifecycler      ring.LifecyclerConfig `yaml:"lifecycler"`
}

// RegisterFlagsWithPrefix adds the flags required to config this to the given
// FlagSet with a specified prefix.
func (c *Config) RegisterFlagsWithPrefix(prefix string, f *flag.FlagSet) {
	f.BoolVar(&c.Enabled, prefix+"enabled", false, "enables the scraping service mode")
	f.DurationVar(&c.ReshardInterval, prefix+"reshard-interval", time.Minute*1, "how often to manually reshard")
	c.KVStore.RegisterFlagsWithPrefix(prefix+"config-store.", "configurations/", f)
	c.Lifecycler.RegisterFlagsWithPrefix(prefix, f)
}

// Server implements the HA scraping service.
type Server struct {
	cfg          Config
	clientConfig client.Config
	logger       log.Logger

	cm ConfigManager

	// Mutex to lock during sharding or responding to a KV store event.
	shardMut sync.Mutex

	// Stored hashes for a key
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
func New(cfg Config, clientConfig client.Config, logger log.Logger, cm ConfigManager) (*Server, error) {
	if cfg.Lifecycler.RingConfig.ReplicationFactor != 1 {
		return nil, errors.New("replication_factor must be 1")
	}

	s := &Server{
		cfg:          cfg,
		clientConfig: clientConfig,
		logger:       log.With(logger, "component", "ha"),

		cm: cm,

		keyToHash: make(map[string]uint32),
	}

	var err error
	s.kv, err = kv.NewClient(cfg.KVStore, GetCodec())
	if err != nil {
		return nil, err
	}

	// Start the ring first so there's a chance that it will be filled by the time we want to
	// tell other agents to reshard.
	s.ring, err = ring.New(cfg.Lifecycler.RingConfig, "agent_viewer", "agent")
	if err != nil {
		return nil, err
	}
	if err := s.ring.StartAsync(context.Background()); err != nil {
		return nil, err
	}

	// TODO(rfratto): switching to a BasicLifecycler would be nice here, it'd allow
	// the joining/leaving process to include waiting for resharding.
	s.lc, err = ring.NewLifecycler(cfg.Lifecycler, s, "agent", "agent", true)
	if err != nil {
		return nil, err
	}
	if err := services.StartAndAwaitRunning(context.Background(), s.lc); err != nil {
		return nil, err
	}

	if err := s.waitNotifyReshard(); err != nil {
		return nil, fmt.Errorf("could not run cluster-wide reshard: %w", err)
	}

	level.Info(s.logger).Log("msg", "cluster-wide reshard finished. running local reshard")
	if _, err := s.Reshard(context.Background(), &agentproto.ReshardRequest{}); err != nil {
		return nil, fmt.Errorf("failed running local reshard: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	s.cancel = cancel

	go s.loop(ctx)
	return s, nil
}

func (s *Server) waitNotifyReshard() error {
	level.Info(s.logger).Log("msg", "retrieving agents in cluster for cluster-wide reshard")

	var (
		rs  ring.ReplicationSet
		err error
	)

	backoff := util.NewBackoff(context.Background(), util.BackoffConfig{
		MinBackoff: time.Second,
		MaxBackoff: 2 * time.Minute,
		MaxRetries: 10,
	})
	for backoff.Ongoing() {
		rs, err = s.ring.GetAll()
		if err == nil {
			break
		}

		level.Warn(s.logger).Log("msg", "could not get agents in cluster", "err", err)
		backoff.Wait()
	}
	if err := backoff.Err(); err != nil {
		return err
	}

	return s.notifyReshard(context.Background(), rs)
}

func (s *Server) notifyReshard(ctx context.Context, set ring.ReplicationSet) error {
	_, err := set.Do(ctx, time.Millisecond*250, func(desc *ring.IngesterDesc) (interface{}, error) {
		// Skip over ourselves; we'll reshard after this process finishes.
		if desc.Addr == s.lc.Addr {
			return nil, nil
		}

		cli, err := client.New(s.clientConfig, desc.Addr)
		if err != nil {
			return nil, err
		}
		defer cli.Close()

		return cli.Reshard(ctx, &agentproto.ReshardRequest{})
	})
	return err
}

func (s *Server) loop(ctx context.Context) {
	kvWatchExit := make(chan bool)

	// KV watch loop
	go func() {
		defer close(kvWatchExit)

		level.Info(s.logger).Log("msg", "watching for changes to configs")

		s.kv.WatchPrefix(ctx, "", func(key string, v interface{}) bool {
			s.shardMut.Lock()
			defer s.shardMut.Unlock()

			if ctx.Err() != nil {
				return false
			}

			s.processKey(key, v)
			return true
		})

		level.Info(s.logger).Log("msg", "stopped watching for changes to configs")
	}()

	// Reshard ticker loop
	go func() {
		level.Info(s.logger).Log("msg", "resharding agent on interval", "interval", s.cfg.ReshardInterval)
		t := time.NewTicker(s.cfg.ReshardInterval)
		defer t.Stop()

		for {
			select {
			case <-t.C:
				level.Info(s.logger).Log("msg", "resharding agent")
				_, err := s.Reshard(ctx, &agentproto.ReshardRequest{})
				if err != nil {
					level.Error(s.logger).Log("msg", "resharding failed", "err", err)
				}
			case <-ctx.Done():
				return
			case <-s.exited:
				return
			}
		}
	}()

	<-kvWatchExit
	defer close(s.exited)
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
		level.Error(s.logger).Log("msg", "failed to check if a config is owned. if the config is owned by this ingester, it won't be used until at least the next poll period.", "err", err)
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

// WireGRPC injects gRPC server handlers into the provided gRPC server.
func (s *Server) WireGRPC(srv *grpc.Server) {
	agentproto.RegisterScrapingServiceServer(srv, s)
}

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

// Flush satisfies ring.FlushTransferer. It is a no-op for the Agent.
func (s *Server) Flush() {}

// TransferOut satisfies ring.FlushTransferer. It connects to all other
// healthy agents in the cluster and tells them to reshard.
func (s *Server) TransferOut(ctx context.Context) error {
	rs, err := s.ring.GetAll()
	if err != nil {
		return err
	}

	// Note that we're still in the ring at this point but we're marked
	// as LEAVING. So when other agents reshard, OwnsConfig will properly
	// detect that our agent is about to leave and will reshard as if
	// we're already completely out of the ring.
	return s.notifyReshard(ctx, rs)
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
