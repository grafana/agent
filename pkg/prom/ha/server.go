// Package ha implements a high availability clustering mode for the agent. It
// is also referred to as the "scraping service" mode, as this is a fairly
// accurate description of what it does: a series of configs are stored in a
// KV store and a cluster of agents pulls configs from the store and shards
// them amongst the cluster, thereby distributing scraping load.
package ha

import (
	"context"
	"flag"
	"fmt"
	"sync"
	"time"

	"github.com/cortexproject/cortex/pkg/ring"
	"github.com/cortexproject/cortex/pkg/ring/kv"
	"github.com/cortexproject/cortex/pkg/util"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/google/go-cmp/cmp"
	"github.com/gorilla/mux"
	"github.com/grafana/agent/pkg/agentproto"
	"github.com/grafana/agent/pkg/prom/ha/client"
	"github.com/grafana/agent/pkg/prom/instance"
	"github.com/grafana/agent/pkg/prom/instance/configstore"
	flagutil "github.com/grafana/agent/pkg/util"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/prometheus/config"
	"github.com/weaveworks/common/user"
	"google.golang.org/grpc"
)

var (
	reshardDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name: "agent_prometheus_scraping_service_reshard_duration",
		Help: "How long it took for resharding to run.",
	}, []string{"success"})
)

var (
	// util.BackoffConfig used for a backoff period when resharding after join.
	backoffConfig = util.BackoffConfig{
		MinBackoff: time.Second,
		MaxBackoff: 2 * time.Minute,
		MaxRetries: 10,
	}

	// DefaultConfig provides default values for the config
	DefaultConfig = *flagutil.DefaultConfigFromFlags(&Config{}).(*Config)
)

// Config describes how to instantiate a scraping service Server instance.
type Config struct {
	Enabled         bool                  `yaml:"enabled"`
	ReshardInterval time.Duration         `yaml:"reshard_interval"`
	ReshardTimeout  time.Duration         `yaml:"reshard_timeout"`
	KVStore         kv.Config             `yaml:"kvstore"`
	Lifecycler      ring.LifecyclerConfig `yaml:"lifecycler"`
}

// UnmarshalYAML implements yaml.Unmarshaler.
func (c *Config) UnmarshalYAML(unmarshal func(interface{}) error) error {
	*c = DefaultConfig

	type plain Config
	return unmarshal((*plain)(c))
}

// RegisterFlags adds the flags required to config the Server to the given
// FlagSet.
func (c *Config) RegisterFlags(f *flag.FlagSet) {
	c.RegisterFlagsWithPrefix("", f)
}

// RegisterFlagsWithPrefix adds the flags required to config this to the given
// FlagSet with a specified prefix.
func (c *Config) RegisterFlagsWithPrefix(prefix string, f *flag.FlagSet) {
	f.BoolVar(&c.Enabled, prefix+"enabled", false, "enables the scraping service mode")
	f.DurationVar(&c.ReshardInterval, prefix+"reshard-interval", time.Minute*1, "how often to manually reshard")
	f.DurationVar(&c.ReshardTimeout, prefix+"reshard-timeout", time.Second*30, "timeout for cluster-wide reshards and local reshards. Timeout of 0s disables timeout.")
	c.KVStore.RegisterFlagsWithPrefix(prefix+"config-store.", "configurations/", f)
	c.Lifecycler.RegisterFlagsWithPrefix(prefix, f)
}

// Server implements the HA scraping service.
type Server struct {
	logger log.Logger

	cfgMut             sync.RWMutex
	cfg                Config
	clientConfig       client.Config
	globalConfig       *config.GlobalConfig
	defaultRemoteWrite []*instance.RemoteWriteConfig

	configManagerMut sync.Mutex
	im               instance.Manager
	configs          map[string]struct{}

	node     *node
	store    *configstore.Remote
	storeAPI *configstore.API

	stateMut sync.Mutex
	reload   chan struct{}
	exited   chan bool
}

// New creates a new HA scraping service instance.
func New(reg prometheus.Registerer, cfg Config, globalConfig *config.GlobalConfig, clientConfig client.Config, logger log.Logger, im instance.Manager, defaultRemoteWrite []*instance.RemoteWriteConfig) (*Server, error) {
	logger = log.With(logger, "component", "ha")

	// Force ReplicationFactor to be 1, since replication isn't supported for the
	// scraping service yet.
	cfg.Lifecycler.RingConfig.ReplicationFactor = 1

	store, err := configstore.NewRemote(logger, reg, cfg.KVStore)
	if err != nil {
		return nil, err
	}

	storeAPI := configstore.NewAPI(logger, store, func(c *instance.Config) error {
		return c.ApplyDefaults(globalConfig, defaultRemoteWrite)
	})
	reg.MustRegister(storeAPI)

	s := &Server{
		cfg:          cfg,
		globalConfig: globalConfig,
		clientConfig: clientConfig,
		logger:       logger,

		im:      im,
		configs: make(map[string]struct{}),

		store:    store,
		storeAPI: storeAPI,

		reload: make(chan struct{}, 1),
		exited: make(chan bool),

		defaultRemoteWrite: defaultRemoteWrite,
	}

	s.node, err = newNode(reg, logger, cfg, s)

	if err := s.ApplyConfig(cfg, globalConfig, clientConfig, defaultRemoteWrite); err != nil {
		return nil, fmt.Errorf("failed to apply config: %w", err)
	}

	go s.run()
	return s, nil
}

