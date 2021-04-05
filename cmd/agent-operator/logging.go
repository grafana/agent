package main

import (
	"os"

	cortex_log "github.com/cortexproject/cortex/pkg/util/log"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/go-logr/logr"
	controller "sigs.k8s.io/controller-runtime"
)

// setupLogger sets up our logger. If this function fails, the program will
// exit.
func setupLogger(l log.Logger, cfg *Config) log.Logger {
	newLogger, err := cortex_log.NewPrometheusLogger(cfg.LogLevel, cfg.LogFormat)
	if err != nil {
		level.Error(l).Log("msg", "failed to create logger", "err", err)
		os.Exit(1)
	}
	l = newLogger

	// Logs made globally are wrapped so they have a different caller level than
	// the manager logger does. We set these up first so our local logger can be
	// given the default caller depth.
	var (
		globalLogger  = &loggrAdapter{l: log.With(l, "caller", log.Caller(5))}
		managerLogger = &loggrAdapter{l: log.With(l, "caller", log.Caller(4))}
	)
	l = log.With(l, "caller", log.DefaultCaller)

	// Set up the global logger and the controller-local logger.
	controller.SetLogger(globalLogger)
	cfg.Controller.Logger = managerLogger
	return l
}

// loggrAdapter implements logr.Logger for go-kit logging. The only log levels
// are info and error, and verbosity information is ignored.
type loggrAdapter struct {
	name string
	l    log.Logger
}

func (a *loggrAdapter) Enabled() bool { return true }

func (a *loggrAdapter) Info(msg string, keysAndValues ...interface{}) {
	var args []interface{}
	if a.name != "" {
		args = append(args, "name", a.name)
	}
	args = append(args, "msg", msg)
	args = append(args, keysAndValues...)
	level.Info(a.l).Log(args...)
}

func (a *loggrAdapter) Error(err error, msg string, keysAndValues ...interface{}) {
	var args []interface{}
	if a.name != "" {
		args = append(args, "name", a.name)
	}
	args = append(args, "msg", msg, "err", err)
	args = append(args, keysAndValues...)
	level.Error(a.l).Log(args...)
}

func (a *loggrAdapter) V(level int) logr.Logger {
	return a
}

func (a *loggrAdapter) WithValues(keysAndValues ...interface{}) logr.Logger {
	return &loggrAdapter{name: a.name, l: log.With(a.l, keysAndValues...)}
}

func (a *loggrAdapter) WithName(name string) logr.Logger {
	newName := name
	if a.name != "" {
		newName = a.name + "." + name
	}
	return &loggrAdapter{name: newName, l: a.l}
}
