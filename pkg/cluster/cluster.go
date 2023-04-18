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

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/rfratto/ckit"
	"github.com/rfratto/ckit/peer"
	"github.com/rfratto/ckit/shard"
	"golang.org/x/net/http2"
)

// Node is a read-only view of a cluster node.
type Node interface {
	// Lookup determines the set of replicationFactor owners for a given key.
	// peer.Peer.Self can be used to determine if the local node is the owner,
	// allowing for short-circuiting logic to connect directly to the local node
	// instead of using the network.
	//
	// Callers can use github.com/rfratto/ckit/shard.StringKey or
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
			addrs = append(addrs, fmt.Sprintf("%s:%d", srv.Target, srv.Port))
		}
	}

	return addrs
}

// New creates a Clusterer.
func New(log log.Logger, clusterEnabled bool, addr, joinAddr string) (*Clusterer, error) {
	// Standalone node.
	if !clusterEnabled {
		return &Clusterer{Node: NewLocalNode(addr)}, nil
	}

	gossipConfig := DefaultGossipConfig

	defaultPort := 80
	if addr != "" {
		host, portStr, err := net.SplitHostPort(addr)
		if err != nil {
			return nil, err
		}
		defaultPort, err = strconv.Atoi(portStr)
		if err != nil {
			return nil, err
		}
		gossipConfig.AdvertiseAddr = host
	}

	err := gossipConfig.ApplyDefaults(defaultPort)
	if err != nil {
		return nil, err
	}

	if joinAddr != "" {
		gossipConfig.JoinPeers = []string{}
		jaddrs := strings.Split(joinAddr, ",")
		for _, jaddr := range jaddrs {
			gossipConfig.JoinPeers = getJoinAddr(gossipConfig.JoinPeers, jaddr)
		}
	}

	cli := &http.Client{
		Transport: &http2.Transport{
			AllowHTTP: true,
			DialTLS: func(network, addr string, _ *tls.Config) (net.Conn, error) {
				return net.Dial(network, addr)
			},
		},
	}

	gossipNode, err := NewGossipNode(log, cli, &gossipConfig)
	if err != nil {
		return nil, err
	}

	// Attempt to start the Node by connecting to the peers in gossipConfig.
	// If we cannot connect to any peers, fall back to bootstrapping a new
	// cluster by ourselves.
	err = gossipNode.Start()
	if err != nil {
		level.Debug(log).Log("msg", "failed to connect to peers; bootstrapping a new cluster")
		gossipConfig.JoinPeers = nil
		err = gossipNode.Start()
		if err != nil {
			return nil, err
		}
	}

	// Nodes initially join the cluster in the Viewer state. We can move to the
	// Participant state to signal that we wish to participate in reading or
	// writing data.
	err = gossipNode.ChangeState(context.Background(), peer.StateParticipant)
	if err != nil {
		return nil, err
	}

	res := &Clusterer{Node: gossipNode}

	gossipNode.Observe(ckit.FuncObserver(func(peers []peer.Peer) (reregister bool) {
		names := make([]string, len(peers))
		for i, p := range peers {
			names[i] = p.Name
		}
		level.Info(log).Log("msg", "peers changed", "new_peers", strings.Join(names, ","))
		return true
	}))

	return res, nil
}
