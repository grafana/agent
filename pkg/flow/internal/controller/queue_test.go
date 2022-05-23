package controller

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEnqueueDequeue(t *testing.T) {
	tn := &ComponentNode{}
	q := NewQueue()
	q.Enqueue(tn)
	require.Lenf(t, q.queued, 1, "queue should be 1")
	fn := q.TryDequeue()
	require.True(t, fn == tn)
}
