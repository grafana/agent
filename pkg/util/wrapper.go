package util

import (
	"sync"

	"github.com/go-kit/kit/log"
	"github.com/weaveworks/common/server"
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

// Log logs a log line.
func (l *Logger) Log(kvps ...interface{}) error {
	l.mut.RLock()
	defer l.mut.RUnlock()
	return l.l.Log(kvps...)
}
