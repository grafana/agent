package servicegraphprocessor

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var noopUpsertCb storeCallback = func(e *edge) {}

func TestStore_upsertEdge(t *testing.T) {
	const keyStr = "key"

	var cbCallCount int
	s := newStore(time.Hour, 100, func(e *edge) {
		cbCallCount++
	})
	assert.Equal(t, 0, s.len())

	s.upsertEdge(keyStr, func(e *edge) {})
	assert.Equal(t, 1, s.len())
	assert.False(t, s.shouldEvictHead()) // ttl is set to 1h
	assert.Equal(t, 0, cbCallCount)

	e := getEdge(s, keyStr)
	assert.NotNil(t, e)
	assert.Equal(t, keyStr, e.key)

	s.upsertEdge(keyStr, func(e *edge) {
		e.clientService = "client"
		e.serverService = "server"
		e.expiration = 0 // expire immediately
	})
	assert.Equal(t, 0, cbCallCount)

	e = getEdge(s, keyStr)
	assert.NotNil(t, e)
	assert.Equal(t, "client", e.clientService)
	assert.Equal(t, "server", e.serverService)
	assert.True(t, s.shouldEvictHead())

	s.evictHead()
	assert.Equal(t, 0, s.len())
	assert.Equal(t, 1, cbCallCount)
}

func TestStore_expire(t *testing.T) {
	keys := map[string]bool{}
	for i := 0; i < 100; i++ {
		keys[fmt.Sprintf("key-%d", i)] = true
	}

	// all new keys are immediately expired.
	s := newStore(-time.Second, 100, func(e *edge) {
		assert.True(t, keys[e.key])
	})

	for key := range keys {
		s.upsertEdge(key, noopUpsertCb)
	}

	s.expire()
	assert.Equal(t, 0, s.len())
}

func getEdge(s *store, k string) *edge {
	ele, ok := s.m[k]
	if !ok {
		return nil
	}
	return ele.Value.(*edge)
}
