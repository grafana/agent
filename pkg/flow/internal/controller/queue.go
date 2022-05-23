package controller

import "sync"

// Queue is an unordered queue of components.
//
// Queue is intended for tracking components that have updated their Exports
// for later reevaluation.
type Queue struct {
	mut    sync.Mutex
	queued map[*ComponentNode]struct{}

	updateCh chan struct{}
}

// NewQueue returns a new unordered component queue.
func NewQueue() *Queue {
	return &Queue{
		updateCh: make(chan struct{}, 1),
		queued:   make(map[*ComponentNode]struct{}),
	}
}

// Enqueue inserts a new component into the Queue. Enqueue is a no-op if the
// component is already in the Queue.
func (q *Queue) Enqueue(c *ComponentNode) {
	q.mut.Lock()
	defer q.mut.Unlock()
	q.queued[c] = struct{}{}
	select {
	case q.updateCh <- struct{}{}:
	default:
	}
}

// Chan returns a channel which is written to when the queue is non-empty.
func (q *Queue) Chan() <-chan struct{} { return q.updateCh }

// TryDequeue dequeues a randomly queued component. TryDequeue will return nil
// if the queue is empty.
func (q *Queue) TryDequeue() *ComponentNode {
	q.mut.Lock()
	defer q.mut.Unlock()

	for c := range q.queued {
		delete(q.queued, c)
		return c
	}

	return nil
}
