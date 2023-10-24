package labelstore

import "github.com/prometheus/prometheus/model/labels"

type LabelStore interface {

	// GetOrAddLink returns the global id for the values, if none found one will be created based on the lbls.
	GetOrAddLink(componentID string, localRefID uint64, lbls labels.Labels) uint64

	// GetOrAddGlobalRefID finds or adds a global id for the given label map.
	GetOrAddGlobalRefID(l labels.Labels) uint64

	// GetGlobalRefID returns the global id for a component and the local id. Returns 0 if nothing found.
	GetGlobalRefID(componentID string, localRefID uint64) uint64

	// GetLocalRefID gets the mapping from global to local id specific to a component. Returns 0 if nothing found.
	GetLocalRefID(componentID string, globalRefID uint64) uint64

	// AddStaleMarker adds a stale marker to a reference, that reference will then get removed on the next check.
	AddStaleMarker(globalRefID uint64, l labels.Labels)

	// RemoveStaleMarker removes the stale marker for a reference, keeping it around.
	RemoveStaleMarker(globalRefID uint64)

	// CheckAndRemoveStaleMarkers identifies any series with a stale marker and removes those entries from the LabelStore.
	CheckAndRemoveStaleMarkers()
}
