package controller

import (
	"fmt"
	"sync"
)

// Queue is an insertion-ordered set of components.
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
	fmt.Printf("\n=== Enque for update: %q\n", c.NodeID())

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

// TryDequeue dequeues the first component in insertion order. TryDequeue will return nil
// if the queue is empty.
func (q *Queue) TryDequeue() *ComponentNode {
	q.mut.Lock()
	defer q.mut.Unlock()

	if len(q.queuedSet) == 0 {
		return nil
	}

	ret := q.queuedOrder[0]
	fmt.Printf("\n=== Deque update: %q\n", ret.NodeID())
	q.queuedOrder = q.queuedOrder[1:]
	delete(q.queuedSet, ret)
	return ret
}
