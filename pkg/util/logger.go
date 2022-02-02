package util

import (
	"sync"

	"github.com/go-kit/log"
	"github.com/weaveworks/common/logging"
	"github.com/weaveworks/common/server"

	cortex_log "github.com/cortexproject/cortex/pkg/util/log"
)

// Logger implements Go Kit's log.Logger interface. It supports being
// dynamically updated at runtime.
type Logger struct {
	// mut protects against race conditions accessing l, which can be modified
	// and accessed concurrently if ApplyConfig and Log are called at the same
	// time.
	mut sync.RWMutex
	l   log.Logger

	// makeLogger will default to defaultLogger. It's a struct
	// member to make testing work properly.
	makeLogger func(*server.Config) (log.Logger, error)
}

// NewLogger creates a new Logger.
func NewLogger(cfg *server.Config) *Logger {
	return newLogger(cfg, defaultLogger)
}

func newLogger(cfg *server.Config, ctor func(*server.Config) (log.Logger, error)) *Logger {
	l := Logger{makeLogger: ctor}
	if err := l.ApplyConfig(cfg); err != nil {
		panic(err)
	}
	return &l
}

// ApplyConfig applies configuration changes to the logger.
func (l *Logger) ApplyConfig(cfg *server.Config) error {
	l.mut.Lock()
	defer l.mut.Unlock()

	newLogger, err := l.makeLogger(cfg)
	if err != nil {
		return err
	}

	l.l = newLogger
	return nil
}

func defaultLogger(cfg *server.Config) (log.Logger, error) {
	var l log.Logger

	l, err := cortex_log.NewPrometheusLogger(cfg.LogLevel, cfg.LogFormat)
	if err != nil {
		return nil, err
	}

	// There are two wrappers on the log so skip two extra stacks vs default
	return log.With(l, "caller", log.Caller(5)), nil
}

// Log logs a log line.
func (l *Logger) Log(kvps ...interface{}) error {
	l.mut.RLock()
	defer l.mut.RUnlock()
	return l.l.Log(kvps...)
}

// GoKitLogger creates a logging.Interface from a log.Logger.
func GoKitLogger(l log.Logger) logging.Interface {
	return logging.GoKit(l)
}
