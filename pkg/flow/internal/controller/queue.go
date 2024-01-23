package controller

import (
	"sync"
)

// Queue is a thread-safe, insertion-ordered set of nodes.
//
// Queue is intended for tracking nodes that have updated their Exports
// for later reevaluation.
type Queue struct {
	mut         sync.Mutex
	queuedSet   map[NodeWithDependants]struct{}
	queuedOrder []NodeWithDependants

	updateCh chan struct{}
}

// NewQueue returns a new queue.
func NewQueue() *Queue {
	return &Queue{
		updateCh:    make(chan struct{}, 1),
		queuedSet:   make(map[NodeWithDependants]struct{}),
		queuedOrder: make([]NodeWithDependants, 0),
	}
}

// Enqueue inserts a new NodeWithDependants into the Queue. Enqueue is a no-op if the
// NodeWithDependants is already in the Queue.
func (q *Queue) Enqueue(c NodeWithDependants) {
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

// DequeueAll removes all NodeWithDependants from the queue and returns them.
func (q *Queue) DequeueAll() []NodeWithDependants {
	q.mut.Lock()
	defer q.mut.Unlock()

	all := q.queuedOrder
	q.queuedOrder = make([]NodeWithDependants, 0)
	q.queuedSet = make(map[NodeWithDependants]struct{})

	return all
}
