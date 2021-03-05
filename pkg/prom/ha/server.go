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
	"github.com/cortexproject/cortex/pkg/util/services"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
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
	"go.uber.org/atomic"
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
	cfg          Config
	clientConfig client.Config
	globalConfig *config.GlobalConfig
	reg          prometheus.Registerer
	logger       log.Logger
	addr         string

	configManagerMut sync.Mutex
	im               instance.Manager
	joined           *atomic.Bool
	configs          map[string]struct{}

	store configstore.Store
	ring  ReadRing

	cancel context.CancelFunc
	exited chan bool

	closeDependencies func() error

	defaultRemoteWrite []*instance.RemoteWriteConfig
}

// New creates a new HA scraping service instance.
func New(reg prometheus.Registerer, cfg Config, globalConfig *config.GlobalConfig, clientConfig client.Config, logger log.Logger, im instance.Manager, defaultRemoteWrite []*instance.RemoteWriteConfig) (*Server, error) {
	// Force ReplicationFactor to be 1, since replication isn't supported for the
	// scraping service yet.
	cfg.Lifecycler.RingConfig.ReplicationFactor = 1

	store, err := configstore.NewRemote(logger, reg, cfg.KVStore)
	if err != nil {
		return nil, err
	}

	r, err := newRing(cfg.Lifecycler.RingConfig, "agent_viewer", "agent", reg)
	if err != nil {
		return nil, err
	}
	if err := reg.Register(r); err != nil {
		return nil, fmt.Errorf("failed to register Agent ring metrics: %w", err)
	}
	if err := services.StartAndAwaitRunning(context.Background(), r); err != nil {
		return nil, err
	}

	lazy := &lazyTransferer{}

	// TODO(rfratto): switching to a BasicLifecycler would be nice here, it'd allow
	// the joining/leaving process to include waiting for resharding.
	lc, err := ring.NewLifecycler(cfg.Lifecycler, lazy, "agent", "agent", true, reg)
	if err != nil {
		return nil, err
	}
	if err := services.StartAndAwaitRunning(context.Background(), lc); err != nil {
		return nil, err
	}

	logger = log.With(logger, "component", "ha")
	s := newServer(
		reg,
		cfg,
		globalConfig,
		clientConfig,
		logger,

		im,

		lc.Addr,
		r,
		store,

		// The lifecycler must stop first since the shutdown process depends on the
		// ring still polling.
		stopServices(lc, r),

		defaultRemoteWrite,
	)

	lazy.inner = s
	return s, nil
}

// newRing creates a new Cortex Ring that ignores unhealthy nodes.
func newRing(cfg ring.Config, name, key string, reg prometheus.Registerer) (*ring.Ring, error) {
	codec := ring.GetCodec()
	store, err := kv.NewClient(
		cfg.KVStore,
		codec,
		kv.RegistererWithKVName(reg, name+"-ring"),
	)
	if err != nil {
		return nil, err
	}
	return ring.NewWithStoreClientAndStrategy(cfg, name, key, store, ring.NewIgnoreUnhealthyInstancesReplicationStrategy())
}

// newServer creates a new Server. Abstracted from New for testing.
func newServer(reg prometheus.Registerer, cfg Config, globalCfg *config.GlobalConfig, clientCfg client.Config, log log.Logger, im instance.Manager, addr string, r ReadRing, store configstore.Store, stopFunc func() error, defaultRemoteWrite []*instance.RemoteWriteConfig) *Server {
	ctx, cancel := context.WithCancel(context.Background())

	s := &Server{
		cfg:          cfg,
		globalConfig: globalCfg,
		clientConfig: clientCfg,
		reg:          reg,
		logger:       log,
		addr:         addr,

		im:      im,
		joined:  atomic.NewBool(false),
		configs: make(map[string]struct{}),

		store: store,
		ring:  r,

		cancel:            cancel,
		exited:            make(chan bool),
		closeDependencies: stopFunc,

		defaultRemoteWrite: defaultRemoteWrite,
	}

	go s.run(ctx)
	return s
}