func (s *Server) ApplyConfig(cfg Config, globalConfig *config.GlobalConfig, clientConfig client.Config, defaultRemoteWrite []*instance.RemoteWriteConfig) error {
	s.stateMut.Lock()
	defer s.stateMut.Unlock()

	if s.hasExited() {
		return fmt.Errorf("clustering server not running")
	}

	s.cfgMut.Lock()
	defer s.cfgMut.RUnlock()

	if cmp.Equal(cfg, s.cfg) && cmp.Equal(globalConfig, s.globalConfig) && cmp.Equal(clientConfig, s.clientConfig) && cmp.Equal(defaultRemoteWrite, s.defaultRemoteWrite) {
		// Nothing changed, quit early.
		return nil
	}

	s.cfg = cfg
	s.globalConfig = globalConfig
	s.clientConfig = clientConfig
	s.defaultRemoteWrite = defaultRemoteWrite

	if err := s.store.ApplyConfig(cfg.KVStore); err != nil {
		return err
	}

	if err := s.node.ApplyConfig(cfg); err != nil {
		return err
	}

	// Reloading will cause a local reshard to run, which will also re-apply the
	// newest set of defaults to the loaded configs. As long as this is the only
	// writer of s.reload, this will never block.
	s.reload <- struct{}{}
	return nil
}

func (s *Server) run() {
	defer close(s.exited)

	var (
		ctx    context.Context
		cancel context.CancelFunc

		wg sync.WaitGroup
	)

	for range s.reload {
		// Create a new context. If there was a previous cancel set up, call it
		// to free the previous resources.
		if cancel != nil {
			cancel()
		}
		ctx, cancel = context.WithCancel(context.Background())

		// Wait for previous resources to release.
		wg.Wait()

		// Perform join operations. Joining can't fail; any failed operation
		// performed will eventually correct itself due to how all Agents in the
		// cluster reshard themselves every reshard_interval.
		s.join(ctx)

		wg.Add(2)
		go func() {
			s.watchKV(ctx)
			wg.Done()
		}()

		go func() {
			s.reshardLoop(ctx)
			wg.Done()
		}()
	}

	// Close the latest context and wait for the goroutines to stop.
	if cancel != nil {
		cancel()
	}
	wg.Wait()

	level.Info(s.logger).Log("msg", "run loop exited")
}

func (s *Server) join(ctx context.Context) {
	if s.cfg.ReshardTimeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, s.cfg.ReshardTimeout)
		defer cancel()
	}

	level.Info(s.logger).Log("msg", "running cluster-wide reshard")
	if err := s.waitNotifyReshard(ctx); err != nil {
		level.Error(s.logger).Log("msg", "could not run cluster-wide reshard", "err", err)
	}

	level.Info(s.logger).Log("msg", "running local reshard")
	if _, err := s.Reshard(ctx, &agentproto.ReshardRequest{}); err != nil {
		level.Error(s.logger).Log("msg", "failed running local reshard", "err", err)
	}
}

func (s *Server) waitNotifyReshard(ctx context.Context) error {
	level.Info(s.logger).Log("msg", "performing cluster-wide reshard")

	// IteratePeers calls the callback in parallel with a 250ms jitter between them.
	return s.node.IteratePeers(ctx, func(ctx context.Context, desc *ring.InstanceDesc) error {
		ctx = user.InjectOrgID(ctx, "fake")
		return s.notifyReshard(ctx, desc)
	})
}

func (s *Server) notifyReshard(ctx context.Context, desc *ring.InstanceDesc) error {
	s.cfgMut.RLock()
	defer s.cfgMut.RUnlock()

	cli, err := client.New(s.clientConfig, desc.Addr)
	if err != nil {
		return err
	}
	defer cli.Close()

	backoff := util.NewBackoff(ctx, backoffConfig)
	for backoff.Ongoing() {
		level.Info(s.logger).Log("msg", "telling remote agent to reshard", "addr", desc.Addr)
		_, err := cli.Reshard(ctx, &agentproto.ReshardRequest{})
		if err == nil {
			break
		}

		level.Warn(s.logger).Log("msg", "failed to tell remote agent to reshard", "err", err, "addr", desc.Addr)
		backoff.Wait()
	}

	return backoff.Err()
}

