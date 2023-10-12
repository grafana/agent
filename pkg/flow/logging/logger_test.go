package logging_test

import (
	"fmt"
	"io"
	"log/slog"
	"testing"
	"time"

	gokitlevel "github.com/go-kit/log/level"
	"github.com/grafana/agent/pkg/flow/logging"
	"github.com/stretchr/testify/require"
)

/* Most recent performance results on M2 Macbook Air:
$ go test -count=1 -benchmem ./pkg/flow/logging -run ^$ -bench BenchmarkLogging_
goos: darwin
goarch: arm64
pkg: github.com/grafana/agent/pkg/flow/logging
BenchmarkLogging_NoLevel_Prints-8             	  811214	      1687 ns/op	     368 B/op	      11 allocs/op
BenchmarkLogging_NoLevel_Drops-8              	 4061053	       293.4 ns/op	      40 B/op	       4 allocs/op
BenchmarkLogging_GoKitLevel_Drops_Sprintf-8   	 2811063	       426.4 ns/op	     400 B/op	       9 allocs/op
BenchmarkLogging_GoKitLevel_Drops-8           	 2567967	       468.0 ns/op	     504 B/op	       9 allocs/op
BenchmarkLogging_GoKitLevel_Prints-8          	  710380	      1686 ns/op	     849 B/op	      16 allocs/op
BenchmarkLogging_Slog_Drops-8                 	78973346	        15.33 ns/op	       8 B/op	       0 allocs/op
BenchmarkLogging_Slog_Prints-8                	 1000000	      1128 ns/op	      32 B/op	       2 allocs/op
*/

const testStr = "this is a test string"

func BenchmarkLogging_NoLevel_Prints(b *testing.B) {
	logger, err := logging.New(io.Discard, debugLevelOptions())
	require.NoError(b, err)

	testErr := fmt.Errorf("test error")
	for i := 0; i < b.N; i++ {
		logger.Log("msg", "test message", "i", i, "err", testErr, "str", testStr, "duration", time.Second)
	}
}

func BenchmarkLogging_NoLevel_Drops(b *testing.B) {
	logger, err := logging.New(io.Discard, warnLevelOptions())
	require.NoError(b, err)

	testErr := fmt.Errorf("test error")
	for i := 0; i < b.N; i++ {
		logger.Log("msg", "test message", "i", i, "err", testErr, "str", testStr, "duration", time.Second)
	}
}

func BenchmarkLogging_GoKitLevel_Drops_Sprintf(b *testing.B) {
	logger, err := logging.New(io.Discard, debugLevelOptions())
	require.NoError(b, err)

	testErr := fmt.Errorf("test error")
	for i := 0; i < b.N; i++ {
		gokitlevel.Debug(logger).Log("msg", fmt.Sprintf("test message %d, error=%v, str=%s, duration=%v", i, testErr, testStr, time.Second))
	}
}

func BenchmarkLogging_GoKitLevel_Drops(b *testing.B) {
	logger, err := logging.New(io.Discard, debugLevelOptions())
	require.NoError(b, err)

	testErr := fmt.Errorf("test error")
	for i := 0; i < b.N; i++ {
		gokitlevel.Debug(logger).Log("msg", "test message", "i", i, "err", testErr, "str", testStr, "duration", time.Second)
	}
}

func BenchmarkLogging_GoKitLevel_Prints(b *testing.B) {
	logger, err := logging.New(io.Discard, debugLevelOptions())
	require.NoError(b, err)

	testErr := fmt.Errorf("test error")
	testStr := "this is a test string"
	for i := 0; i < b.N; i++ {
		gokitlevel.Warn(logger).Log("msg", "test message", "i", i, "err", testErr, "str", testStr, "duration", time.Second)
	}
}

func BenchmarkLogging_Slog_Drops(b *testing.B) {
	logger := slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	testErr := fmt.Errorf("test error")
	for i := 0; i < b.N; i++ {
		logger.Debug("test message", "i", i, "err", testErr, "str", testStr, "duration", time.Second)
	}
}

func BenchmarkLogging_Slog_Prints(b *testing.B) {
	logger := slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	testErr := fmt.Errorf("test error")
	for i := 0; i < b.N; i++ {
		logger.Info("test message", "i", i, "err", testErr, "str", testStr, "duration", time.Second)
	}
}

func debugLevelOptions() logging.Options {
	opts := logging.Options{}
	opts.SetToDefault()
	opts.Level = logging.LevelInfo
	return opts
}

func warnLevelOptions() logging.Options {
	opts := debugLevelOptions()
	opts.Level = logging.LevelWarn
	return opts
}
