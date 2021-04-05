package util

import (
	"errors"
	"fmt"
	"runtime"
	"strings"

	"github.com/go-kit/kit/log/level"

	"github.com/go-kit/kit/log"
	"github.com/weaveworks/common/server"
	el "golang.org/x/sys/windows/svc/eventlog"
)

func NewWinFmtLogger(cfg *server.Config) *Logger {
	l := Logger{makeLogger: makeWinLogger}
	if err := l.ApplyConfig(cfg); err != nil {
		panic(err)
	}
	return &l
}

func makeWinLogger(cfg *server.Config) (log.Logger, error) {
	// error is necessary for the Log function to return an error
	notAllowedError := errors.New("not_allowed")

	// Setup the log in windows events
	err := el.InstallAsEventCreate("Grafana Agent", el.Error|el.Info|el.Warning)

	// We expect an error of 'already exists' for subsequent runs,
	if err != nil && !strings.Contains(err.Error(), "already exists") {
		return nil, err
	}
	il, err := el.Open("Grafana Agent")
	if err != nil {
		return nil, err
	}

	// Cleanup the handle when exits scope, this could be handled via explicit close but would need to change
	// more upstream, and its not a huge issue if it hangs around slightly longer than it should
	runtime.SetFinalizer(il, func(l *el.Log) {
		l.Close()
	})

	//  These are setup to be writers for each Windows log level
	//  Setup this way so we can utilize all the benefits of logformatter
	infoWriter := &winLogWriter{writer: func(p []byte) {
		il.Info(1, string(p))
	}}
	infoLogger := log.NewLogfmtLogger(infoWriter)
	infoLogger = level.NewFilter(infoLogger, cfg.LogLevel.Gokit, level.ErrNotAllowed(notAllowedError))

	warningWriter := &winLogWriter{writer: func(p []byte) {
		il.Warning(1, string(p))
	}}
	warningLogger := log.NewLogfmtLogger(warningWriter)
	warningLogger = level.NewFilter(warningLogger, cfg.LogLevel.Gokit, level.ErrNotAllowed(notAllowedError))

	errorWriter := &winLogWriter{writer: func(p []byte) {
		il.Error(1, string(p))
	}}
	errorLogger := log.NewLogfmtLogger(errorWriter)
	errorLogger = level.NewFilter(errorLogger, cfg.LogLevel.Gokit, level.ErrNotAllowed(notAllowedError))

	wl := &WinLoggerFmt{
		internalLog:   il,
		errorLogger:   errorLogger,
		infoLogger:    infoLogger,
		warningLogger: warningLogger,
	}
	return wl, nil

}

type WinLoggerFmt struct {
	internalLog *el.Log

	errorLogger   log.Logger
	infoLogger    log.Logger
	warningLogger log.Logger
}

func (w *WinLoggerFmt) Log(keyvals ...interface{}) error {
	lvl, err := getLevel(keyvals...)
	// If we don't know what happened move on
	if err != nil {
		w.infoLogger.Log(keyvals...)
		return err
	}
	// If the messages level matches one of these we try to write to the logger
	// the loggers are configured to reject the message if it isn't allowed.
	if lvl == level.DebugValue() {
		err = w.infoLogger.Log(keyvals...)
	} else if lvl == level.InfoValue() {
		err = w.infoLogger.Log(keyvals...)
	} else if lvl == level.ErrorValue() {
		err = w.errorLogger.Log(keyvals...)
	} else if lvl == level.WarnValue() {
		err = w.warningLogger.Log(keyvals...)
	}

	return err
}

// Looks through the key value pairs in the log for level and extract the value
func getLevel(keyvals ...interface{}) (level.Value, error) {
	if len(keyvals)%2 == 1 {
		keyvals = append(keyvals, nil)
	}
	for i := 0; i < len(keyvals); i += 2 {
		k, v := keyvals[i], keyvals[i+1]
		if k == "level" {
			if vo, ok := v.(level.Value); ok {
				return vo, nil
			}
			return nil, fmt.Errorf("unknown level")

		}
	}
	return nil, fmt.Errorf("no level found")
}

type winLogWriter struct {
	writer func(p []byte)
}

func (i *winLogWriter) Write(p []byte) (n int, err error) {
	i.writer(p)
	return len(p), nil
}
