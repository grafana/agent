package ha

import (
	"context"
	"fmt"
	"hash/fnv"
	"net/http"
	"sync"
	"time"

	"github.com/cortexproject/cortex/pkg/ring"
	"github.com/cortexproject/cortex/pkg/ring/kv"
	"github.com/cortexproject/cortex/pkg/util"
	"github.com/cortexproject/cortex/pkg/util/services"
	"github.com/go-kit/kit/log"
	"github.com/google/go-cmp/cmp"
	"github.com/gorilla/mux"
	prom_util "github.com/grafana/agent/pkg/util"
	"github.com/prometheus/client_golang/prometheus"
)

// node managements membership within a ring.
type node struct {
	reg    *prom_util.Unregisterer
	server *Server

	mut  sync.Mutex
	cfg  Config
	ring *ring.Ring
	lc   *ring.Lifecycler
}

func newNode(reg prometheus.Registerer, log log.Logger, cfg Config, s *Server) (*node, error) {
	n := &node{
		reg:    prom_util.WrapWithUnregisterer(reg),
		server: s,
	}
	if err := n.ApplyConfig(cfg); err != nil {
		return nil, err
	}
	return n, nil
}

func (n *node) ApplyConfig(cfg Config) error {
	n.mut.Lock()
	defer n.mut.Unlock()

	if cmp.Equal(n.cfg, cfg) {
		return nil
	}

	n.reg.UnregisterAll()

	if n.ring != nil {
		err := services.StopAndAwaitTerminated(context.Background(), n.ring)
		if err != nil {
			return err
		}
		n.ring = nil
	}

	if n.lc != nil {
		err := services.StopAndAwaitTerminated(context.Background(), n.lc)
		if err != nil {
			return err
		}
		n.lc = nil
	}

	if !cfg.Enabled {
		n.cfg = cfg
		return nil
	}

	r, err := newRing(cfg.Lifecycler.RingConfig, "agent_viewer", "agent", n.reg)
	if err != nil {
		return err
	}
	if err := n.reg.Register(r); err != nil {
		return fmt.Errorf("failed to register Agent ring metrics: %w", err)
	}
	if err := services.StartAndAwaitRunning(context.Background(), r); err != nil {
		return fmt.Errorf("failed to start ring: %w", err)
	}
	n.ring = r

	lc, err := ring.NewLifecycler(cfg.Lifecycler, n.server, "agent", "agent", true, n.reg)
	if err != nil {
		return err
	}
	if err := services.StartAndAwaitRunning(context.Background(), lc); err != nil {
		return err
	}
	n.lc = lc

	n.cfg = cfg
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

func (n *node) WireAPI(r *mux.Router) {
	r.HandleFunc("/debug/ring", func(rw http.ResponseWriter, r *http.Request) {
		n.mut.Lock()
		defer n.mut.Unlock()

		if n.ring == nil {
			http.Error(rw, "Agent not connected to cluster", http.StatusPreconditionFailed)
		}

		n.ring.ServeHTTP(rw, r)
	})
}

// Owns returns true if this node owns the key specified by key.
// The key will be hashed using a 32-bit FNV-1 hash.
func (n *node) Owns(key string) (bool, error) {
	n.mut.Lock()
	defer n.mut.Unlock()

	rs, err := n.ring.Get(keyHash(key), ring.Write, nil, nil, nil)
	if err != nil {
		return false, err
	}
	for _, r := range rs.Ingesters {
		if r.Addr == n.lc.Addr {
			return true, nil
		}
	}
	return false, nil
}

// IteratePeers will call f for each peer (which is not the current node) in the cluster.
func (n *node) IteratePeers(ctx context.Context, f func(ctx context.Context, desc *ring.InstanceDesc) error) error {
	var (
		rs  ring.ReplicationSet
		err error
	)

	n.mut.Lock()

	backoff := util.NewBackoff(ctx, backoffConfig)
	for backoff.Ongoing() {
		rs, err = n.ring.GetAllHealthy(ring.Read)
		if err == nil {
			break
		}
		backoff.Wait()
	}

	n.mut.Unlock()

	if err := backoff.Err(); err != nil {
		return err
	}

	_, err = rs.Do(ctx, time.Millisecond*250, func(c context.Context, id *ring.InstanceDesc) (interface{}, error) {
		n.mut.Lock()
		localAddr := n.lc.Addr
		n.mut.Unlock()

		if id.Addr == localAddr {
			return nil, nil
		}

		return nil, f(ctx, id)
	})
	return err
}

func keyHash(key string) uint32 {
	h := fnv.New32()
	_, _ = h.Write([]byte(key))
	return h.Sum32()
}

// Stop stops the associated services with the node.
func (n *node) Stop() error {
	n.mut.Lock()
	defer n.mut.Unlock()

	var firstErr error
	for _, svc := range []services.Service{n.lc, n.ring} {
		if svc == nil {
			continue
		}

		err := services.StopAndAwaitTerminated(context.Background(), svc)
		if err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}
