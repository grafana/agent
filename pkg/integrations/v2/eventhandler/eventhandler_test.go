package eventhandler

import (
	"os"
	"testing"
	"time"

	"github.com/go-kit/log"
	"github.com/stretchr/testify/require"
)

func TestCacheLoad(t *testing.T) {
	l := log.NewNopLogger()
	testTime, _ := time.Parse(time.RFC3339, "2022-01-26T13:39:40-05:00")
	expectedEvents := &ShippedEvents{
		Timestamp: testTime,
		RvMap:     map[string]struct{}{"58588": {}},
	}
	cacheFile, err := os.OpenFile("testdata/eventhandler.cache", os.O_RDWR|os.O_CREATE, cacheFileMode)
	require.NoError(t, err, "Failed to open test eventhandler cache file")
	actualEvents, err := readInitEvent(cacheFile, l)
	require.NoError(t, err, "Failed to parse last event from eventhandler cache file")
	require.Equal(t, expectedEvents, actualEvents)
}
