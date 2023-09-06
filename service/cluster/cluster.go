// Package cluster implements the cluster service for Flow, where multiple
// instances of Flow connect to each other for work distribution.
package cluster

import (
	"context"
	"crypto/tls"
	"fmt"
	"math/rand"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/grafana/agent/component"
	"github.com/grafana/agent/service"
	http_service "github.com/grafana/agent/service/http"
	"github.com/grafana/ckit"
	"github.com/grafana/ckit/peer"
	"github.com/grafana/ckit/shard"
	"github.com/prometheus/client_golang/prometheus"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"golang.org/x/net/http2"
)

// tokensPerNode is used to decide how many tokens each node should be given in
// the hash ring. All nodes must use the same value, otherwise they will have
// different views of the ring and assign work differently.
//
// Using 512 tokens strikes a good balance between distribution accuracy and
// memory consumption. A cluster of 1,000 nodes with 512 tokens per node
// requires 12MB for the hash ring.
//
// Distribution accuracy measures how close a node was to being responsible for
// exactly 1/N keys during simulation. Simulation tests used a cluster of 10
// nodes and hashing 100,000 random keys:
//
//	512 tokens per node: min 96.1%, median 99.9%, max 103.2% (stddev: 197.9 hashes)
const tokensPerNode = 512

// ServiceName defines the name used for the cluster service.
const ServiceName = "cluster"

// Options are used to configure the cluster service. Options are constant for
// the lifetime of the cluster service.
type Options struct {
	Log     log.Logger            // Where to send logs to.
	Metrics prometheus.Registerer // Where to send metrics to.
	Tracer  trace.TracerProvider  // Where to send traces.

	// EnableClustering toggles clustering as a whole. When EnableClustering is
	// false, the instance of Flow acts as a single-node cluster and it is not
	// possible for other nodes to join the cluster.
	EnableClustering bool

	NodeName            string        // Name to use for this node in the cluster.
	AdvertiseAddress    string        // Address to advertise to other nodes in the cluster.
	RejoinInterval      time.Duration // How frequently to rejoin the cluster to address split brain issues.
	ClusterMaxJoinPeers int           // Number of initial peers to join from the discovered set.
	ClusterName         string        // Name to prevent nodes without this identifier from joining the cluster.

	// Function to discover peers to join. If this function is nil or returns an
	// empty slice, no peers will be joined.
	DiscoverPeers func() ([]string, error)
}

// Service is the cluster service.
type Service struct {
	log    log.Logger
	tracer trace.TracerProvider
	opts   Options

	sharder shard.Sharder
	node    *ckit.Node
	randGen *rand.Rand
}

var (
	_ service.Service             = (*Service)(nil)
	_ http_service.ServiceHandler = (*Service)(nil)
)

// New returns a new, unstarted instance of the cluster service.
func New(opts Options) (*Service, error) {
	var (
		l = opts.Log
		t = opts.Tracer
	)
	if l == nil {
		l = log.NewNopLogger()
	}
	if t == nil {
		t = trace.NewNoopTracerProvider()
	}

	ckitConfig := ckit.Config{
		Name:          opts.NodeName,
		AdvertiseAddr: opts.AdvertiseAddress,
		Log:           l,
		Sharder:       shard.Ring(tokensPerNode),
		Label:         opts.ClusterName,
	}

	httpClient := &http.Client{
		Transport: &http2.Transport{
			AllowHTTP: true,
			DialTLSContext: func(ctx context.Context, network, addr string, _ *tls.Config) (net.Conn, error) {
				// Set a maximum timeout for establishing the connection. If our
				// context has a deadline earlier than our timeout, we shrink the
				// timeout to it.
				//
				// TODO(rfratto): consider making the max timeout configurable.
				timeout := 30 * time.Second
				if dur, ok := deadlineDuration(ctx); ok && dur < timeout {
					timeout = dur
				}

				return net.DialTimeout(network, addr, timeout)
			},
		},
	}

	node, err := ckit.NewNode(httpClient, ckitConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create cluster node: %w", err)
	}
	if opts.EnableClustering && opts.Metrics != nil {
		if err := opts.Metrics.Register(node.Metrics()); err != nil {
			return nil, fmt.Errorf("failed to register metrics: %w", err)
		}
	}

	return &Service{
		log:    l,
		tracer: t,
		opts:   opts,

		sharder: ckitConfig.Sharder,
		node:    node,
		randGen: rand.New(rand.NewSource(time.Now().UnixNano())),
	}, nil
}

