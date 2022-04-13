package ratelimiting

import (
	"sync"
	"testing"

	"go.uber.org/atomic"

	"github.com/stretchr/testify/require"
)

func TestIsRateLimited(t *testing.T) {
	const (
		rps      = 1
		burst    = 5
		requests = 50
	)

	var (
		wg  sync.WaitGroup
		sum = atomic.NewUint32(0)
	)

	rl := NewRateLimiter(rps, burst)

	op := func() {
		defer wg.Done()
		if ok := rl.IsRateLimited(); !ok {
			sum.Add(1)
		}
	}

	wg.Add(requests)
	for i := 0; i < requests; i++ {
		go op()
	}
	wg.Wait()

	require.Equal(t, burst, int(sum.Load()))
}
