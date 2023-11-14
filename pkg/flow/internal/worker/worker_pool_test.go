package worker

import (
	"container/list"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/atomic"
	"go.uber.org/goleak"
)

func TestWorkerPool(t *testing.T) {
	t.Run("worker pool", func(t *testing.T) {
		t.Run("should start and stop cleanly", func(t *testing.T) {
			defer goleak.VerifyNone(t)
			pool := NewFixedWorkerPool(4, 1)
			require.Equal(t, 0, pool.QueueSize())
			defer pool.Stop()
		})

		t.Run("should reject invalid worker count", func(t *testing.T) {
			defer goleak.VerifyNone(t)

			require.Panics(t, func() {
				NewFixedWorkerPool(0, 0)
			})

			require.Panics(t, func() {
				NewFixedWorkerPool(-1, 0)
			})
		})

		t.Run("should reject invalid buffer size", func(t *testing.T) {
			defer goleak.VerifyNone(t)
			require.Panics(t, func() {
				NewFixedWorkerPool(1, -1)
			})
		})

		t.Run("should process a single task", func(t *testing.T) {
			defer goleak.VerifyNone(t)
			done := make(chan struct{})
			pool := NewFixedWorkerPool(4, 1)
			defer pool.Stop()

			err := pool.SubmitWithKey("123", func() {
				done <- struct{}{}
			})
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
			pool := NewFixedWorkerPool(4, 1)
			defer pool.Stop()

			err := pool.SubmitWithKey("testKey", func() {
				done <- struct{}{}
			})
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
			pool := NewFixedWorkerPool(4, 10)
			defer pool.Stop()
			tasksDone := atomic.Int32{}

			// First task will block the worker
			blockFirstTask := make(chan struct{})
			firstTaskRunning := make(chan struct{})
			err := pool.SubmitWithKey("k1", func() {
				firstTaskRunning <- struct{}{}
				<-blockFirstTask
				tasksDone.Inc()
			})
			require.NoError(t, err)
			defer func() { close(blockFirstTask) }()

			// Wait for the first task to be running already and blocking the worker
			<-firstTaskRunning
			require.Equal(t, 1, pool.QueueSize())

			// Second task will be queued
			err = pool.SubmitWithKey("k1", func() {
				tasksDone.Inc()
			})
			require.NoError(t, err)
			require.Equal(t, 2, pool.QueueSize())

			// Third task will be skipped, as we already have k1 in the queue
			err = pool.SubmitWithKey("k1", func() {
				tasksDone.Inc()
			})
			require.NoError(t, err)

			// No tasks done yet as we're blocking the first task
			require.Equal(t, int32(0), tasksDone.Load())

			// After we unblock first task, two tasks should get done
			blockFirstTask <- struct{}{}
			require.Eventually(t, func() bool {
				return tasksDone.Load() == 2
			}, 3*time.Second, 1*time.Millisecond)
			require.Equal(t, 0, pool.QueueSize())

			// No more tasks should be done, verify again with some delay
			time.Sleep(10 * time.Millisecond)
			require.Equal(t, int32(2), tasksDone.Load())
		})

		t.Run("should concurrently process for different keys", func(t *testing.T) {
			defer goleak.VerifyNone(t)
			pool := NewFixedWorkerPool(4, 10)
			defer pool.Stop()
			tasksDone := atomic.Int32{}

			// First task will block the worker
			blockFirstTask := make(chan struct{})
			firstTaskRunning := make(chan struct{})
			err := pool.SubmitWithKey("k1", func() {
				firstTaskRunning <- struct{}{}
				<-blockFirstTask
				tasksDone.Inc()
			})
			require.NoError(t, err)
			defer func() { close(blockFirstTask) }()

			// Wait for the first task to be running already and blocking the worker
			<-firstTaskRunning

			// Second and third tasks will complete as it has a key that will hash to a different shard
			err = pool.SubmitWithKey("k2", func() { tasksDone.Inc() })
			require.NoError(t, err)

			err = pool.SubmitWithKey("k3", func() { tasksDone.Inc() })
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

		t.Run("should reject when queue is full", func(t *testing.T) {
			defer goleak.VerifyNone(t)
			// Pool with one worker and queue size of 1 - all work goes to one queue
			pool := NewFixedWorkerPool(1, 2)
			defer pool.Stop()
			tasksDone := atomic.Int32{}

			// First task will block the worker
			blockFirstTask := make(chan struct{})
			firstTaskRunning := make(chan struct{})
			err := pool.SubmitWithKey("k1", func() {
				firstTaskRunning <- struct{}{}
				<-blockFirstTask
				tasksDone.Inc()
			})
			require.NoError(t, err)
			defer func() { close(blockFirstTask) }()

			// Wait for the first task to be running already and blocking the worker
			<-firstTaskRunning
			require.Equal(t, 1, pool.QueueSize())

			// Second task will be queued
			err = pool.SubmitWithKey("k2", func() { tasksDone.Inc() })
			require.NoError(t, err)
			require.Equal(t, 2, pool.QueueSize())

			// Third task cannot be accepted, because the queue is full
			err = pool.SubmitWithKey("k3", func() { tasksDone.Inc() })
			require.ErrorContains(t, err, "queue is full")
			require.Equal(t, 2, pool.QueueSize())
		})

		t.Run("should not block when one task is stuck", func(t *testing.T) {
			defer goleak.VerifyNone(t)
			tasksCount := 1000

			// Queue size is sufficient to queue all tasks
			pool := NewFixedWorkerPool(3, tasksCount+1)
			defer pool.Stop()
			tasksDone := atomic.Int32{}

			// First task will block
			blockFirstTask := make(chan struct{})
			firstTaskRunning := make(chan struct{})
			err := pool.SubmitWithKey("k-blocking", func() {
				firstTaskRunning <- struct{}{}
				<-blockFirstTask
				tasksDone.Inc()
			})
			require.NoError(t, err)
			defer func() { close(blockFirstTask) }()

			// Wait for the first task to be running already and blocking the worker
			<-firstTaskRunning
			require.Equal(t, 1, pool.QueueSize())

			// Submit a lot of tasks with random keys - no task should be blocked by the above one.
			for i := 0; i < tasksCount; i++ {
				err = pool.SubmitWithKey(fmt.Sprintf("t%d", i), func() { tasksDone.Inc() })
				require.NoError(t, err)
			}

			// Ensure all tasks are done
			require.Eventually(t, func() bool {
				return tasksDone.Load() == int32(tasksCount)
			}, 3*time.Second, 1*time.Millisecond)
		})

		t.Run("should NOT run concurrently tasks with the same key", func(t *testing.T) {
			defer goleak.VerifyNone(t)
			tasksCount := 1000

			// Queue size is sufficient to queue all tasks
			pool := NewFixedWorkerPool(10, 10)
			defer pool.Stop()
			tasksDone := atomic.Int32{}

			// First task will block
			blockFirstTask := make(chan struct{})
			firstTaskRunning := make(chan struct{})
			err := pool.SubmitWithKey("k1", func() {
				firstTaskRunning <- struct{}{}
				<-blockFirstTask
				tasksDone.Inc()
			})
			require.NoError(t, err)
			defer func() { close(blockFirstTask) }()

			// Wait for the first task to be running already and blocking the worker
			<-firstTaskRunning
			require.Equal(t, 1, pool.QueueSize())

			// Enqueue one more task with the same key - it should be allowed
			err = pool.SubmitWithKey("k1", func() { tasksDone.Inc() })
			require.NoError(t, err)

			// Submit a lot of tasks with same key - all should be a no-op, since this task is already in queue
			for i := 0; i < tasksCount; i++ {
				err = pool.SubmitWithKey("k1", func() { tasksDone.Inc() })
				require.NoError(t, err)
			}

			require.Equal(t, int32(0), tasksDone.Load())

			// Unblock the first task
			blockFirstTask <- struct{}{}

			// The first task and the second one should be the only ones that complete
			require.Eventually(t, func() bool {
				return tasksDone.Load() == 2
			}, 3*time.Second, 1*time.Millisecond)
		})
	})
}

func BenchmarkQueue(b *testing.B) {
	/* The slice-based implementation is faster when queue size is less than 300 elements, as it doesn't allocate:

	BenchmarkQueue/slice_10_elements-8         	266251536	        4.427 ns/op        0 B/op	       0 allocs/op
	BenchmarkQueue/list_10_elements-8          	38036170	        31.20 ns/op	      56 B/op	       1 allocs/op
	BenchmarkQueue/slice_100_elements-8        	85725766	        14.29 ns/op	       0 B/op	       0 allocs/op
	BenchmarkQueue/list_100_elements-8         	37889650	        31.17 ns/op	      56 B/op	       1 allocs/op
	BenchmarkQueue/slice_300_elements-8        	40504732	        29.55 ns/op	       0 B/op	       0 allocs/op
	BenchmarkQueue/list_300_elements-8         	38032604	        31.20 ns/op	      56 B/op	       1 allocs/op
	BenchmarkQueue/slice_1000_elements-8       	12571960	        95.50 ns/op	       0 B/op	       0 allocs/op
	BenchmarkQueue/list_1000_elements-8        	37922080	        31.07 ns/op	      56 B/op	       2 allocs/op
	BenchmarkQueue/slice_10000_elements-8      	 1000000	      1026 ns/op	       0 B/op	       0 allocs/op
	BenchmarkQueue/list_10000_elements-8       	34379028	        34.24 ns/op	      56 B/op	       1 allocs/op
	*/

	queueSizes := []int{10, 100, 300, 1000, 10000}
	for _, queueSize := range queueSizes {
		elementsStr := fmt.Sprintf("%d elements", queueSize)
		b.Run("slice "+elementsStr, func(b *testing.B) {
			var slice []int
			for i := 0; i < queueSize; i++ {
				slice = append(slice, i)
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				// simulate what the `workQueue` does with its `waitingOrder` field.

				// add to queue
				slice = append(slice, i)

				// iterate to an arbitrary element
				ind := 0
				for ; ind < 5; ind++ {
					_ = slice[ind]
				}

				// remove it from the queue
				slice = append(slice[:ind], slice[ind+1:]...)
			}
		})

		b.Run("list "+elementsStr, func(b *testing.B) {
			l := list.New()
			for i := 0; i < queueSize; i++ {
				l.PushBack(i)
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				// simulate what the `workQueue` does with its `waitingOrder` field.

				// add to queue
				l.PushBack(i)

				// iterate to an arbitrary element
				toRemove := l.Front()
				for j := 0; j < 5; j++ {
					toRemove = toRemove.Next()
				}
				// remove it from the queue using copy
				l.Remove(toRemove)
			}
		})
	}
}
