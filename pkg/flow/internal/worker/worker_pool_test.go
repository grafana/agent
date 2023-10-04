package worker

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/atomic"
	"go.uber.org/goleak"
)

var defaultTestTimeout = 10 * time.Millisecond

func TestWorkerPool(t *testing.T) {
	t.Run("worker pool", func(t *testing.T) {
		t.Run("should start and stop cleanly", func(t *testing.T) {
			defer goleak.VerifyNone(t)
			pool := NewShardedWorkerPool(4, 1, defaultTestTimeout)
			require.Equal(t, 0, pool.QueueSize())
			defer pool.Stop()
		})

		t.Run("should reject invalid worker count", func(t *testing.T) {
			defer goleak.VerifyNone(t)

			require.Panics(t, func() {
				NewShardedWorkerPool(0, 0, defaultTestTimeout)
			})

			require.Panics(t, func() {
				NewShardedWorkerPool(-1, 0, defaultTestTimeout)
			})
		})

		t.Run("should reject invalid buffer size", func(t *testing.T) {
			defer goleak.VerifyNone(t)
			require.Panics(t, func() {
				NewShardedWorkerPool(1, -1, defaultTestTimeout)
			})
		})

		t.Run("should process a single task", func(t *testing.T) {
			defer goleak.VerifyNone(t)
			done := make(chan struct{})
			pool := NewShardedWorkerPool(4, 1, defaultTestTimeout)
			defer pool.Stop()

			err := pool.TrySubmit(func() {
				done <- struct{}{}
			}, pool.DefaultTimeout())
			require.NoError(t, err)
			select {
			case <-done:
				return
			case <-time.After(3 * time.Second):
				t.Fatal("timeout waiting for task to be processed")
				return
			}
		})

		t.Run("should process a single task with key", func(t *testing.T) {
			defer goleak.VerifyNone(t)
			done := make(chan struct{})
			pool := NewShardedWorkerPool(4, 1, defaultTestTimeout)
			defer pool.Stop()

			err := pool.TrySubmitWithKey("testKey", func() {
				done <- struct{}{}
			}, pool.DefaultTimeout())
			require.NoError(t, err)
			select {
			case <-done:
				return
			case <-time.After(3 * time.Second):
				t.Fatal("timeout waiting for task to be processed")
				return
			}
		})

		t.Run("should not queue duplicated keys", func(t *testing.T) {
			defer goleak.VerifyNone(t)
			pool := NewShardedWorkerPool(4, 10, defaultTestTimeout)
			defer pool.Stop()
			tasksDone := atomic.Int32{}

			// First task will block the worker
			blockFirstTask := make(chan struct{})
			firstTaskRunning := make(chan struct{})
			err := pool.TrySubmitWithKey("k1", func() {
				firstTaskRunning <- struct{}{}
				<-blockFirstTask
				tasksDone.Inc()
			}, pool.DefaultTimeout())
			require.NoError(t, err)

			// Wait for the first task to be running already and blocking the worker
			<-firstTaskRunning
			require.Equal(t, 0, pool.QueueSize())

			// Second task will be queued
			err = pool.TrySubmitWithKey("k1", func() {
				tasksDone.Inc()
			}, pool.DefaultTimeout())
			require.NoError(t, err)
			require.Equal(t, 1, pool.QueueSize())

			// Third task will be skipped, as we already have k1 in the queue
			err = pool.TrySubmitWithKey("k1", func() {
				tasksDone.Inc()
			}, pool.DefaultTimeout())
			require.NoError(t, err)

			// No tasks done yet as we're blocking the first task
			require.Equal(t, int32(0), tasksDone.Load())

			// After we unblock first task, two tasks should get done
			blockFirstTask <- struct{}{}
			require.Eventually(t, func() bool {
				return tasksDone.Load() == 2
			}, 3*time.Second, 5*time.Millisecond)
			require.Equal(t, 0, pool.QueueSize())

			// No more tasks should be done, verify again with some delay
			time.Sleep(100 * time.Millisecond)
			require.Equal(t, int32(2), tasksDone.Load())
		})

		t.Run("should concurrently process for different keys", func(t *testing.T) {
			defer goleak.VerifyNone(t)
			pool := NewShardedWorkerPool(4, 10, defaultTestTimeout)
			defer pool.Stop()
			tasksDone := atomic.Int32{}

			// First task will block the worker
			blockFirstTask := make(chan struct{})
			firstTaskRunning := make(chan struct{})
			err := pool.TrySubmitWithKey("k1", func() {
				firstTaskRunning <- struct{}{}
				<-blockFirstTask
				tasksDone.Inc()
			}, pool.DefaultTimeout())
			require.NoError(t, err)

			// Wait for the first task to be running already and blocking the worker
			<-firstTaskRunning

			// Second and third tasks will complete as it has a key that will hash to a different shard
			err = pool.TrySubmitWithKey("k2", func() { tasksDone.Inc() }, pool.DefaultTimeout())
			require.NoError(t, err)

			err = pool.TrySubmitWithKey("k3", func() { tasksDone.Inc() }, pool.DefaultTimeout())
			require.NoError(t, err)

			// Ensure the k2 and k3 tasks are done
			require.Eventually(t, func() bool {
				return tasksDone.Load() == 2
			}, 3*time.Second, 5*time.Millisecond)

			// After we unblock first task, it should get done as well
			blockFirstTask <- struct{}{}
			require.Eventually(t, func() bool {
				return tasksDone.Load() == 3
			}, 3*time.Second, 5*time.Millisecond)
		})

		t.Run("should reject when queue is full and times out waiting", func(t *testing.T) {
			defer goleak.VerifyNone(t)
			// Pool with one worker and queue size of 1 - all work goes to one queue
			pool := NewShardedWorkerPool(1, 1, defaultTestTimeout)
			defer pool.Stop()
			tasksDone := atomic.Int32{}

			// First task will block the worker
			blockFirstTask := make(chan struct{})
			firstTaskRunning := make(chan struct{})
			err := pool.TrySubmitWithKey("k1", func() {
				firstTaskRunning <- struct{}{}
				<-blockFirstTask
				tasksDone.Inc()
			}, pool.DefaultTimeout())
			require.NoError(t, err)
			defer func() { blockFirstTask <- struct{}{} }()

			// Wait for the first task to be running already and blocking the worker
			<-firstTaskRunning
			require.Equal(t, 0, pool.QueueSize())

			// Second task will be queued
			err = pool.TrySubmitWithKey("k2", func() { tasksDone.Inc() }, pool.DefaultTimeout())
			require.NoError(t, err)
			require.Equal(t, 1, pool.QueueSize())

			// Third task cannot be accepted, because the queue is full
			err = pool.TrySubmitWithKey("k3", func() { tasksDone.Inc() }, pool.DefaultTimeout())
			require.ErrorContains(t, err, "timed out adding a task to worker queue")
			require.Equal(t, 1, pool.QueueSize())
		})

		t.Run("should process when queue frees up within the timeout", func(t *testing.T) {
			defer goleak.VerifyNone(t)
			// Pool with one worker and queue size of 1 - all work goes to one queue
			pool := NewShardedWorkerPool(1, 1, defaultTestTimeout)
			defer pool.Stop()
			tasksDone := atomic.Int32{}

			// First task will block the worker
			blockFirstTask := make(chan struct{})
			firstTaskRunning := make(chan struct{})
			err := pool.TrySubmitWithKey("k1", func() {
				firstTaskRunning <- struct{}{}
				<-blockFirstTask
				tasksDone.Inc()
			}, pool.DefaultTimeout())
			require.NoError(t, err)

			// Wait for the first task to be running already and blocking the worker
			<-firstTaskRunning
			require.Equal(t, 0, pool.QueueSize())

			// Second task will be queued
			err = pool.TrySubmitWithKey("k2", func() { tasksDone.Inc() }, pool.DefaultTimeout())
			require.NoError(t, err)
			require.Equal(t, 1, pool.QueueSize())

			// Third task cannot be accepted right away with default timeout
			err = pool.TrySubmitWithKey("k3", func() { tasksDone.Inc() }, pool.DefaultTimeout())
			require.ErrorContains(t, err, "timed out adding a task to worker queue")
			require.Equal(t, 1, pool.QueueSize())

			// Unblock first task after some delay
			go func() {
				time.Sleep(10 * time.Millisecond)
				blockFirstTask <- struct{}{}
			}()

			// If we submit the third task again with a longer timeout, it will eventually be accepted, as we allowed the
			// first task to complete above.
			err = pool.TrySubmitWithKey("k3", func() { tasksDone.Inc() }, time.Second)
			require.NoError(t, err)
		})
	})
}
