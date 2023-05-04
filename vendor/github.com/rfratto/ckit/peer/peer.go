// Package peer describes a ckit peer.
package peer

// Peer is a discovered node within the cluster.
type Peer struct {
	Name  string // Name of the Peer. Unique across the cluster.
	Addr  string // host:port address of the peer.
	Self  bool   // True if Peer is the local Node.
	State State  // State of the peer.
}

// String returns the name of p.
func (p Peer) String() string { return p.Name }
