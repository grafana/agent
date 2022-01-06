package util

import (
	"context"
	"sync"
	"time"

	"go.uber.org/atomic"
)

// WaitTrigger allows for waiting for a specific condition to be met.
// Useful for tests.
type WaitTrigger struct {
	completed atomic.Bool
	mut       *sync.Mutex
	cond      *sync.Cond
}

// NewWaitTrigger creates a new WaitTrigger.
func NewWaitTrigger() *WaitTrigger {
	var mut sync.Mutex
	cond := sync.NewCond(&mut)
	return &WaitTrigger{mut: &mut, cond: cond}
}

// Trigger completes the trigger and alerts all waiting. Calling Trigger again
// after the first invocation is a no-op.
func (wt *WaitTrigger) Trigger() {
	wt.mut.Lock()
	defer wt.mut.Unlock()
	wt.completed.Store(true)
	wt.cond.Broadcast()
}

// Wait waits for trigger to complete up to the specified timeout. Returns an
// error if the timeout expires.
func (wt *WaitTrigger) Wait(timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	go func() {
		<-ctx.Done()
		if ctx.Err() == context.DeadlineExceeded {
			// Ignore cancellations.
			wt.cond.Broadcast()
		}
	}()

	wt.mut.Lock()
	for ctx.Err() == nil && !wt.completed.Load() {
		wt.cond.Wait()
	}
	err := ctx.Err()
	wt.mut.Unlock()
	return err
}