func deadlineDuration(ctx context.Context) (d time.Duration, ok bool) {
	if t, ok := ctx.Deadline(); ok {
		return time.Until(t), true
	}
	return 0, false
}

// Definition returns the definition of the cluster service.
func (s *Service) Definition() service.Definition {
	return service.Definition{
		Name:       ServiceName,
		ConfigType: nil, // cluster does not accept configuration.
		DependsOn: []string{
			// Cluster depends on the HTTP service to work properly.
			http_service.ServiceName,
		},
	}
}

// ServiceHandler returns the service handler for the clustering service. The
// resulting handler always returns 404 when clustering is disabled.
func (s *Service) ServiceHandler(host service.Host) (base string, handler http.Handler) {
	base, handler = s.node.Handler()

	if !s.opts.EnableClustering {
		handler = http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			http.Error(w, "clustering is disabled", http.StatusNotFound)
		})
	}

	return base, handler
}

// ChangeState changes the state of the service. If clustering is enabled,
// ChangeState will block until the state change has been propagated to another
// node; cancel the current context to stop waiting. ChangeState fails if the
// current state cannot move to the provided targetState.
//
// Note that the state must be StateParticipant to receive writes.
func (s *Service) ChangeState(ctx context.Context, targetState peer.State) error {
	return s.node.ChangeState(ctx, targetState)
}

// Run starts the cluster service. It will run until the provided context is
// canceled or there is a fatal error.
func (s *Service) Run(ctx context.Context, host service.Host) error {
	// Stop the node on shutdown.
	defer s.stop()

	var wg sync.WaitGroup
	defer wg.Wait()

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	s.node.Observe(ckit.FuncObserver(func(peers []peer.Peer) (reregister bool) {
		if ctx.Err() != nil {
			// Unregister our observer if we exited.
			return false
		}

		tracer := s.tracer.Tracer("")
		spanCtx, span := tracer.Start(ctx, "NotifyClusterChange", trace.WithSpanKind(trace.SpanKindInternal))
		defer span.End()

		names := make([]string, len(peers))
		for i, p := range peers {
			names[i] = p.Name
		}
		level.Info(s.log).Log("msg", "peers changed", "new_peers", strings.Join(names, ","))

		// Notify dependant components about the clustering change.
		for _, consumer := range host.GetServiceConsumers(ServiceName) {
			if ctx.Err() != nil {
				// Stop early if we exited so we don't do unnecessary work notifying
				// consumers that do not need to be notified.
				break
			}

			clusterComponent, ok := consumer.Value.(Component)
			if !ok {
				continue
			}

			_, span := tracer.Start(spanCtx, "NotifyClusterChange", trace.WithSpanKind(trace.SpanKindInternal))
			span.SetAttributes(attribute.String("consumer_id", consumer.ID))

			clusterComponent.NotifyClusterChange()

			span.End()
		}

		return true
	}))

	peers, err := s.getPeers()
	if err != nil {
		return fmt.Errorf("failed to get peers to join: %w", err)
	}

	level.Info(s.log).Log("msg", "starting cluster node", "peers", strings.Join(peers, ","),
		"advertise_addr", s.opts.AdvertiseAddress)

	if err := s.node.Start(peers); err != nil {
		level.Warn(s.log).Log("msg", "failed to connect to peers; bootstrapping a new cluster", "err", err)

		err := s.node.Start(nil)
		if err != nil {
			return fmt.Errorf("failed to bootstrap a fresh cluster with no peers: %w", err)
		}
	}

	if s.opts.EnableClustering && s.opts.RejoinInterval > 0 {
		wg.Add(1)

		go func() {
			defer wg.Done()

			t := time.NewTicker(s.opts.RejoinInterval)
			defer t.Stop()

			for {
				select {
				case <-ctx.Done():
					return

				case <-t.C:
					peers, err := s.getPeers()
					if err != nil {
						level.Warn(s.log).Log("msg", "failed to refresh list of peers", "err", err)
						continue
					}

					level.Info(s.log).Log("msg", "rejoining peers", "peers", strings.Join(peers, ","))
					if err := s.node.Start(peers); err != nil {
						level.Error(s.log).Log("msg", "failed to rejoin list of peers", "err", err)
						continue
					}
				}
			}
		}()
	}

	<-ctx.Done()
	return nil
}

