package ratelimiting

import (
	"sync"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsRateLimited(t *testing.T) {
	const (
		rps      = 1
		burst    = 5
		requests = 15
	)

	var (
		wg  sync.WaitGroup
		sum = uint32(0)
	)

	rl := NewRateLimiter(rps, burst)

	op := func() {
		defer wg.Done()
		if ok := rl.IsRateLimited(); ok {
			atomic.AddUint32(&sum, 1)
		}
	}

	wg.Add(requests)
	for i := 0; i < requests; i++ {
		go op()
	}
	wg.Wait()

	assert.Equal(t, burst, int(sum))
}
