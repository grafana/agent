package flow

import (
	"context"
	"sync"
)

// updateQueue is an unordered queue of updated components to be processed.
type updateQueue struct {
	mut     sync.Mutex
	updated []*componentNode

	updateCh chan struct{}
}

func newUpdateQueue() *updateQueue {
	return &updateQueue{
		updateCh: make(chan struct{}, 1),
	}
}

// Enqueue enqueues a new componentNode to be dequeued later.
func (uq *updateQueue) Enqueue(cn *componentNode) {
	uq.mut.Lock()
	uq.updated = append(uq.updated, cn)
	uq.mut.Unlock()

	select {
	case uq.updateCh <- struct{}{}:
	default:
	}
}

// Dequeue dequeues a componentNode from the queue. If the queue is empty,
// Dequeue blocks until there is an element to dequeue or until ctx is
// canceled.
func (uq *updateQueue) Dequeue(ctx context.Context) (*componentNode, error) {
Start:
	// Try to dequeue immediately if there's something in the queue.
	if elem := uq.dequeue(); elem != nil {
		return elem, nil
	}

	// Otherwise, wait for updateCh to be readable.
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-uq.updateCh:
		if elem := uq.dequeue(); elem != nil {
			return elem, nil
		}
		goto Start
	}
}

func (uq *updateQueue) dequeue() *componentNode {
	uq.mut.Lock()
	defer uq.mut.Unlock()

	if len(uq.updated) == 0 {
		return nil
	}

	res := uq.updated[len(uq.updated)-1]
	uq.updated = uq.updated[:len(uq.updated)-1]
	return res
}
