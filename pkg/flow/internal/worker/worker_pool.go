package worker

import (
	"fmt"
	"hash/fnv"
	"math/rand"
	"runtime"
	"sync"
)

type Pool interface {
	// Stop stops the worker pool. It does not wait to drain any internal queues, but it does wait for the currently
	// running tasks to complete. It must only be called once.
	Stop()
	// Submit submits a function to be executed by the worker pool on a random worker. Error is returned if the pool
	// is unable to accept extra work.
	Submit(func()) error
	// SubmitWithKey submits a function to be executed by the worker pool, ensuring that only one
	// job with given key can be queued at the time. Adding a job with a key that is already queued is a no-op (even if
	// the submitted function is different). Error is returned if the pool is unable to accept extra work - the caller
	// can decide how to handle this situation.
	SubmitWithKey(string, func()) error
	// QueueSize returns the number of tasks currently queued.
	QueueSize() int
}

// shardedWorkerPool is a Pool that distributes work across a fixed number of workers, using a hash of the key to
// determine which worker to use. This, to an extent, defends the pool from a slow task hogging all the workers.
type shardedWorkerPool struct {
	workersCount int
	workQueues   []chan func()
	quit         chan struct{}
	allStopped   sync.WaitGroup

	lock sync.Mutex
	set  map[string]struct{}
}

var _ Pool = (*shardedWorkerPool)(nil)

func NewDefaultWorkerPool() Pool {
	return NewShardedWorkerPool(runtime.NumCPU(), 1024)
}

// NewShardedWorkerPool creates a new worker pool with the given number of workers and queue size for each worker.
// The queued tasks are sharded across the workers using a hash of the key. The pool is automatically started and
// ready to accept work. To prevent resource leak, Stop() must be called when the pool is no longer needed.
func NewShardedWorkerPool(workersCount int, queueSize int) Pool {
	if workersCount <= 0 {
		panic(fmt.Sprintf("workersCount must be positive, got %d", workersCount))
	}
	queues := make([]chan func(), workersCount)
	for i := 0; i < workersCount; i++ {
		queues[i] = make(chan func(), queueSize)
	}
	pool := &shardedWorkerPool{
		workersCount: workersCount,
		workQueues:   queues,
		quit:         make(chan struct{}),
		set:          make(map[string]struct{}),
	}
	pool.start()
	return pool
}

func (w *shardedWorkerPool) Submit(f func()) error {
	return w.SubmitWithKey(fmt.Sprintf("%d", rand.Int()), f)
}

func (w *shardedWorkerPool) SubmitWithKey(key string, f func()) error {
	wrapped := func() {
		// NOTE: we intentionally remove from the queue before executing the function. This means that while a task is
		// executing, another task with the same key can be added to the queue, potentially even by the task itself.
		w.lock.Lock()
		delete(w.set, key)
		w.lock.Unlock()

		f()
	}
	queue := w.workQueues[w.workerFor(key)]

	w.lock.Lock()
	defer w.lock.Unlock()
	if _, exists := w.set[key]; exists {
		return nil // only queue one job for given key
	}

	select {
	case queue <- wrapped:
		w.set[key] = struct{}{}
		return nil
	default:
		return fmt.Errorf("worker queue is full")
	}
}

func (w *shardedWorkerPool) QueueSize() int {
	w.lock.Lock()
	defer w.lock.Unlock()
	return len(w.set)
}

func (w *shardedWorkerPool) Stop() {
	close(w.quit)
	w.allStopped.Wait()
}

func (w *shardedWorkerPool) start() {
	for i := 0; i < w.workersCount; i++ {
		queue := w.workQueues[i]
		w.allStopped.Add(1)
		go func() {
			defer w.allStopped.Done()
			for {
				select {
				case <-w.quit:
					return
				case f := <-queue:
					f()
				}
			}
		}()
	}
}

func (w *shardedWorkerPool) workerFor(s string) int {
	h := fnv.New32a()
	_, _ = h.Write([]byte(s))
	return int(h.Sum32()) % w.workersCount
}
