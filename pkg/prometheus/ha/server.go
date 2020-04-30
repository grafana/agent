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
	"sync"
	"time"

	"github.com/cortexproject/cortex/pkg/ring"
	"github.com/cortexproject/cortex/pkg/ring/kv"
	"github.com/cortexproject/cortex/pkg/util"
	"github.com/cortexproject/cortex/pkg/util/services"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/grafana/agent/pkg/agentproto"
	"github.com/grafana/agent/pkg/prometheus/ha/client"
	"github.com/grafana/agent/pkg/prometheus/instance"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
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

		exited: make(chan bool),
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

	ctx, cancel := context.WithCancel(context.Background())
	s.cancel = cancel

	go s.run(ctx)
	return s, nil
}

func (s *Server) run(ctx context.Context) {
	defer close(s.exited)

	if err := s.join(ctx); err != nil {
		level.Error(s.logger).Log("msg", "exiting scraping service loop due to error", "err", err)
		return
	}

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

func (s *Server) join(ctx context.Context) error {
	if err := s.waitNotifyReshard(ctx); err != nil {
		level.Error(s.logger).Log("msg", "could not run cluster-wide reshard", "err", err)
		return fmt.Errorf("could not complete join: %w", err)
	}

	level.Info(s.logger).Log("msg", "cluster-wide reshard finished. running local reshard")
	if _, err := s.Reshard(ctx, &agentproto.ReshardRequest{}); err != nil {
		level.Error(s.logger).Log("msg", "failed running local reshard", "err", err)
		return fmt.Errorf("could not complete join: %w", err)
	}

	return nil
}

func (s *Server) waitNotifyReshard(ctx context.Context) error {
	level.Info(s.logger).Log("msg", "retrieving agents in cluster for cluster-wide reshard")

	var (
		rs  ring.ReplicationSet
		err error
	)

	backoff := util.NewBackoff(ctx, backoffConfig)
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

	_, err = rs.Do(ctx, time.Millisecond*250, func(desc *ring.IngesterDesc) (interface{}, error) {
		// Skip over ourselves; we'll reshard locally after this process finishes.
		if desc.Addr == s.lc.Addr {
			return nil, nil
		}

		ctx = user.InjectOrgID(ctx, "fake")
		return nil, s.notifyReshard(ctx, desc)
	})
	return err
}

func (s *Server) notifyReshard(ctx context.Context, desc *ring.IngesterDesc) error {
	cli, err := client.New(s.clientConfig, desc.Addr)
	if err != nil {
		return err
	}
	defer cli.Close()

	backoff := util.NewBackoff(context.Background(), backoffConfig)
	for backoff.Ongoing() {
		_, err := cli.Reshard(ctx, &agentproto.ReshardRequest{})
		if err == nil {
			break
		}

		level.Warn(s.logger).Log("failed to tell remote agent to reshard", "err", err, "addr", desc.Addr)
		backoff.Wait()
	}

	return backoff.Err()
}

func (s *Server) watchKV(ctx context.Context) {
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
}

func (s *Server) reshardLoop(ctx context.Context) {
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
	level.Info(s.logger).Log("msg", "leaving cluster, starting cluster-wide reshard process")
	return s.waitNotifyReshard(ctx)
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
