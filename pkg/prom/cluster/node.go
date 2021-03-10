package cluster

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/cortexproject/cortex/pkg/ring"
	"github.com/cortexproject/cortex/pkg/ring/kv"
	cortex_util "github.com/cortexproject/cortex/pkg/util"
	"github.com/cortexproject/cortex/pkg/util/services"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/google/go-cmp/cmp"
	pb "github.com/grafana/agent/pkg/agentproto"
	"github.com/grafana/agent/pkg/prom/ha/client"
	"github.com/grafana/agent/pkg/util"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/weaveworks/common/user"
)

var backoffConfig = cortex_util.BackoffConfig{
	MinBackoff: time.Second,
	MaxBackoff: 2 * time.Minute,
	MaxRetries: 10,
}

// node manages membership within a ring. when a node joins or leaves the ring,
// it will inform other nodes to reshard their workloads. After a node joins
// the ring, it will inform the local service to reshard.
type node struct {
	log log.Logger
	reg *util.Unregisterer
	srv pb.ScrapingServiceServer

	mut  sync.RWMutex
	cfg  Config
	ring *ring.Ring
	lc   *ring.Lifecycler

	exit   chan struct{}
	reload chan struct{}
}

// newNode creates a new node and registers it to the ring.
func newNode(reg prometheus.Registerer, log log.Logger, cfg Config, s pb.ScrapingServiceServer) (*node, error) {
	n := &node{
		reg: util.WrapWithUnregisterer(reg),
		srv: s,

		reload: make(chan struct{}, 1),
		exit:   make(chan struct{}),
	}
	if err := n.ApplyConfig(cfg); err != nil {
		return nil, err
	}
	go n.run()
	return n, nil
}

func (n *node) ApplyConfig(cfg Config) error {
	n.mut.Lock()
	defer n.mut.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Detect if the config changed.
	if cmp.Equal(n.cfg, cfg) {
		return nil
	}

	select {
	case <-n.exit:
		return fmt.Errorf("node stopped")
	default:
		// Node is running, can continue.
	}

	// Shut down old components before re-creating the updated ones.
	n.reg.UnregisterAll()

	if n.ring != nil {
		err := services.StopAndAwaitTerminated(ctx, n.ring)
		if err != nil {
			return fmt.Errorf("failed to stop ring: %w", err)
		}
		n.ring = nil
	}

	if n.lc != nil {
		err := services.StopAndAwaitTerminated(ctx, n.lc)
		if err != nil {
			return fmt.Errorf("failed to stop lifecycler: %w", err)
		}
		n.lc = nil
	}

	if !cfg.Enabled {
		n.cfg = cfg
		<-n.reload
		return nil
	}

	r, err := newRing(cfg.Lifecycler.RingConfig, "agent_viewer", "agent", n.reg)
	if err != nil {
		return fmt.Errorf("failed to create ring: %w", err)
	}
	if err := n.reg.Register(r); err != nil {
		return fmt.Errorf("failed to register ring metrics: %w", err)
	}
	if err := services.StartAndAwaitRunning(ctx, r); err != nil {
		return fmt.Errorf("failed to start ring: %w", err)
	}
	n.ring = r

	lc, err := ring.NewLifecycler(cfg.Lifecycler, n, "agent", "agent", true, n.reg)
	if err != nil {
		return fmt.Errorf("failed to create lifecycler: %w", err)
	}
	if err := services.StartAndAwaitRunning(ctx, lc); err != nil {
		r.StopAsync()
		return fmt.Errorf("failed to start lifecycler: %w", err)
	}
	n.lc = lc

	n.cfg = cfg

	<-n.reload
	return nil
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

// run waits for connection to the ring and kickstarts the join process.
func (n *node) run() {
	for range n.reload {
		if err := n.performClusterReshard(context.Background(), true); err != nil {
			level.Warn(n.log).Log("msg", "dynamic cluster reshard did not succeed", "err", err)
		}
	}

	level.Info(n.log).Log("msg", "node run loop exiting")
}

// performClusterReshard informs the cluster to immediately trigger a reshard
// of their workloads. if includeSelf is true, the server provided to newNode will
// also be informed. includeSelf should be true when joining the cluster, and false
// when leaving.
func (n *node) performClusterReshard(ctx context.Context, includeSelf bool) error {
	n.mut.RLock()
	defer n.mut.RUnlock()

	if n.cfg.ReshardTimeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, n.cfg.ReshardTimeout)
		defer cancel()
	}

	level.Info(n.log).Log("msg", "informing all nodes to reshard")

	var (
		rs  ring.ReplicationSet
		err error

		firstError error
	)

	backoff := cortex_util.NewBackoff(ctx, backoffConfig)
	for backoff.Ongoing() {
		rs, err = n.ring.GetAllHealthy(ring.Read)
		if err == nil {
			break
		}
		backoff.Wait()
	}
	if err := backoff.Err(); err != nil && firstError == nil {
		firstError = err
	}

	_, err = rs.Do(ctx, 250*time.Millisecond, func(c context.Context, id *ring.InstanceDesc) (interface{}, error) {
		// Skip over ourselves.
		if id.Addr == n.lc.Addr {
			return nil, nil
		}

		ctx = user.InjectOrgID(ctx, "fake")
		return nil, n.notifyReshard(ctx, id)
	})
	if err != nil && firstError == nil {
		firstError = err
	}

	if includeSelf {
		level.Info(n.log).Log("msg", "running local reshard")
		if _, err := n.srv.Reshard(ctx, &pb.ReshardRequest{}); err != nil {
			level.Warn(n.log).Log("msg", "dynamic local reshard did not succeed", "err", err)
		}
	}

	return firstError
}

