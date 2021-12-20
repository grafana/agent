package metrics

import (
	"fmt"
	"sort"
	"sync"

	"github.com/grafana/agent/pkg/cluster"
	"github.com/rfratto/ckit"
	"github.com/rfratto/ckit/chash"
)

// hasher attaches to a cluster and will compute read-only copies of the
// cluster state for hashing.
//
// This is required for sharding, where the calculation of shards must be set
// against a fixed set of nodes. This is incompatible with the builtin sharding
// mechanism of ckit which returns the current owner dynamically.
type hasher struct {
	mut        sync.Mutex
	watchers   []func(hr *hashReader) bool
	lastReader *hashReader
}

// newHasher creates a new hasher.
func newHasher(builder func() chash.Hash, node *cluster.Node) *hasher {
	h := &hasher{}

	node.OnPeersChanged(func(ps ckit.PeerSet) (reregister bool) {
		h.mut.Lock()
		defer h.mut.Unlock()

		// Create a new hashreader.
		var (
			hr = &hashReader{
				hf: builder(),
				ps: make(map[string]*ckit.Peer, len(ps)),
			}
			nodes = make([]string, len(ps))
		)
		for i, p := range ps {
			nodes[i] = p.Name
			hr.ps[p.Name] = &p
		}
		sort.Strings(nodes)
		hr.hf.SetNodes(nodes)

		// Notify anyone watching for changes.
		newWatchers := make([]func(*hashReader) bool, 0, len(h.watchers))
		for _, w := range h.watchers {
			register := w(hr)
			if register {
				newWatchers = append(newWatchers, w)
			}
		}
		h.watchers = newWatchers
		h.lastReader = hr

		return true
	})

	return h
}

// OnPeersChanged registers cb to be invoked whenever peers change. cb will be
// unregistered if it returns false.
//
// cb will be provided a static hashReader. New hashReaders will be created
// whenever the peer set changes.
//
// cb will be invoked immediately with the most recent set of peers, if any.
func (h *hasher) OnPeersChanged(cb func(*hashReader) bool) {
	h.mut.Lock()
	defer h.mut.Unlock()

	if h.lastReader != nil && !cb(h.lastReader) {
		// Don't register if our initial call didn't want to be re-registered.
		return
	}
	h.watchers = append(h.watchers, cb)
}

type hashReader struct {
	hf chash.Hash
	ps map[string]*ckit.Peer
}

// Peers returns all known peers. The resulting map must not be modified
// directly; make a copy if the data needs to be modified.
func (hr *hashReader) Peers() map[string]*ckit.Peer { return hr.ps }

// Get returns the owner of a key. An error will be returned if the hash fails
// or if there aren't enough peers.
func (hr *hashReader) Get(key string) (*ckit.Peer, error) {
	owners, err := hr.hf.Get(key, 1)
	if err != nil {
		return nil, fmt.Errorf("failed to get owner: %w", err)
	} else if len(owners) == 0 {
		return nil, fmt.Errorf("failed to get owner: no peers")
	}
	owner := hr.ps[owners[0]]
	if owner == nil {
		return nil, fmt.Errorf("failed to get owner: %q does not exist in cluster", owners[0])
	}
	return owner, nil
}
