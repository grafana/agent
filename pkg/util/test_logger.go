package util

import (
	"os"
	"testing"
	"time"

	"github.com/go-kit/log"
	"github.com/grafana/agent/pkg/flow/logging/v2"
	"github.com/stretchr/testify/require"
)

// TestLogger generates a logger for a test.
func TestLogger(t *testing.T) log.Logger {
	t.Helper()

	l := log.NewSyncLogger(log.NewLogfmtLogger(os.Stderr))
	l = log.WithPrefix(l,
		"test", t.Name(),
		"ts", log.Valuer(testTimestamp),
	)

	return l
}

// TestFlowLogger generates a Flow-compatible logger for a test.
func TestFlowLogger(t require.TestingT) *logging.Logger {
	if t, ok := t.(*testing.T); ok {
		t.Helper()
	}

	sink, err := logging.WriterSink(os.Stderr, logging.SinkOptions{
		Level:  logging.LevelDebug,
		Format: logging.FormatLogfmt,
	})
	require.NoError(t, err)

	return logging.New(sink)
}

// testTimestamp is a log.Valuer that returns the timestamp
// without the date or timezone, reducing the noise in the test.
func testTimestamp() interface{} {
	t := time.Now().UTC()
	return t.Format("15:04:05.000")
}
