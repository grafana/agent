package controller

import (
	"sync"
)

// Queue is a thread-safe, insertion-ordered set of components.
//
// Queue is intended for tracking components that have updated their Exports
// for later reevaluation.
type Queue struct {
	mut         sync.Mutex
	queuedSet   map[*ComponentNode]struct{}
	queuedOrder []*ComponentNode

	updateCh chan struct{}
}

// NewQueue returns a new queue.
func NewQueue() *Queue {
	return &Queue{
		updateCh:    make(chan struct{}, 1),
		queuedSet:   make(map[*ComponentNode]struct{}),
		queuedOrder: make([]*ComponentNode, 0),
	}
}

// Enqueue inserts a new component into the Queue. Enqueue is a no-op if the
// component is already in the Queue.
func (q *Queue) Enqueue(c *ComponentNode) {
	q.mut.Lock()
	defer q.mut.Unlock()

	// Skip if already queued.
	if _, ok := q.queuedSet[c]; ok {
		return
	}

	q.queuedOrder = append(q.queuedOrder, c)
	q.queuedSet[c] = struct{}{}
	select {
	case q.updateCh <- struct{}{}:
	default:
	}
}

// Chan returns a channel which is written to when the queue is non-empty.
func (q *Queue) Chan() <-chan struct{} { return q.updateCh }

// DequeueAll removes all components from the queue and returns them.
func (q *Queue) DequeueAll() []*ComponentNode {
	q.mut.Lock()
	defer q.mut.Unlock()

	all := q.queuedOrder
	q.queuedOrder = make([]*ComponentNode, 0)
	q.queuedSet = make(map[*ComponentNode]struct{})

	return all
}
