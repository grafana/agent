package metrics

import (
	"sync"

	"github.com/prometheus/prometheus/model/labels"
)

var GlobalRefMapping = &GlobalRefMap{}

func init() {
	GlobalRefMapping = NewGlobalRefMap()
}

type GlobalRefMap struct {
	mut                sync.Mutex
	globalRefID        uint64
	mappings           map[string]*mapping
	labelsHashToGlobal map[uint64]uint64
}

type mapping struct {
	RemoteWriteID string
	localToGlobal map[uint64]uint64
	globalToLocal map[uint64]uint64
}

func NewGlobalRefMap() *GlobalRefMap {
	return &GlobalRefMap{
		globalRefID:        0,
		mappings:           make(map[string]*mapping),
		labelsHashToGlobal: make(map[uint64]uint64),
	}
}

func (g *GlobalRefMap) AddLink(componentID string, localRefID uint64, l labels.Labels) uint64 {
	g.mut.Lock()
	defer g.mut.Unlock()

	m, found := g.mappings[componentID]
	if !found {
		m = &mapping{
			RemoteWriteID: componentID,
			localToGlobal: make(map[uint64]uint64),
			globalToLocal: make(map[uint64]uint64),
		}
		g.mappings[componentID] = m
	}
	labelHash := l.Hash()
	globalID, found := g.labelsHashToGlobal[labelHash]
	if found {
		m.localToGlobal[localRefID] = g.globalRefID
		m.globalToLocal[g.globalRefID] = localRefID
		return globalID
	}
	g.globalRefID++
	g.labelsHashToGlobal[labelHash] = g.globalRefID
	m.localToGlobal[localRefID] = g.globalRefID
	m.globalToLocal[g.globalRefID] = localRefID
	return g.globalRefID
}

func (g *GlobalRefMap) CreateGlobalRefID(l labels.Labels) uint64 {
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

func (g *GlobalRefMap) GetGlobal(componentID string, localRefID uint64) uint64 {
	g.mut.Lock()
	defer g.mut.Unlock()
	m, found := g.mappings[componentID]
	if !found {
		return 0
	}
	global, _ := m.localToGlobal[localRefID]
	return global
}

func (g *GlobalRefMap) GetLocal(componentID string, globalRefID uint64) uint64 {
	g.mut.Lock()
	defer g.mut.Unlock()
	m, found := g.mappings[componentID]
	if !found {
		return 0
	}
	local, _ := m.globalToLocal[globalRefID]
	return local
}
