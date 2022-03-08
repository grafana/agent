// Package cluster enables an agent-wide cluster mechanism which subsystems can
// use to determine ownership of some key.
package cluster

import (
	"fmt"

	"github.com/rfratto/ckit"
)

// NOTE(rfratto): pkg/cluster currently isn't wired in yet, but will be used
// for the implementation of RFC-0003. Try to remember to remove this comment
// once it gets used :)

// Node is a read-only view of a cluster node.
type Node interface {
	// Lookup determines the set of replicationFactor owners for a given key.
	// ckit.Peer.Self can be used to determine if this Node is the owner for
	// short-circuiting logic.
	//
	// Callers can use github.com/rfratto/ckit/chash.Key or chash.NewKeyBuilder
	// to create a key.
	Lookup(key uint64, replicationFactor int) ([]ckit.Peer, error)

	// Observe registers an Observer to receive notifications when the set of
	// Peers for a Node changes.
	Observe(ckit.Observer)

	// Peers returns the current set of peers for a Node.
	Peers() []ckit.Peer
}

// NewLocalNode returns a Node which forms a single-node cluster and never
// connects to other nodes.
//
// selfAddr is the address for a Node to use to connect to itself over gRPC.
func NewLocalNode(selfAddr string) Node {
	p := ckit.Peer{
		Name:  "local",
		Addr:  selfAddr,
		Self:  true,
		State: ckit.StateParticipant,
	}

	return &localNode{self: p}
}

type localNode struct{ self ckit.Peer }

func (ln *localNode) Lookup(key uint64, replicationFactor int) ([]ckit.Peer, error) {
	if replicationFactor == 0 {
		return nil, nil
	} else if replicationFactor > 1 {
		return nil, fmt.Errorf("need %d nodes; only 1 available", replicationFactor)
	}

	return []ckit.Peer{ln.self}, nil
}

func (ln *localNode) Observe(ckit.Observer) {
	// no-op: the cluster will never change for a local-only node.
}

func (ln *localNode) Peers() []ckit.Peer {
	return []ckit.Peer{ln.self}
}
