package metrics

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
	globalRefID        RefID
	mappings           map[string]*remoteWriteMapping
	labelsHashToGlobal map[uint64]RefID
	staleGlobals       map[RefID]*staleMarker
}

type staleMarker struct {
	globalID        RefID
	lastMarkedStale time.Time
	labelHash       uint64
}

// newGlobalRefMap creates a refmap for usage, there should ONLY be one of these
func newGlobalRefMap() *GlobalRefMap {
	return &GlobalRefMap{
		globalRefID:        0,
		mappings:           make(map[string]*remoteWriteMapping),
		labelsHashToGlobal: make(map[uint64]RefID),
		staleGlobals:       make(map[RefID]*staleMarker),
	}
}

// UnregisterComponent deletes all the mappings for a given component
func (g *GlobalRefMap) UnregisterComponent(componentID string) {
	g.mut.Lock()
	defer g.mut.Unlock()

	delete(g.mappings, componentID)
}

// GetOrAddLink is called by a remote_write endpoint component to add mapping and get back the global id.
func (g *GlobalRefMap) GetOrAddLink(componentID string, localRefID uint64, fm *FlowMetric) RefID {
	g.mut.Lock()
	defer g.mut.Unlock()

	// If the mapping doesn't exist then we need to create it
	m, found := g.mappings[componentID]
	if !found {
		m = &remoteWriteMapping{
			RemoteWriteID: componentID,
			localToGlobal: make(map[uint64]RefID),
			globalToLocal: make(map[RefID]uint64),
		}
		g.mappings[componentID] = m
	}

	labelHash := fm.labels.Hash()
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

// getGlobalRefIDByLabels retrieves a global id based on the labels
func (g *GlobalRefMap) getGlobalRefIDByLabels(l labels.Labels) RefID {
	g.mut.Lock()
	defer g.mut.Unlock()

	labelHash := l.Hash()
	globalID, found := g.labelsHashToGlobal[labelHash]
	if found {
		return globalID
	}
	g.globalRefID++
	g.labelsHashToGlobal[labelHash] = g.globalRefID
	return g.globalRefID
}

// GetGlobalRefIDForComponent returns the global refid for a component local combo, or 0 if not found
func (g *GlobalRefMap) GetGlobalRefIDForComponent(componentID string, localRefID uint64) RefID {
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
func (g *GlobalRefMap) GetLocalRefID(componentID string, globalRefID RefID) uint64 {
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
func (g *GlobalRefMap) AddStaleMarker(globalRefID RefID, l labels.Labels) {
	g.mut.Lock()
	defer g.mut.Unlock()

	g.staleGlobals[globalRefID] = &staleMarker{
		lastMarkedStale: time.Now(),
		labelHash:       l.Hash(),
		globalID:        globalRefID,
	}
}

// RemoveStaleMarker removes a stale marker
func (g *GlobalRefMap) RemoveStaleMarker(globalRefID RefID) {
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
		// If the difference between now and the last time the stale was marked doesnt exceed stale then let it stay
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