func (s *Server) run(ctx context.Context) {
	defer close(s.exited)

	// Perform join operations. Joining can't fail; any failed operation
	// performed will eventually correct itself due to how all Agents in the
	// cluster reshard themselves every reshard_interval.
	s.join(ctx)
	s.joined.Store(true)

	wg := sync.WaitGroup{}
	wg.Add(2)

	go func() {
		s.watchKV(ctx)
		wg.Done()
		level.Info(s.logger).Log("msg", "watch kv store process exited")
	}()

	go func() {
		s.reshardLoop(ctx)
		wg.Done()
		level.Info(s.logger).Log("msg", "reshard loop process exited")
	}()

	wg.Wait()
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
	level.Info(s.logger).Log("msg", "retrieving agents in cluster for cluster-wide reshard")

	var (
		rs  ring.ReplicationSet
		err error
	)

	backoff := util.NewBackoff(ctx, backoffConfig)
	for backoff.Ongoing() {
		rs, err = s.ring.GetAllHealthy(ring.Read)
		if err == nil {
			break
		}

		level.Warn(s.logger).Log("msg", "could not get agents in cluster", "err", err)
		backoff.Wait()
	}
	if err := backoff.Err(); err != nil {
		return err
	}

	_, err = rs.Do(ctx, time.Millisecond*250, func(ctx context.Context, desc *ring.InstanceDesc) (interface{}, error) {
		// Skip over ourselves; we'll reshard locally after this process finishes.
		if desc.Addr == s.addr {
			return nil, nil
		}

		ctx = user.InjectOrgID(ctx, "fake")
		return nil, s.notifyReshard(ctx, desc)
	})
	return err
}

func (s *Server) notifyReshard(ctx context.Context, desc *ring.InstanceDesc) error {
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

		owned, err := s.owns(ev.Key)
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
	level.Info(s.logger).Log("msg", "resharding agent on interval", "interval", s.cfg.ReshardInterval)
	t := time.NewTicker(s.cfg.ReshardInterval)
	defer t.Stop()

	for {
		select {
		case <-t.C:
			s.localReshard(ctx)
		case <-ctx.Done():
			return
		case <-s.exited:
			return
		}
	}
}

func (s *Server) localReshard(ctx context.Context) {
	if s.cfg.ReshardTimeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, s.cfg.ReshardTimeout)
		defer cancel()
	}

	level.Info(s.logger).Log("msg", "resharding agent")
	_, err := s.Reshard(ctx, &agentproto.ReshardRequest{})
	if err != nil {
		level.Error(s.logger).Log("msg", "resharding failed", "err", err)
	}
}

// WireAPI injects routes into the provided mux router for the config
// management API.
func (s *Server) WireAPI(r *mux.Router) {
	storeAPI := configstore.NewAPI(s.logger, s.store, func(c *instance.Config) error {
		return c.ApplyDefaults(s.globalConfig, s.defaultRemoteWrite)
	})
	s.reg.MustRegister(storeAPI)
	storeAPI.WireAPI(r)

	// Debug ring page
	r.Handle("/debug/ring", s.ring)
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
	if s.cfg.ReshardTimeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, s.cfg.ReshardTimeout)
		defer cancel()
	}

	level.Info(s.logger).Log("msg", "leaving cluster, starting cluster-wide reshard process")
	return s.waitNotifyReshard(ctx)
}

// Stop stops the HA server and its dependencies.
func (s *Server) Stop() error {
	// Close the loop and wait for it to stop.
	s.cancel()
	<-s.exited

	err := s.closeDependencies()

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

// stopServices stops services in argument-call order. Blocks until all services
// have stopped.
func stopServices(svcs ...services.Service) func() error {
	return func() error {
		var firstErr error
		for _, s := range svcs {
			err := services.StopAndAwaitTerminated(context.Background(), s)
			if err != nil && firstErr == nil {
				firstErr = err
			}
		}
		return firstErr
	}
}
