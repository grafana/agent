// Package cluster enables an agent-wide cluster mechanism which subsystems can
// use to determine ownership of some key.
package cluster

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/grafana/ckit"
	"github.com/grafana/ckit/peer"
	"github.com/grafana/ckit/shard"
	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/net/http2"
)

// Node is a read-only view of a cluster node.
type Node interface {
	// Lookup determines the set of replicationFactor owners for a given key.
	// peer.Peer.Self can be used to determine if the local node is the owner,
	// allowing for short-circuiting logic to connect directly to the local node
	// instead of using the network.
	//
	// Callers can use github.com/grafana/ckit/shard.StringKey or
	// shard.NewKeyBuilder to create a key.
	Lookup(key shard.Key, replicationFactor int, op shard.Op) ([]peer.Peer, error)

	// Observe registers an Observer to receive notifications when the set of
	// Peers for a Node changes.
	Observe(ckit.Observer)

	// Peers returns the current set of peers for a Node.
	Peers() []peer.Peer

	Handler() (string, http.Handler)
}

// NewLocalNode returns a Node which forms a single-node cluster and never
// connects to other nodes.
//
// selfAddr is the address for a Node to use to connect to itself over gRPC.
func NewLocalNode(selfAddr string) Node {
	p := peer.Peer{
		Name:  "local",
		Addr:  selfAddr,
		Self:  true,
		State: peer.StateParticipant,
	}

	return &localNode{self: p}
}

type localNode struct{ self peer.Peer }

func (ln *localNode) Lookup(key shard.Key, replicationFactor int, op shard.Op) ([]peer.Peer, error) {
	if replicationFactor == 0 {
		return nil, nil
	} else if replicationFactor > 1 {
		return nil, fmt.Errorf("need %d nodes; only 1 available", replicationFactor)
	}

	return []peer.Peer{ln.self}, nil
}

func (ln *localNode) Observe(ckit.Observer) {
	// no-op: the cluster will never change for a local-only node.
}

func (ln *localNode) Peers() []peer.Peer {
	return []peer.Peer{ln.self}
}

func (ln *localNode) Handler() (string, http.Handler) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("clustering is disabled"))
		w.WriteHeader(http.StatusBadRequest)
	}))

	return "/api/v1/ckit/transport/", mux
}

// Clusterer implements the behavior required for operating Flow controllers
// in a distributed fashion.
type Clusterer struct {
	Node Node
}

func getJoinAddr(addrs []string, in string) []string {
	_, _, err := net.SplitHostPort(in)
	if err == nil {
		addrs = append(addrs, in)
		return addrs
	}

	ip := net.ParseIP(in)
	if ip != nil {
		addrs = append(addrs, ip.String())
		return addrs
	}

	_, srvs, err := net.LookupSRV("", "", in)
	if err == nil {
		for _, srv := range srvs {
			addrs = append(addrs, srv.Target)
		}
	}

	return addrs
}

// New creates a Clusterer.
func New(log log.Logger, reg prometheus.Registerer, clusterEnabled bool, name, listenAddr, advertiseAddr, joinAddr string) (*Clusterer, error) {
	// Standalone node.
	if !clusterEnabled {
		return &Clusterer{Node: NewLocalNode(listenAddr)}, nil
	}

	gossipConfig := DefaultGossipConfig

	defaultPort := 80
	_, portStr, err := net.SplitHostPort(listenAddr)
	if err == nil { // there was a port
		defaultPort, err = strconv.Atoi(portStr)
		if err != nil {
			return nil, err
		}
	}

	if name != "" {
		gossipConfig.NodeName = name
	}

	if advertiseAddr != "" {
		gossipConfig.AdvertiseAddr = advertiseAddr
	}

	if joinAddr != "" {
		gossipConfig.JoinPeers = []string{}
		jaddrs := strings.Split(joinAddr, ",")
		for _, jaddr := range jaddrs {
			gossipConfig.JoinPeers = getJoinAddr(gossipConfig.JoinPeers, jaddr)
		}
	}

	err = gossipConfig.ApplyDefaults(defaultPort)
	if err != nil {
		return nil, err
	}

	cli := &http.Client{
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

	level.Info(log).Log("msg", "starting a new gossip node", "join-peers", gossipConfig.JoinPeers)

	gossipNode, err := NewGossipNode(log, reg, cli, &gossipConfig)
	if err != nil {
		return nil, err
	}

	return &Clusterer{Node: gossipNode}, nil
}

// Start starts the node.
// For the localNode implementation, this is a no-op.
// For the gossipNode implementation, Start will attempt to connect to the
// configured list of peers; if this fails it will fall back to bootstrapping a
// new cluster of its own.
func (c *Clusterer) Start(ctx context.Context) error {
	switch node := c.Node.(type) {
	case *localNode:
		return nil // no-op, always ready
	case *GossipNode:
		err := node.Start() // TODO(@tpaschalis) Should we backoff and retry before moving on to the fallback here?
		if err != nil {
			level.Debug(node.log).Log("msg", "failed to connect to peers; bootstrapping a new cluster")
			node.cfg.JoinPeers = nil
			err = node.Start()
			if err != nil {
				return err
			}
		}

		// We now have either joined or started a new cluster.
		// Nodes initially join in the Viewer state. We can move to the
		// Participant state to signal that we wish to participate in reading
		// or writing data.
		ctx, ccl := context.WithTimeout(ctx, 5*time.Second)
		defer ccl()
		err = node.ChangeState(ctx, peer.StateParticipant)
		if err != nil {
			return err
		}

		node.Observe(ckit.FuncObserver(func(peers []peer.Peer) (reregister bool) {
			names := make([]string, len(peers))
			for i, p := range peers {
				names[i] = p.Name
			}
			level.Info(node.log).Log("msg", "peers changed", "new_peers", strings.Join(names, ","))
			return true
		}))
		return nil
	default:
		msg := fmt.Sprintf("node type: %T", c.Node)
		panic("cluster: unreachable:" + msg)
	}
}

// Stop stops the Clusterer.
func (c *Clusterer) Stop() error {
	switch node := c.Node.(type) {
	case *GossipNode:
		// The node is going away. We move to the Terminating state to signal
		// that we should not be owners for write hashing operations anymore.
		ctx, ccl := context.WithTimeout(context.Background(), 5*time.Second)
		defer ccl()

		// TODO(rfratto): should we enter terminating state earlier to allow for
		// some kind of hand-off between components?
		err := node.ChangeState(ctx, peer.StateTerminating)
		if err != nil {
			level.Error(node.log).Log("msg", "failed to change state to Terminating before shutting down", "err", err)
		}
		return node.Stop()
	}

	// Nothing to do for unrecognized types.
	return nil
}

func deadlineDuration(ctx context.Context) (d time.Duration, ok bool) {
	if t, ok := ctx.Deadline(); ok {
		return time.Until(t), true
	}
	return 0, false
}
