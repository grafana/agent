package server

import (
	"sync"

	"github.com/go-kit/log"
	"github.com/weaveworks/common/logging"

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

	// HookLogger is used to temporarily hijack logs for support bundles.
	HookLogger HookLogger

	// makeLogger will default to defaultLogger. It's a struct
	// member to make testing work properly.
	makeLogger func(*Config) (log.Logger, error)
}

// HookLogger is used to temporarily redirect
type HookLogger struct {
	mut     sync.RWMutex
	enabled bool
	logger  log.Logger
}

// NewLogger creates a new Logger.
func NewLogger(cfg *Config) *Logger {
	return newLogger(cfg, defaultLogger)
}

// NewLoggerFromLevel creates a new logger from logging.Level and logging.Format.
func NewLoggerFromLevel(lvl logging.Level, fmt logging.Format) *Logger {
	logger, err := makeDefaultLogger(lvl, fmt)
	if err != nil {
		panic(err)
	}
	return &Logger{
		l: logger,
	}
}

func newLogger(cfg *Config, ctor func(*Config) (log.Logger, error)) *Logger {
	l := Logger{makeLogger: ctor}
	if err := l.ApplyConfig(cfg); err != nil {
		panic(err)
	}
	return &l
}

// ApplyConfig applies configuration changes to the logger.
func (l *Logger) ApplyConfig(cfg *Config) error {
	l.mut.Lock()
	defer l.mut.Unlock()

	newLogger, err := l.makeLogger(cfg)
	if err != nil {
		return err
	}

	l.l = newLogger
	return nil
}

func defaultLogger(cfg *Config) (log.Logger, error) {
	return makeDefaultLogger(cfg.LogLevel.Level, cfg.LogFormat)
}

func makeDefaultLogger(lvl logging.Level, fmt logging.Format) (log.Logger, error) {
	var l log.Logger

	l, err := cortex_log.NewPrometheusLogger(lvl, fmt)
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
	err := l.HookLogger.Log(kvps...)
	if err != nil {
		return err
	}
	return l.l.Log(kvps...)
}

// Log implements log.Logger.
func (hl *HookLogger) Log(kvps ...interface{}) error {
	hl.mut.RLock()
	defer hl.mut.RUnlock()
	if hl.enabled {
		return hl.logger.Log(kvps...)
	}
	return nil
}

// Set where HookedLogger should tee logs to.
// If a nil logger is passed, the HookedLogger is disabled.
func (hl *HookLogger) Set(l log.Logger) {
	hl.mut.Lock()
	defer hl.mut.Unlock()

	hl.enabled = l != nil
	hl.logger = l
}

// GoKitLogger creates a logging.Interface from a log.Logger.
func GoKitLogger(l log.Logger) logging.Interface {
	return logging.GoKit(l)
}
