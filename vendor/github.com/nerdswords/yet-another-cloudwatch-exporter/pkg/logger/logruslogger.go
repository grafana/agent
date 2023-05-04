package logger

import (
	"encoding"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"

	log "github.com/sirupsen/logrus"
)

type Logger interface {
	Info(message string, keyvals ...interface{})
	Debug(message string, keyvals ...interface{})
	Error(err error, message string, keyvals ...interface{})
	Warn(message string, keyvals ...interface{})
	With(keyvals ...interface{}) Logger
	IsDebugEnabled() bool
}

type logrusLogger struct {
	entry *log.Entry
}

func (l logrusLogger) Info(message string, keyvals ...interface{}) {
	l.entry.WithFields(toFields(keyvals...)).Info(message)
}

func (l logrusLogger) Debug(message string, keyvals ...interface{}) {
	l.entry.WithFields(toFields(keyvals...)).Debug(message)
}

func (l logrusLogger) Error(err error, message string, keyvals ...interface{}) {
	l.entry.WithFields(toFields(keyvals...)).WithError(err).Error(message)
}

func (l logrusLogger) Warn(message string, keyvals ...interface{}) {
	l.entry.WithFields(toFields(keyvals...)).Warn(message)
}

func (l logrusLogger) With(keyvals ...interface{}) Logger {
	return logrusLogger{l.entry.WithFields(toFields(keyvals...))}
}

func (l logrusLogger) IsDebugEnabled() bool {
	return l.entry.Logger.IsLevelEnabled(log.DebugLevel)
}

func NewLogrusLogger(logger *log.Logger) logrusLogger { //nolint:revive
	return logrusLogger{log.NewEntry(logger)}
}

var ErrMissingValue = errors.New("(MISSING)")

// This code is from https://github.com/go-kit/log/blob/main/json_logger.go#L23-L91 which safely handles odd keyvals
// lengths, and safely converting interface{} -> string
func toFields(keyvals ...interface{}) log.Fields {
	n := (len(keyvals) + 1) / 2 // +1 to handle case when len is odd
	m := make(map[string]interface{}, n)
	for i := 0; i < len(keyvals); i += 2 {
		k := keyvals[i]
		var v interface{} = ErrMissingValue
		if i+1 < len(keyvals) {
			v = keyvals[i+1]
		}
		merge(m, k, v)
	}

	return m
}

func merge(dst map[string]interface{}, k, v interface{}) {
	var key string
	switch x := k.(type) {
	case string:
		key = x
	case fmt.Stringer:
		key = safeString(x)
	default:
		key = fmt.Sprint(x)
	}

	// We want json.Marshaler and encoding.TextMarshaller to take priority over
	// err.Error() and v.String(). But json.Marshall (called later) does that by
	// default so we force a no-op if it's one of those 2 case.
	switch x := v.(type) {
	case json.Marshaler:
	case encoding.TextMarshaler:
	case error:
		v = safeError(x)
	case fmt.Stringer:
		v = safeString(x)
	}

	dst[key] = v
}

func safeString(str fmt.Stringer) (s string) {
	defer func() {
		if panicVal := recover(); panicVal != nil {
			if v := reflect.ValueOf(str); v.Kind() == reflect.Ptr && v.IsNil() {
				s = "NULL"
			} else {
				panic(panicVal)
			}
		}
	}()
	s = str.String()
	return
}

func safeError(err error) (s interface{}) {
	defer func() {
		if panicVal := recover(); panicVal != nil {
			if v := reflect.ValueOf(err); v.Kind() == reflect.Ptr && v.IsNil() {
				s = nil
			} else {
				panic(panicVal)
			}
		}
	}()
	s = err.Error()
	return
}
