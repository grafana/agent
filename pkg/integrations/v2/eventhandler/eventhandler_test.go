package eventhandler

import (
	"os"
	"testing"
	"time"

	"github.com/go-kit/log"
	"github.com/stretchr/testify/require"
)

// TODO: fix this
func TestCacheLoad(t *testing.T) {
	l := log.NewNopLogger()
	testTime, _ := time.Parse(time.RFC3339, "2022-01-20T17:12:58-06:00")
	expectedEvents := &ShippedEvents{
		Timestamp: testTime,
		RvMap:     map[string]struct{}{"22819": {}, "22820": {}, "22821": {}},
	}
	cacheFile, err := os.OpenFile("testdata/eventhandler.cache", os.O_RDWR|os.O_CREATE, cacheFileMode)
	require.NoError(t, err, "Failed to open test eventhandler cache file")
	actualEvents, err := readInitEvent(cacheFile, l)
	require.NoError(t, err, "Failed to parse last event from eventhandler cache file")
	require.Equal(t, expectedEvents, actualEvents)
}