func (s *Server) watchKV(ctx context.Context) {
	level.Info(s.logger).Log("msg", "watching for changes to configs")

	handleEvent := func(ev configstore.WatchEvent) {
		s.configManagerMut.Lock()
		defer s.configManagerMut.Unlock()

		if ctx.Err() != nil {
			return
		}

		var (
			_, isRunning = s.configs[ev.Key]
			isDeleted    = ev.Config == nil
		)

		owned, err := s.node.Owns(ev.Key)
		if err != nil {
			level.Error(s.logger).Log("msg", "failed to see if config is owned, will retry on next reshard", "name", ev.Key, "err", err)
			return
		}

		switch {
		// Two deletion scenarios:
		// 1. A config we're running got moved to a new owner
		// 2. A config we're running got deleted
		case (isRunning && !owned) || (isDeleted && isRunning):
			if err := s.im.DeleteConfig(ev.Key); err != nil {
				level.Error(s.logger).Log("msg", "failed to delete config", "name", ev.Key, "err", err)
			}
			delete(s.configs, ev.Key)

		// New config should be applied if we own it
		case !isDeleted && owned:
			if s.applyConfig(ev.Key, ev.Config) {
				s.configs[ev.Key] = struct{}{}
			}
		}
	}

	storeEvents := s.store.Watch()

Outer:
	for {
		select {
		case <-ctx.Done():
			break Outer
		case ev := <-storeEvents:
			handleEvent(ev)
		}
	}

	level.Info(s.logger).Log("msg", "stopped watching for changes to configs")
}

// applyConfig applies a config to the InstanceManager. Returns true if the
// application succeed.
func (s *Server) applyConfig(key string, cfg *instance.Config) bool {
	s.cfgMut.RLock()
	defer s.cfgMut.RUnlock()

	// Configs from the store aren't immediately valid and must be given the
	// global config before running. Configs are validated against the current
	// global config at upload time, but if the global config has since changed,
	// they can be invalid at read time.
	if err := cfg.ApplyDefaults(s.globalConfig, s.defaultRemoteWrite); err != nil {
		level.Error(s.logger).Log("msg", "failed to apply defaults to config. this config cannot run until the globals are adjusted or the config is updated with either explicit overrides to defaults or tweaked to operate within the globals", "name", key, "err", err)
		return false
	}

	// Applying configs should only fail if the config is invalid
	err := s.im.ApplyConfig(*cfg)
	if err != nil {
		level.Error(s.logger).Log("msg", "failed to apply config, will retry on next reshard", "name", key, "err", err)
		return false
	}

	return true
}

func (s *Server) reshardLoop(ctx context.Context) {
	for {
		s.cfgMut.RLock()
		tickTime := s.cfg.ReshardInterval
		s.cfgMut.RUnlock()

		select {
		case <-time.After(tickTime):
			s.localReshard(ctx)
		case <-ctx.Done():
			return
		case <-s.exited:
			return
		}
	}
}

func (s *Server) localReshard(ctx context.Context) {
	s.cfgMut.RLock()
	reshardTimeout := s.cfg.ReshardTimeout
	s.cfgMut.RUnlock()

	if reshardTimeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, reshardTimeout)
		defer cancel()
	}

	level.Info(s.logger).Log("msg", "resharding agent", "timeout", reshardTimeout)
	_, err := s.Reshard(ctx, &agentproto.ReshardRequest{})
	if err != nil {
		level.Error(s.logger).Log("msg", "resharding failed", "err", err)
	}
}

// WireAPI injects routes into the provided mux router for the config
// management API.
func (s *Server) WireAPI(r *mux.Router) {
	s.storeAPI.WireAPI(r)
	s.node.WireAPI(r)
}

// WireGRPC injects gRPC server handlers into the provided gRPC server.
func (s *Server) WireGRPC(srv *grpc.Server) {
	agentproto.RegisterScrapingServiceServer(srv, s)
}

// Flush satisfies ring.FlushTransferer. It is a no-op for the Agent.
func (s *Server) Flush() {}

// TransferOut satisfies ring.FlushTransferer. It connects to all other
// healthy agents in the cluster and tells them to reshard.
func (s *Server) TransferOut(ctx context.Context) error {
	s.cfgMut.RLock()
	reshardTimeout := s.cfg.ReshardTimeout
	s.cfgMut.RUnlock()

	if reshardTimeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, reshardTimeout)
		defer cancel()
	}

	level.Info(s.logger).Log("msg", "leaving cluster, starting cluster-wide reshard process", "timeout", reshardTimeout)
	return s.waitNotifyReshard(ctx)
}

// Stop stops the HA server and its dependencies. Once stopped, the Server cannot run again.
func (s *Server) Stop() error {
	s.stateMut.Lock()
	defer s.stateMut.Unlock()

	if s.hasExited() {
		return fmt.Errorf("already exited")
	}

	// Close the reload loop and wait for it to stop.
	close(s.reload)
	<-s.exited

	// Stop the dependencies now and wait to return the error until after we've
	// stopped running any local scrape jobs.
	err := s.node.Stop()

	// Delete all the local configs that were running.
	s.configManagerMut.Lock()
	defer s.configManagerMut.Unlock()

	for cfg := range s.configs {
		if err := s.im.DeleteConfig(cfg); err != nil {
			level.Warn(s.logger).Log("msg", "failed to delete config on shutdown", "config", cfg, "err", err)
		}

		// Deletes only fail if the config doesn't exist, so either way we want to
		// stop tracking it here.
		delete(s.configs, cfg)
	}

	return err
}

func (s *Server) hasExited() bool {
	select {
	case <-s.exited:
		return true
	default:
		return false
	}
}
