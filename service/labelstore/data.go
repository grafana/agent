package labelstore

import "github.com/prometheus/prometheus/model/labels"

type LabelStore interface {
	GetOrAddLink(componentID string, localRefID uint64, lbls labels.Labels) uint64
	GetOrAddGlobalRefID(l labels.Labels) uint64
	GetGlobalRefID(componentID string, localRefID uint64) uint64
	GetLocalRefID(componentID string, globalRefID uint64) uint64
	AddStaleMarker(globalRefID uint64, l labels.Labels)
	RemoveStaleMarker(globalRefID uint64)
	CheckStaleMarkers()
}
