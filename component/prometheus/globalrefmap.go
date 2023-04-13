package prometheus

import (
	"sync"
	"time"

	"github.com/prometheus/prometheus/model/labels"
)

// GlobalRefMapping is used when translating to and from remote writes and the rest of the system (mostly scrapers)
// normal components except those should in general NOT need this.
var GlobalRefMapping = &GlobalRefMap{}

func init() {
	GlobalRefMapping = newGlobalRefMap()
}

// staleDuration determines how often we should wait after a stale value is received to GC that value
var staleDuration = time.Minute * 10

// GlobalRefMap allows conversion from remote_write refids to global refs ids that everything else can use
type GlobalRefMap struct {
	mut                sync.Mutex
	globalRefID        uint64
	mappings           map[string]*remoteWriteMapping
	labelsHashToGlobal map[uint64]uint64
	staleGlobals       map[uint64]*staleMarker
}

type staleMarker struct {
	globalID        uint64
	lastMarkedStale time.Time
	labelHash       uint64
}

// newGlobalRefMap creates a refmap for usage, there should ONLY be one of these
func newGlobalRefMap() *GlobalRefMap {
	return &GlobalRefMap{
		globalRefID:        0,
		mappings:           make(map[string]*remoteWriteMapping),
		labelsHashToGlobal: make(map[uint64]uint64),
		staleGlobals:       make(map[uint64]*staleMarker),
	}
}

// GetOrAddLink is called by a remote_write endpoint component to add mapping and get back the global id.
func (g *GlobalRefMap) GetOrAddLink(componentID string, localRefID uint64, lbls labels.Labels) uint64 {
	g.mut.Lock()
	defer g.mut.Unlock()

	// If the mapping doesn't exist then we need to create it
	m, found := g.mappings[componentID]
	if !found {
		m = &remoteWriteMapping{
			RemoteWriteID: componentID,
			localToGlobal: make(map[uint64]uint64),
			globalToLocal: make(map[uint64]uint64),
		}
		g.mappings[componentID] = m
	}

	labelHash := lbls.Hash()
	globalID, found := g.labelsHashToGlobal[labelHash]
	if found {
		m.localToGlobal[localRefID] = globalID
		m.globalToLocal[globalID] = localRefID
		return globalID
	}
	// We have a value we have never seen before so increment the globalrefid and assign
	g.globalRefID++
	g.labelsHashToGlobal[labelHash] = g.globalRefID
	m.localToGlobal[localRefID] = g.globalRefID
	m.globalToLocal[g.globalRefID] = localRefID
	return g.globalRefID
}

// GetOrAddGlobalRefID is used to create a global refid for a labelset
func (g *GlobalRefMap) GetOrAddGlobalRefID(l labels.Labels) uint64 {
	g.mut.Lock()
	defer g.mut.Unlock()

	// Guard against bad input.
	if l == nil {
		return 0
	}

	labelHash := l.Hash()
	globalID, found := g.labelsHashToGlobal[labelHash]
	if found {
		return globalID
	}
	g.globalRefID++
	g.labelsHashToGlobal[labelHash] = g.globalRefID
	return g.globalRefID
}

// GetGlobalRefID returns the global refid for a component local combo, or 0 if not found
func (g *GlobalRefMap) GetGlobalRefID(componentID string, localRefID uint64) uint64 {
	g.mut.Lock()
	defer g.mut.Unlock()

	m, found := g.mappings[componentID]
	if !found {
		return 0
	}
	global := m.localToGlobal[localRefID]
	return global
}

// GetLocalRefID returns the local refid for a component global combo, or 0 if not found
func (g *GlobalRefMap) GetLocalRefID(componentID string, globalRefID uint64) uint64 {
	g.mut.Lock()
	defer g.mut.Unlock()

	m, found := g.mappings[componentID]
	if !found {
		return 0
	}
	local := m.globalToLocal[globalRefID]
	return local
}

// AddStaleMarker adds a stale marker
func (g *GlobalRefMap) AddStaleMarker(globalRefID uint64, l labels.Labels) {
	g.mut.Lock()
	defer g.mut.Unlock()

	g.staleGlobals[globalRefID] = &staleMarker{
		lastMarkedStale: time.Now(),
		labelHash:       l.Hash(),
		globalID:        globalRefID,
	}
}

// RemoveStaleMarker removes a stale marker
func (g *GlobalRefMap) RemoveStaleMarker(globalRefID uint64) {
	g.mut.Lock()
	defer g.mut.Unlock()

	delete(g.staleGlobals, globalRefID)
}

// CheckStaleMarkers is called to garbage collect and items that have grown stale over stale duration (10m)
func (g *GlobalRefMap) CheckStaleMarkers() {
	g.mut.Lock()
	defer g.mut.Unlock()

	curr := time.Now()
	idsToBeGCed := make([]*staleMarker, 0)
	for _, stale := range g.staleGlobals {
		// If the difference between now and the last time the stale was marked doesn't exceed stale then let it stay
		if curr.Sub(stale.lastMarkedStale) < staleDuration {
			continue
		}
		idsToBeGCed = append(idsToBeGCed, stale)
	}
	for _, marker := range idsToBeGCed {
		delete(g.staleGlobals, marker.globalID)
		delete(g.labelsHashToGlobal, marker.labelHash)
		// Delete our mapping keys
		for _, mapping := range g.mappings {
			mapping.deleteStaleIDs(marker.globalID)
		}
	}
}