// notifyReshard informs an individual node to reshard.
func (n *node) notifyReshard(ctx context.Context, id *ring.InstanceDesc) error {
	cli, err := client.New(n.cfg.Client, id.Addr)
	if err != nil {
		return err
	}
	defer cli.Close()

	level.Info(n.log).Log("msg", "attempting to notify remote agent to reshard", "addr", id.Addr)

	backoff := cortex_util.NewBackoff(ctx, backoffConfig)
	for backoff.Ongoing() {
		_, err := cli.Reshard(ctx, &pb.ReshardRequest{})
		if err == nil {
			break
		}

		level.Warn(n.log).Log("msg", "reshard notification attempt failed", "addr", id.Addr, "err", err, "attempt", backoff.NumRetries())
		backoff.Wait()
	}

	return backoff.Err()
}

// Stop stops the node and cancels it from running. The node cannot be used
// again once Stop is called.
func (n *node) Stop() error {
	n.mut.Lock()
	defer n.mut.Unlock()

	select {
	case <-n.exit:
		return fmt.Errorf("node stopped")
	default:
		// Node is running, can continue.
	}

	close(n.reload)
	close(n.exit)

	var firstError error

	if n.ring != nil {
		err := services.StopAndAwaitTerminated(context.Background(), n.ring)
		if err != nil && firstError == nil {
			firstError = fmt.Errorf("failed to stop ring: %w", err)
		}
		n.ring = nil
	}

	if n.lc != nil {
		err := services.StopAndAwaitTerminated(context.Background(), n.lc)
		if err != nil {
			firstError = fmt.Errorf("failed to stop lifecycler: %w", err)
		}
		n.lc = nil
	}

	return firstError
}

// Flush implements ring.FlushTransferer. It's a no-op.
func (n *node) Flush() {}

// TransferOut implements ring.FlushTransferer. It connects to all other healthy agents and
// tells them to reshard.
func (n *node) TransferOut(ctx context.Context) error {
	// Only inform other nodes in the cluster to reshard since we're leaving.
	return n.performClusterReshard(ctx, false)
}
