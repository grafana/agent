package servicegraphprocessor

import (
	"container/list"
	"sync"
	"time"
)

type store struct {
	l   *list.List
	mtx *sync.RWMutex
	m   map[string]*list.Element

	closeCh chan struct{}

	evictCallback func(e *edge)
	ttl           time.Duration
}

func newStore(ttl time.Duration, evictCallback func(e *edge)) *store {
	s := &store{
		l:   list.New(),
		mtx: &sync.RWMutex{},
		m:   make(map[string]*list.Element),

		closeCh: make(chan struct{}, 1),

		evictCallback: evictCallback,
		ttl:           ttl,
	}

	return s
}

func (s *store) shutdown() {
	close(s.closeCh)
}

func (s *store) len() int {
	s.mtx.RLock()
	defer s.mtx.RUnlock()

	return s.l.Len()
}

// shouldEvictHead checks if the oldest item (head of list) has expired and should be evicted.
// Returns true if the item has expired, false otherwise.
//
// Must be called under lock.
func (s *store) shouldEvictHead() bool {
	h := s.l.Front()
	if h == nil {
		return false
	}
	ts := h.Value.(*edge).expiration
	return ts < time.Now().Unix()
}

// evictHead removes the head from the store (and map).
// It also collects metrics for the evicted edge.
//
// Must be called under lock.
func (s *store) evictHead() {
	front := s.l.Front().Value.(*edge)
	s.evictEdge(front.key)
}

// evictEdge evicts and edge under lock
func (s *store) evictEdgeWithLock(key string) {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	s.evictEdge(key)
}

// evictEdge removes the edge from the store (and map).
// It also collects metrics for the evicted edge.
//
// Must be called under lock.
func (s *store) evictEdge(key string) {
	ele := s.m[key]
	if ele == nil { // it may already have been processed
		return
	}

	edge := ele.Value.(*edge)
	s.evictCallback(edge)

	delete(s.m, key)
	s.l.Remove(ele)
}

// Fetches an edge from the store.
// If the edge doesn't exist, it creates a new one with the default TTL.
func (s *store) upsertEdge(k string, cb func(e *edge)) *edge {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	if storedEdge, ok := s.m[k]; ok {
		edge := storedEdge.Value.(*edge)
		cb(edge)
		return edge
	}

	newEdge := newEdge(k, s.ttl)
	ele := s.l.PushBack(newEdge)
	s.m[k] = ele
	cb(newEdge)

	return newEdge
}

// expire evicts all expired items in the store.
func (s *store) expire() {
	s.mtx.RLock()
	if !s.shouldEvictHead() {
		s.mtx.RUnlock()
		return
	}
	s.mtx.RUnlock()

	s.mtx.Lock()
	defer s.mtx.Unlock()

	for s.shouldEvictHead() {
		s.evictHead()
	}
}
