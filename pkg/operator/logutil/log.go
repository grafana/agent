// Package logutil implements an adaptor for the go-kit logger, which is used in the
// Grafana Agent project, and go-logr, which is used in controller-runtime.
package logutil

import (
	"context"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/go-logr/logr"
	clog "sigs.k8s.io/controller-runtime/pkg/log"
)

// Wrap wraps a log.Logger into a logr.Logger.
func Wrap(l log.Logger) logr.Logger {
	return &goKitLogger{l: l}
}

// FromContext returns a log.Logger from a context. Panics if the context doesn't
// have a Logger set.
func FromContext(ctx context.Context, kvps ...interface{}) log.Logger {
	gkl := clog.FromContext(ctx, kvps...).(*goKitLogger)
	return gkl.namedLogger()
}

type goKitLogger struct {
	// name is a name field used by logr which can be appended to dynamically.
	name string
	kvps []interface{}
	l    log.Logger
}

// namedLogger gets log.Logger with component applied.
func (l *goKitLogger) namedLogger() log.Logger {
	logger := l.l
	if l.name != "" {
		logger = log.With(logger, "component", l.name)
	}
	logger = log.With(logger, l.kvps...)
	return logger
}

func (l *goKitLogger) Log(keyvals ...interface{}) error {
	return l.namedLogger().Log(keyvals...)
}

func (l *goKitLogger) Enabled() bool { return true }

func (l *goKitLogger) Info(msg string, keysAndValues ...interface{}) {
	args := append([]interface{}{"msg", msg}, keysAndValues...)
	level.Info(l.namedLogger()).Log(args...)
}

func (l *goKitLogger) Error(err error, msg string, keysAndValues ...interface{}) {
	args := append([]interface{}{"msg", msg, "err", err}, keysAndValues...)
	level.Error(l.namedLogger()).Log(args...)
}

func (l *goKitLogger) V(level int) logr.Logger { return l }

func (l *goKitLogger) WithValues(keysAndValues ...interface{}) logr.Logger {
	return &goKitLogger{name: l.name, l: l.l, kvps: append(l.kvps, keysAndValues...)}
}

func (l *goKitLogger) WithName(name string) logr.Logger {
	newName := name
	if l.name != "" {
		newName = l.name + "." + name
	}
	return &goKitLogger{name: newName, l: l.l}
}
