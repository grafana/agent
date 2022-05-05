package config

import (
	"encoding"
	"fmt"
)

// LogLevel is the logging level used by a Controller.
type LogLevel string

// Support log levels.
const (
	LogLevelDebug LogLevel = "debug"
	LogLevelInfo  LogLevel = "info"
	LogLevelWarn  LogLevel = "warn"
	LogLevelError LogLevel = "error"

	LogLevelDefault = LogLevelInfo
)

var (
	_ encoding.TextMarshaler   = LogLevel("")
	_ encoding.TextUnmarshaler = (*LogLevel)(nil)
)

// MarshalText implements encoding.TextMarshaler.
func (ll LogLevel) MarshalText() (text []byte, err error) {
	return []byte(ll), nil
}

// MarshalText implements encoding.TextUnmarshaler.
func (ll *LogLevel) UnmarshalText(text []byte) error {
	switch LogLevel(text) {
	case "":
		*ll = LogLevelDefault
	case LogLevelDebug, LogLevelInfo, LogLevelWarn, LogLevelError:
		*ll = LogLevel(text)
	default:
		return fmt.Errorf("unrecognized log level %q", string(text))
	}
	return nil
}

// LogFormat is the logging format used by a Controller.
type LogFormat string

// Support log levels.
const (
	LogFormatLogfmt LogFormat = "logfmt"
	LogFormatJSON   LogFormat = "json"

	LogFormatDefault = LogFormatLogfmt
)

var (
	_ encoding.TextMarshaler   = LogFormat("")
	_ encoding.TextUnmarshaler = (*LogFormat)(nil)
)

// MarshalText implements encoding.TextMarshaler.
func (ll LogFormat) MarshalText() (text []byte, err error) {
	return []byte(ll), nil
}

// MarshalText implements encoding.TextUnmarshaler.
func (ll *LogFormat) UnmarshalText(text []byte) error {
	switch LogFormat(text) {
	case "":
		*ll = LogFormatDefault
	case LogFormatLogfmt, LogFormatJSON:
		*ll = LogFormat(text)
	default:
		return fmt.Errorf("unrecognized log format %q", string(text))
	}
	return nil
}
