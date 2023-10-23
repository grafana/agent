package level

import (
	"context"
	"log/slog"

	"github.com/go-kit/log"
	"github.com/grafana/agent/pkg/flow/logging"
)

const (
	levelKey = "level"
)

// Error returns a logger that includes a Key/ErrorValue pair.
func Error(logger log.Logger) log.Logger {
	return toLevel(logger, "error", slog.LevelError)
}

// Warn returns a logger that includes a Key/WarnValue pair.
func Warn(logger log.Logger) log.Logger {
	return toLevel(logger, "warn", slog.LevelWarn)
}

// Info returns a logger that includes a Key/InfoValue pair.
func Info(logger log.Logger) log.Logger {
	return toLevel(logger, "info", slog.LevelInfo)
}

// Debug returns a logger that includes a Key/DebugValue pair.
func Debug(logger log.Logger) log.Logger {
	return toLevel(logger, "debug", slog.LevelDebug)
}

func toLevel(logger log.Logger, level string, slogLevel slog.Level) log.Logger {
	switch l := logger.(type) {
	case logging.EnabledAware:
		if !l.Enabled(context.Background(), slogLevel) {
			return disabledLogger
		}
	}
	return log.WithPrefix(logger, levelKey, level)
}

var disabledLogger = &noopLogger{}

type noopLogger struct{}

func (d *noopLogger) Log(_ ...interface{}) error {
	return nil
}
