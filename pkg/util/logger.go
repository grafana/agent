package util

import (
	cortex_log "github.com/cortexproject/cortex/pkg/util/log"
	"github.com/go-kit/kit/log"
	"github.com/weaveworks/common/logging"
	"github.com/weaveworks/common/server"
)

// NewLogger creates a new Logger.
func NewLogger(cfg *server.Config) *Logger {
	return newLogger(cfg, defaultLogger)
}

func newLogger(cfg *server.Config, ctor func(*server.Config) (log.Logger, error)) *Logger {
	l := Logger{makeLogger: ctor}
	if err := l.ApplyConfig(cfg); err != nil {
		panic(err)
	}

	// cfg.Log wraps the log function, so we need to skip one extra stack from
	// than the default caller to get the caller information.
	cfg.Log = logging.GoKit(log.With(&l, "caller", log.Caller(5)))

	return &l
}

func defaultLogger(cfg *server.Config) (log.Logger, error) {
	var l log.Logger

	l, err := cortex_log.NewPrometheusLogger(cfg.LogLevel, cfg.LogFormat)
	if err != nil {
		return nil, err
	}

	return l, nil
}
