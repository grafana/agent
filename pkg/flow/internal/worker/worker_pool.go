package worker

import (
	"fmt"
	"hash/fnv"
	"math/rand"
	"runtime"
	"sync"
	"time"
)

type Pool interface {
	// Stop stops the worker pool. It does not wait to drain any internal queues, but it does wait for the currently
	// running tasks to complete. It must only be called once.
	Stop()
	// TrySubmit submits a function fn to be executed by the worker pool on a random worker. Error is returned if the
	// pool is unable to accept extra work within the provided timeout.
	TrySubmit(fn func(), timeout time.Duration) error
	// TrySubmitWithKey submits a function to be executed by the worker pool, ensuring that only one
	// job with given key can be queued at the time. Adding a job with a key that is already queued is a no-op (even if
	// the submitted function is different). Error is returned if the pool is unable to accept extra work within the
	// provided timeout - the caller can decide how to handle this situation.
	TrySubmitWithKey(key string, fn func(), timeout time.Duration) error
	// QueueSize returns the number of tasks currently queued.
	QueueSize() int
	// DefaultTimeout returns the default timeout that can be used with TrySubmit and TrySubmitWithKey.
	DefaultTimeout() time.Duration
}

// shardedWorkerPool is a Pool that distributes work across a fixed number of workers, using a hash of the key to
// determine which worker to use. This, to an extent, defends the pool from a slow task hogging all the workers.
type shardedWorkerPool struct {
	workersCount   int
	workQueues     []chan func()
	quit           chan struct{}
	allStopped     sync.WaitGroup
	defaultTimeout time.Duration

	lock sync.Mutex
	set  map[string]struct{}
}

var _ Pool = (*shardedWorkerPool)(nil)

// NewDefaultWorkerPool creates a new worker pool suitable for use in Flow controllers. Since Flow retries component
// dependency updates, the timeout used will mostly determine how often we'll log an error when the queue is full.
func NewDefaultWorkerPool() Pool {
	return NewShardedWorkerPool(runtime.NumCPU(), 1024, 10*time.Second)
}

// NewShardedWorkerPool creates a new worker pool with the given number of workers and queue size for each worker.
// The queued tasks are sharded across the workers using a hash of the key. The pool is automatically started and
// ready to accept work. To prevent resource leak, Stop() must be called when the pool is no longer needed.
func NewShardedWorkerPool(workersCount int, queueSize int, defaultTimeout time.Duration) Pool {
	if workersCount <= 0 {
		panic(fmt.Sprintf("workersCount must be positive, got %d", workersCount))
	}
	queues := make([]chan func(), workersCount)
	for i := 0; i < workersCount; i++ {
		queues[i] = make(chan func(), queueSize)
	}
	pool := &shardedWorkerPool{
		workersCount:   workersCount,
		workQueues:     queues,
		quit:           make(chan struct{}),
		set:            make(map[string]struct{}),
		defaultTimeout: defaultTimeout,
	}
	pool.start()
	return pool
}

func (w *shardedWorkerPool) TrySubmit(f func(), timeout time.Duration) error {
	return w.TrySubmitWithKey(fmt.Sprintf("%d", rand.Int()), f, timeout)
}

func (w *shardedWorkerPool) TrySubmitWithKey(key string, f func(), timeout time.Duration) error {
	wrapped := func() {
		// NOTE: we intentionally remove from the queue before executing the function. This means that while a task is
		// executing, another task with the same key can be added to the queue, potentially even by the task itself.
		w.lock.Lock()
		delete(w.set, key)
		w.lock.Unlock()

		f()
	}
	queue := w.workQueues[w.workerFor(key)]
	deadline := time.Now().Add(timeout)

	for {
		w.lock.Lock()
		if _, exists := w.set[key]; exists {
			w.lock.Unlock()
			return nil // Allow only queue one job for given key
		}

		select {
		case queue <- wrapped:
			w.set[key] = struct{}{}
			w.lock.Unlock()
			return nil // We successfully queued the task
		default:
			if time.Now().After(deadline) {
				w.lock.Unlock()
				// We've timed out trying to queue the task
				return fmt.Errorf("timed out adding a task to worker queue with key %q after %q", key, timeout)
			}
		}
		// Couldn't queue the task, release the lock and let other goroutines work on the tasks
		w.lock.Unlock()
		runtime.Gosched()
	}
}

func (w *shardedWorkerPool) QueueSize() int {
	w.lock.Lock()
	defer w.lock.Unlock()
	return len(w.set)
}

func (w *shardedWorkerPool) DefaultTimeout() time.Duration {
	return w.defaultTimeout
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
