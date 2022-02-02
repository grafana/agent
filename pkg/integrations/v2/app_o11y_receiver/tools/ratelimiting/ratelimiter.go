package ratelimiting

import (
	"fmt"
	"sync"
	"time"
)

// RateLimiter interface is responsible for controlling the rate of
// certain process. It implements the token bucket algorithm [1] with rate `r`
// and burstiness `b`.
//
// The RateLimiter allows for a `IsRateLimited() bool` method which returns
// whether the request should be completed or not
//
// [1]: https://en.wikipedia.org/wiki/Token_bucket
type RateLimiter struct {
	// Maximum allowed requests per second
	rps float64
	// Burstiness of the limit (bucket token size)
	b int
	// Mutex used to lock access between competing resources
	mu sync.Mutex
	// Number of available requests (token) in a given window
	tokens int
	// Timstamp of the last occurrence of an event
	lastEvent time.Time
}

// NewRateLimiter creates a new token bucket rate limiter
// of maximum rate _r_ and bustiness _b_
func NewRateLimiter(r float64, b int) *RateLimiter {
	return &RateLimiter{
		rps: r,
		b:   b,
		// Start the bucket full
		tokens: b,
	}
}

// IsRateLimited checks whether the action should be rate limited
// or not
func (rl *RateLimiter) IsRateLimited() (ok bool) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	last := rl.lastEvent

	fmt.Println("first time %w", last)
	if now.Before(last) {
		last = now
	}

	tokens := rl.tokens
	// Tokens accumulated over time before last request
	extraTokens := int(now.Sub(last).Seconds() * rl.rps)
	// If the token limit (bucket size) is exceeded, set the
	// tokens to the limit.
	if tokens += extraTokens; tokens > rl.b {
		tokens = rl.b
	}

	tokens--

	ok = tokens >= 0

	rl.lastEvent = now

	if ok {
		rl.tokens = tokens
	}

	return !ok
}
