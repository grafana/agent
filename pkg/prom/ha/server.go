// Package ha implements a high availability clustering mode for the agent. It
// is also referred to as the "scraping service" mode, as this is a fairly
// accurate description of what it does: a series of configs are stored in a
// KV store and a cluster of agents pulls configs from the store and shards
// them amongst the cluster, thereby distributing scraping load.
package ha

import (
	"context"
	"flag"
	"net/http"
	"sync"
	"time"

	"github.com/cortexproject/cortex/pkg/ring"
	"github.com/cortexproject/cortex/pkg/ring/kv"
	"github.com/cortexproject/cortex/pkg/util"
	"github.com/cortexproject/cortex/pkg/util/services"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/grafana/agent/pkg/agentproto"
	"github.com/grafana/agent/pkg/prom/ha/client"
	"github.com/grafana/agent/pkg/prom/instance"
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
	c.KVStore.RegisterFlagsWithPrefix(prefix+"config-store.", "configurations/", f)
	c.Lifecycler.RegisterFlagsWithPrefix(prefix, f)
}

// readRing is implemented by ring.Ring. Brought out to a minimal interface for
// testing.
type readRing interface {
	http.Handler

	Get(key uint32, op ring.Operation, buf []ring.IngesterDesc) (ring.ReplicationSet, error)
	GetAll() (ring.ReplicationSet, error)
}

// Server implements the HA scraping service.
type Server struct {
	cfg          Config
	clientConfig client.Config
	globalConfig *config.GlobalConfig
	logger       log.Logger
	addr         string

	configManagerMut sync.Mutex
	im               instance.Manager
	joined           *atomic.Bool

	kv   kv.Client
	ring readRing

	cancel context.CancelFunc
	exited chan bool

	closeDependencies func() error
}

// New creates a new HA scraping service instance.
func New(cfg Config, globalConfig *config.GlobalConfig, clientConfig client.Config, logger log.Logger, im instance.Manager) (*Server, error) {
	// Force ReplicationFactor to be 1, since replication isn't supported for the
	// scraping service yet.
	cfg.Lifecycler.RingConfig.ReplicationFactor = 1

	kvClient, err := kv.NewClient(cfg.KVStore, GetCodec())
	if err != nil {
		return nil, err
	}

	r, err := ring.New(cfg.Lifecycler.RingConfig, "agent_viewer", "agent")
	if err != nil {
		return nil, err
	}
	if err := services.StartAndAwaitRunning(context.Background(), r); err != nil {
		return nil, err
	}

	lazy := &lazyTransferer{}

	// TODO(rfratto): switching to a BasicLifecycler would be nice here, it'd allow
	// the joining/leaving process to include waiting for resharding.
	lc, err := ring.NewLifecycler(cfg.Lifecycler, lazy, "agent", "agent", true)
	if err != nil {
		return nil, err
	}
	if err := services.StartAndAwaitRunning(context.Background(), lc); err != nil {
		return nil, err
	}

	logger = log.With(logger, "component", "ha")
	s := newServer(
		cfg,
		globalConfig,
		clientConfig,
		logger,

		NewShardingInstanceManager(logger, im, r, lc.Addr),

		lc.Addr,
		r,
		kvClient,

		// The lifecycler must stop first since the shutdown process depends on the
		// ring still polling.
		stopServices(lc, r),
	)

	lazy.inner = s
	return s, nil
}

// newServer creates a new Server. Abstracted from New for testing.
func newServer(cfg Config, globalCfg *config.GlobalConfig, clientCfg client.Config, log log.Logger, im instance.Manager, addr string, r readRing, kv kv.Client, stopFunc func() error) *Server {
	ctx, cancel := context.WithCancel(context.Background())

	s := &Server{
		cfg:          cfg,
		globalConfig: globalCfg,
		clientConfig: clientCfg,
		logger:       log,
		addr:         addr,

		im:     im,
		joined: atomic.NewBool(false),

		kv:   kv,
		ring: r,

		cancel:            cancel,
		exited:            make(chan bool),
		closeDependencies: stopFunc,
	}

	go s.run(ctx)
	return s
}

func (s *Server) run(ctx context.Context) {
	defer close(s.exited)

	if err := s.join(ctx); err != nil {
		level.Error(s.logger).Log("msg", "could not complete join, stopping HA server", "err", err)
		return
	}

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

func (s *Server) join(ctx context.Context) error {
	if err := s.waitNotifyReshard(ctx); err != nil {
		level.Error(s.logger).Log("msg", "could not run cluster-wide reshard", "err", err)
		return err
	}

	level.Info(s.logger).Log("msg", "cluster-wide reshard finished. running local reshard")
	if _, err := s.Reshard(ctx, &agentproto.ReshardRequest{}); err != nil {
		level.Error(s.logger).Log("msg", "failed running local reshard", "err", err)
		return err
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
		if desc.Addr == s.addr {
			return nil, nil
		}

		ctx := user.InjectOrgID(ctx, "fake")
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

	backoff := util.NewBackoff(ctx, backoffConfig)
	for backoff.Ongoing() {
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

	s.kv.WatchPrefix(ctx, "", func(key string, v interface{}) bool {
		s.configManagerMut.Lock()
		defer s.configManagerMut.Unlock()

		if ctx.Err() != nil {
			return false
		}

		if v == nil {
			if err := s.im.DeleteConfig(key); err != nil {
				level.Error(s.logger).Log("msg", "failed to delete config", "name", key, "err", err)
			}
			return true
		}

		cfg := v.(*instance.Config)
		if err := s.im.ApplyConfig(*cfg); err != nil {
			level.Error(s.logger).Log("msg", "failed to apply config, will retry on next reshard", "name", key, "err", err)
		}
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

	return s.closeDependencies()
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
