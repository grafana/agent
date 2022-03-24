// Package cluster enables an agent-wide cluster mechanism which subsystems can
// use to determine ownership of some key.
package cluster

import (
	"fmt"

	"github.com/rfratto/ckit"
	"github.com/rfratto/ckit/peer"
	"github.com/rfratto/ckit/shard"
)

// NOTE(rfratto): pkg/cluster currently isn't wired in yet, but will be used
// for the implementation of RFC-0003. Try to remember to remove this comment
// once it gets used :)

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