func (s *Service) getPeers() ([]string, error) {
	if !s.opts.EnableClustering || s.opts.DiscoverPeers == nil {
		return nil, nil
	}

	peers, err := s.opts.DiscoverPeers()
	if err != nil {
		return nil, err
	}

	// Here we return the entire list because we can't take a subset.
	if s.opts.ClusterMaxJoinPeers == 0 || len(peers) < s.opts.ClusterMaxJoinPeers {
		return peers, nil
	}

	// We shuffle the list and return only a subset of the peers.
	s.randGen.Shuffle(len(peers), func(i, j int) {
		peers[i], peers[j] = peers[j], peers[i]
	})
	return peers[:s.opts.ClusterMaxJoinPeers], nil
}

func (s *Service) stop() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// The node is going away. We move to the Terminating state to signal
	// that we should not be owners for write hashing operations anymore.
	//
	// TODO(rfratto): should we enter terminating state earlier to allow for
	// some kind of hand-off between components?
	if err := s.node.ChangeState(ctx, peer.StateTerminating); err != nil {
		level.Error(s.log).Log("msg", "failed to change state to Terminating", "err", err)
	}

	if err := s.node.Stop(); err != nil {
		level.Error(s.log).Log("msg", "failed to gracefully stop node", "err", err)
	}
}

// Update implements [service.Service]. It returns an error since the cluster
// service does not support runtime configuration.
func (s *Service) Update(newConfig any) error {
	return fmt.Errorf("cluster service does not support configuration")
}

// Data returns an instance of [Cluster].
func (s *Service) Data() any {
	return &sharderCluster{sharder: s.sharder}
}

// Component is a Flow component which subscribes to clustering updates.
type Component interface {
	component.Component

	// NotifyClusterChange notifies the component that the state of the cluster
	// has changed.
	//
	// Implementations should ignore calls to this method if they are configured
	// to not utilize clustering.
	NotifyClusterChange()
}

// Cluster is a read-only view of a cluster.
type Cluster interface {
	// Lookup determines the set of replicationFactor owners for a given key.
	// peer.Peer.Self can be used to determine if the local node is the owner,
	// allowing for short-circuiting logic to connect directly to the local node
	// instead of using the network.
	//
	// Callers can use github.com/grafana/ckit/shard.StringKey or
	// shard.NewKeyBuilder to create a key.
	Lookup(key shard.Key, replicationFactor int, op shard.Op) ([]peer.Peer, error)

	// Peers returns the current set of peers for a Node.
	Peers() []peer.Peer
}

// sharderCluster shims an implementation of [shard.Sharder] to [Cluster] which
// removes the ability to change peers.
type sharderCluster struct{ sharder shard.Sharder }

var _ Cluster = (*sharderCluster)(nil)

func (sc *sharderCluster) Lookup(key shard.Key, replicationFactor int, op shard.Op) ([]peer.Peer, error) {
	return sc.sharder.Lookup(key, replicationFactor, op)
}

func (sc *sharderCluster) Peers() []peer.Peer {
	return sc.sharder.Peers()
}
