package util

import (
	"errors"
	"fmt"
	"strings"

	"github.com/go-kit/kit/log/level"

	"github.com/go-kit/kit/log"
	"github.com/weaveworks/common/server"
	el "golang.org/x/sys/windows/svc/eventlog"
)

func NewWinFmtLogger(cfg *server.Config) *Logger {
	return newWinFmtLogger(cfg, newWinLoggerFmt)
}

func newWinFmtLogger(cfg *server.Config, ctor func(*server.Config) (log.Logger, error)) *Logger {
	l := Logger{makeLogger: ctor}
	if err := l.ApplyConfig(cfg); err != nil {
		panic(err)
	}
	return &l
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
	if lvl == level.DebugValue() {
		err = w.infoLogger.Log(keyvals...)
		return err
	} else if lvl == level.ErrorValue() {
		err = w.errorLogger.Log(keyvals...)
		return err
	} else if lvl == level.WarnValue() {
		err = w.warningLogger.Log(keyvals...)
		return err
	} else if lvl == level.InfoValue() {
		err = w.infoLogger.Log(keyvals...)
		return err
	}
	return nil
}

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

type WinLogWriter struct {
	Writer func(p []byte)
}

func (i *WinLogWriter) Write(p []byte) (n int, err error) {
	i.Writer(p)
	return len(p), nil
}

func newWinLoggerFmt(cfg *server.Config) (log.Logger, error) {
	notAllowedError := errors.New("not_allowed")
	err := el.InstallAsEventCreate("Grafana Agent", el.Error|el.Info|el.Warning)
	if err != nil && !strings.Contains(err.Error(), "already exists") {
		fmt.Println(err)
	}
	il, err := el.Open("Grafana Agent")
	if err != nil {
		fmt.Println(err)
	}

	infoWriter := &WinLogWriter{Writer: func(p []byte) {
		il.Info(1, string(p))
	}}
	infoLogger := log.NewLogfmtLogger(infoWriter)
	infoLogger = level.NewFilter(infoLogger, cfg.LogLevel.Gokit, level.ErrNotAllowed(notAllowedError))

	warningWriter := &WinLogWriter{Writer: func(p []byte) {
		il.Warning(1, string(p))
	}}
	warningLogger := log.NewLogfmtLogger(warningWriter)
	warningLogger = level.NewFilter(warningLogger, cfg.LogLevel.Gokit, level.ErrNotAllowed(notAllowedError))

	errorWriter := &WinLogWriter{Writer: func(p []byte) {
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
	return log.With(wl, "caller", log.Caller(5)), nil

}
