package cluster

import (
	"context"
	"fmt"
	"sync"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/google/go-cmp/cmp"
	"github.com/gorilla/mux"
	"github.com/grafana/agent/pkg/agentproto"
	"github.com/grafana/agent/pkg/prom/instance"
	"github.com/grafana/agent/pkg/prom/instance/configstore"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/prometheus/config"
	"google.golang.org/grpc"
)

// Cluster connects an Agent to other Agents and allows them to distribute
// workload.
type Cluster struct {
	mut sync.RWMutex

	log log.Logger

	cfg                Config
	global             *config.GlobalConfig
	defaultRemoteWrite []*instance.RemoteWriteConfig

	//
	// Internally, Cluster glues together four separate pieces of logic.
	// See comments below to get an understanding of what is going on.
	//

	// node manages membership in the cluster and performs cluster-wide reshards.
	node *node

	// store connects to a configstore for changes. storeAPI is an HTTP API for it.
	store    *configstore.Remote
	storeAPI *configstore.API

	// watcher watches the store and applies changes to an instance.Manager,
	// triggering metrics to be collected and sent. configWatcher also does a
	// complete refresh of its state on an interval.
	watcher *configWatcher
}

// New creates a new Cluster.
func New(
	l log.Logger,
	reg prometheus.Registerer,
	cfg Config,
	global *config.GlobalConfig,
	defaultRemoteWrite []*instance.RemoteWriteConfig,
	im instance.Manager,
) (*Cluster, error) {

	var (
		c = &Cluster{
			log: l,

			cfg:                cfg,
			global:             global,
			defaultRemoteWrite: defaultRemoteWrite,
		}
		err error
	)

	// Hold the lock for the initialization. This is necessary since newNode will
	// eventually call Reshard, and we want c.watcher to be initialized when that
	// happens.
	c.mut.Lock()
	defer c.mut.Unlock()

	c.node, err = newNode(reg, l, cfg, c)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize node membership: %w", err)
	}

	c.store, err = configstore.NewRemote(l, reg, cfg.KVStore)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize configstore: %w", err)
	}
	c.storeAPI = configstore.NewAPI(l, c.store, c.Validate)
	reg.MustRegister(c.storeAPI)

	c.watcher, err = newConfigWatcher(l, cfg, c.store, im, c.node.Owns, c.Validate)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize configwatcher: %w", err)
	}

	// NOTE(rfratto): ApplyConfig isn't necessary for the initialization but must
	// be called for any changes to the configuration.
	return c, nil
}

// Reshard implements agentproto.ScrapingServiceServer, and syncs the state of
// configs with the configstore.
func (c *Cluster) Reshard(ctx context.Context, _ *agentproto.ReshardRequest) (*empty.Empty, error) {
	err := c.watcher.Refresh(ctx)
	if err != nil {
		level.Error(c.log).Log("msg", "failed to perform local reshard", "err", err)
	}
	return &empty.Empty{}, err
}

// Validate will validate the incoming Config and mutate it to apply defaults.
func (c *Cluster) Validate(cfg *instance.Config) error {
	c.mut.RLock()
	defer c.mut.RUnlock()

	if err := cfg.ApplyDefaults(c.global, c.defaultRemoteWrite); err != nil {
		return fmt.Errorf("failed to apply defaults to %q: %w", cfg.Name, err)
	}

	return nil
}

// ApplyConfig applies configuration changes to Cluster.
func (c *Cluster) ApplyConfig(
	cfg Config,
	global *config.GlobalConfig,
	defaultRemoteWrite []*instance.RemoteWriteConfig,
) error {
	c.mut.Lock()
	defer c.mut.Unlock()

	if cmp.Equal(c.cfg, cfg) &&
		cmp.Equal(c.global, global) &&
		cmp.Equal(c.defaultRemoteWrite, defaultRemoteWrite) {
		return nil
	}

	if err := c.node.ApplyConfig(cfg); err != nil {
		return fmt.Errorf("failed to apply config to node membership: %w", err)
	}

	if err := c.store.ApplyConfig(cfg.Lifecycler.RingConfig.KVStore); err != nil {
		return fmt.Errorf("failed to apply config to config store: %w", err)
	}

	if err := c.watcher.ApplyConfig(cfg); err != nil {
		return fmt.Errorf("failed to apply config to watcher: %w", err)
	}

	c.cfg = cfg
	c.global = global
	c.defaultRemoteWrite = defaultRemoteWrite

	// Force a refresh so all the configs get updated with new defaults.
	level.Info(c.log).Log("msg", "cluster config changed, refreshing from configstore in background")
	go func() {
		ctx := context.Background()
		if c.cfg.ReshardTimeout > 0 {
			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(ctx, c.cfg.ReshardTimeout)
			defer cancel()
		}
		err := c.watcher.Refresh(ctx)
		if err != nil {
			level.Error(c.log).Log("msg", "failed to perform local reshard", "err", err)
		}
	}()

	return nil
}

// WireAPI injects routes into the provided mux router for the config
// management API.
func (c *Cluster) WireAPI(r *mux.Router) {
	c.storeAPI.WireAPI(r)
	c.node.WireAPI(r)
}

// WireGRPC injects gRPC server handlers into the provided gRPC server.
func (c *Cluster) WireGRPC(srv *grpc.Server) {
	agentproto.RegisterScrapingServiceServer(srv, c)
}
