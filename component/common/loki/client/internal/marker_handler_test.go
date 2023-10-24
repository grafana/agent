package internal

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type mockMarkerFileHandler struct {
	lastMarkedSegment int
}

func (m *mockMarkerFileHandler) LastMarkedSegment() int {
	return m.lastMarkedSegment
}

func (m *mockMarkerFileHandler) MarkSegment(segment int) {
	m.lastMarkedSegment = segment
}

func TestMarkerHandler(t *testing.T) {
	t.Run("returns last marked segment from file handler on start", func(t *testing.T) {
		mockMFH := &mockMarkerFileHandler{lastMarkedSegment: 10}
		mh := NewMarkerHandler(mockMFH)
		defer mh.Stop()

		require.Equal(t, 10, mh.LastMarkedSegment())
	})

	t.Run("last marked segment is updated when sends complete", func(t *testing.T) {
		mockMFH := &mockMarkerFileHandler{lastMarkedSegment: 10}
		mh := NewMarkerHandler(mockMFH)
		defer mh.Stop()

		mh.UpdateReceivedData(11, 10)
		mh.UpdateSentData(11, 5)
		mh.UpdateSentData(11, 5)

		require.Eventually(t, func() bool {
			return mh.LastMarkedSegment() == 11
		}, time.Second, time.Millisecond*100, "expected last marked segment to catch up")
		require.Equal(t, 11, mockMFH.LastMarkedSegment())
	})
}
