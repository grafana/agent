package cluster

import (
	"context"
	"fmt"
	"io"
	stdlog "log"
	"net"
	"net/http"
	"os"

	"github.com/go-kit/log"
	"github.com/grafana/dskit/flagext"
	"github.com/hashicorp/go-discover"
	"github.com/hashicorp/go-discover/provider/k8s"
	"github.com/rfratto/ckit"
	"github.com/rfratto/ckit/advertise"
	"github.com/rfratto/ckit/clientpool"
	"github.com/rfratto/ckit/peer"
	"github.com/rfratto/ckit/shard"
	"go.uber.org/atomic"
)

// extraDiscoverProviders used in tests.
var extraDiscoverProviders map[string]discover.Provider

// tokensPerNode is used to decide how many tokens each node should be given in
// the hash ring. All nodes must use the same value, otherwise they will have
// different views of the ring and assign work differently.
//
// Using 256 tokens strikes a good balance between distribution accuracy and
// memory consumption. A cluster of 1,000 nodes with 256 tokens per node
// requires 6MB for the hash ring, while 12MB is used for 512 tokens per node.
//
// Distribution accuracy measures how close a node was to being responsible for
// exactly 1/N keys during simulation. Simulation tests used a cluster of 10
// nodes and hashing 100,000 random keys:
//
//	256 tokens per node: min 94.0%, median 96.3%, max 115.3%
//	512 tokens per node: min 96.1%, median 99.9%, max 103.2%
//
// While 512 tokens per node is closer to perfect distribution, 256 tokens per
// node is good enough, optimizing for lower memory usage.
const tokensPerNode = 256

// GossipConfig controls clustering of Agents through gRPC-based gossip.
// GossipConfig cannot be changed at runtime.
type GossipConfig struct {
	// Name of the node within the cluster. Must be unique cluster-wide.
	NodeName string

	// host:port address to advertise to peers to connect to. When unset, the
	// first discovered IP from AdvertiseInterfaces will be used to find.
	AdvertiseAddr string

	// Slice of interface names to infer an advertise IP from. Must be set if
	// AdvertiseAddr is unset.
	AdvertiseInterfaces flagext.StringSlice

	// List of one or more host:port peer addresses to connect to. Mutually
	// exclusive with DiscoverPeers.
	//
	// If an agent connects to no peers, it will form a one-node cluster until a
	// peer connects to it explicitly.
	JoinPeers flagext.StringSlice

	// Discover peers to connect to using go-discover. Mutually exclusive with
	// JoinPeers.
	DiscoverPeers string

	// Client pool to use for connecting to peers.
	Pool *clientpool.Pool
}

// DefaultGossipConfig holds default GossipConfig options.
var DefaultGossipConfig = GossipConfig{
	AdvertiseInterfaces: advertise.DefaultInterfaces,
}

// ApplyDefaults mutates c with default settings applied. defaultPort is
// added as the default port for addresses that do not have port numbers
// assigned.
//
// An error will be returned if the configuration is invalid or if an error
// occurred while applying defaults.
func (c *GossipConfig) ApplyDefaults(defaultPort int) error {
	if c.NodeName == "" {
		hn, err := os.Hostname()
		if err != nil {
			return fmt.Errorf("generating node name: %w", err)
		}
		c.NodeName = hn
	}

	if c.AdvertiseAddr == "" {
		if len(c.AdvertiseInterfaces) == 0 {
			return fmt.Errorf("one of advertise address or advertise interfaces must be set")
		}

		addr, err := advertise.FirstAddress(c.AdvertiseInterfaces)
		if err != nil {
			return fmt.Errorf("determining advertise address: %w", err)
		}
		c.AdvertiseAddr = fmt.Sprintf("%s:%d", addr.String(), defaultPort)
	} else {
		c.AdvertiseAddr = appendDefaultPort(c.AdvertiseAddr, defaultPort)
	}

	if len(c.JoinPeers) > 0 && c.DiscoverPeers != "" {
		return fmt.Errorf("at most one of join peers and discover peers may be set")
	} else if c.DiscoverPeers != "" {
		providers := make(map[string]discover.Provider, len(discover.Providers)+1)
		for k, v := range discover.Providers {
			providers[k] = v
		}
		// Extra providers used by tests
		for k, v := range extraDiscoverProviders {
			providers[k] = v
		}

		// Custom providers that aren't enabled by default
		providers["k8s"] = &k8s.Provider{}

		d, err := discover.New(discover.WithProviders(providers))
		if err != nil {
			return fmt.Errorf("bootstrapping peer discovery: %w", err)
		}

		addrs, err := d.Addrs(c.DiscoverPeers, stdlog.New(io.Discard, "", 0)) // TODO(rfratto): log to log.Logger?
		if err != nil {
			return fmt.Errorf("discovering peers: %w", err)
		}
		c.JoinPeers = addrs
	}

	for i := range c.JoinPeers {
		// Default to using the same advertise port as the local node. This may
		// break in some cases, so the user should make sure the port numbers
		// align on as many nodes as possible.
		c.JoinPeers[i] = appendDefaultPort(c.JoinPeers[i], defaultPort)
	}

	return nil
}

