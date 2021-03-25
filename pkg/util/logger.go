package util

import (
	"sync"

	"github.com/go-kit/kit/log"
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

	// cfg.Log wraps the log function, so we need to skip one extra stack from
	// than the default caller to get the caller information.
	cfg.Log = logging.GoKit(log.With(&l, "caller", log.Caller(4)))

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
	newLogger = log.With(newLogger, "caller", log.DefaultCaller)

	l.l = newLogger
	return nil
}

func defaultLogger(cfg *server.Config) (log.Logger, error) {
	var l log.Logger

	l, err := cortex_log.NewPrometheusLogger(cfg.LogLevel, cfg.LogFormat)
	if err != nil {
		return nil, err
	}

	return l, nil
}

// Log logs a log line.
func (l *Logger) Log(kvps ...interface{}) error {
	l.mut.RLock()
	defer l.mut.RUnlock()
	return l.l.Log(kvps...)
}
