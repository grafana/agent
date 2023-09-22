package controller

import (
	"fmt"
	"hash/fnv"
)

type stripedWorkerPool struct {
	workersCount int
	workQueues   []chan func()
	quit         chan struct{}
}

func newStripedWorkerPool(workersCount int, queueSize int) *stripedWorkerPool {
	queues := make([]chan func(), workersCount)
	for i := 0; i < workersCount; i++ {
		queues[i] = make(chan func(), queueSize)
	}
	pool := &stripedWorkerPool{
		workersCount: workersCount,
		workQueues:   queues,
		quit:         make(chan struct{}),
	}
	pool.Start()
	return pool
}

func (w *stripedWorkerPool) Start() {
	for i := 0; i < w.workersCount; i++ {
		queue := w.workQueues[i]
		go func() {
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

func (w *stripedWorkerPool) Stop() {
	close(w.quit)
}

func (w *stripedWorkerPool) AddWork(key string, f func()) error {
	queue := w.workQueues[w.workerFor(key)]
	select {
	case queue <- f:
		return nil
	default:
		return fmt.Errorf("worker queue is full")
	}
}

func (w *stripedWorkerPool) workerFor(s string) int {
	h := fnv.New32a()
	_, _ = h.Write([]byte(s))
	return int(h.Sum32()) % w.workersCount
}