func appendDefaultPort(addr string, port int) string {
	_, _, err := net.SplitHostPort(addr)
	if err == nil {
		// No error means there was a port in the string
		return addr
	}
	return fmt.Sprintf("%s:%d", addr, port)
}

// GossipNode is a Node which uses gRPC and gossip to discover peers.
type GossipNode struct {
	// NOTE(rfratto): GossipNode is a *very* thin wrapper over ckit.Node, but it
	// still abstracted out as its own type to have more agent-specific control
	// over the exposed API.

	cfg       *GossipConfig
	innerNode *ckit.Node
	log       log.Logger
	sharder   shard.Sharder

	started atomic.Bool
}

// NewGossipNode creates an unstarted GossipNode. The GossipNode will register
// HTTP endpoints to communicate with other nodes over HTTP/2. GossipConfig is
// expected to be valid and have already had ApplyDefaults called on it.
//
// GossipNode operations are unavailable until the node is started.
func NewGossipNode(l log.Logger, mux *http.ServeMux, c *GossipConfig) (*GossipNode, error) {
	if l == nil {
		l = log.NewNopLogger()
	}

	sharder := shard.Ring(tokensPerNode)

	ckitConfig := ckit.Config{
		Name:          c.NodeName,
		AdvertiseAddr: c.AdvertiseAddr,
		Sharder:       sharder,
		Log:           l,
		Pool:          c.Pool,
	}

	inner, err := ckit.NewHTTPNode(mux, ckitConfig)
	if err != nil {
		return nil, err
	}

	return &GossipNode{
		cfg:       c,
		innerNode: inner,
		log:       l,
		sharder:   sharder,
	}, nil
}

// ChangeState changes the state of n. ChangeState will block until the state
// change has been received by another node; cancel the context to stop
// waiting. ChangeState will fail if the current state cannot move to the
// target state.
//
// Nodes must be a StateParticipant to receive writes.
func (n *GossipNode) ChangeState(ctx context.Context, to peer.State) error {
	if !n.started.Load() {
		return fmt.Errorf("node not started")
	}
	return n.innerNode.ChangeState(ctx, to)
}

// CurrentState returns the current state of the node. Note that other nodes
// may have an older view of the state while a state change propagates
// throughout the cluster.
func (n *GossipNode) CurrentState() peer.State {
	return n.innerNode.CurrentState()
}

// Lookup implements Node and returns numOwners Peers that are responsible for
// key. Only peers in StateParticipant are considered during a lookup; if no
// peers are in StateParticipant, the Lookup will fail.
func (n *GossipNode) Lookup(key shard.Key, numOwners int, op shard.Op) ([]peer.Peer, error) {
	if !n.started.Load() {
		return nil, fmt.Errorf("node not started")
	}
	return n.sharder.Lookup(key, numOwners, op)
}

// Observe registers o to be informed when the cluster changes, including peers
// appearing, disappearing, or changing state.
//
// Calls will have to filter events if they are only interested in a subset of
// changes.
func (n *GossipNode) Observe(o ckit.Observer) {
	n.innerNode.Observe(o)
}

// Peers returns the current set of Peers.
func (n *GossipNode) Peers() []peer.Peer {
	return n.innerNode.Peers()
}

// Start starts the node. Start will connect to peers if configured to do so.
//
// Start must only be called after the gRPC server is running, otherwise Start
// will block forever.
func (n *GossipNode) Start() (err error) {
	defer func() {
		if err == nil {
			n.started.Store(true)
		}
	}()
	return n.innerNode.Start(n.cfg.JoinPeers)
}

// Stop leaves the cluster and terminates n. n cannot be re-used after
// stopping.
//
// It is advisable to ChangeState to StateTerminating and StateGone before
// stopping so the local node has an opportunity to move work to other nodes.
func (n *GossipNode) Stop() error {
	return n.innerNode.Stop()
}
