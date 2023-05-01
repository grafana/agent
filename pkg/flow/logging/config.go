package logging

import (
	"encoding"
	"fmt"

	"github.com/go-kit/log/level"
	"github.com/grafana/agent/pkg/river"
)

// Options is a set of options used to construct and configure a Logger.
type Options struct {
	Level  Level  `river:"level,attr,optional"`
	Format Format `river:"format,attr,optional"`

	// TODO: log sink parameter (e.g., to use the Windows Event logger)
}

// DefaultOptions holds defaults for creating a Logger.
var DefaultOptions = Options{
	Level:  LevelDefault,
	Format: FormatDefault,
}

var _ river.Unmarshaler = (*Options)(nil)

// UnmarshalRiver implements river.Unmarshaler.
func (o *Options) UnmarshalRiver(f func(interface{}) error) error {
	*o = DefaultOptions

	type options Options
	return f((*options)(o))
}

// Level represents how verbose logging should be.
type Level string

// Supported log levels
const (
	LevelDebug Level = "debug"
	LevelInfo  Level = "info"
	LevelWarn  Level = "warn"
	LevelError Level = "error"

	LevelDefault = LevelInfo
)

var (
	_ encoding.TextMarshaler   = LevelDefault
	_ encoding.TextUnmarshaler = (*Level)(nil)
)

// MarshalText implements encoding.TextMarshaler.
func (ll Level) MarshalText() (text []byte, err error) {
	return []byte(ll), nil
}

// UnmarshalText implements encoding.TextUnmarshaler.
func (ll *Level) UnmarshalText(text []byte) error {
	switch Level(text) {
	case "":
		*ll = LevelDefault
	case LevelDebug, LevelInfo, LevelWarn, LevelError:
		*ll = Level(text)
	default:
		return fmt.Errorf("unrecognized log level %q", string(text))
	}
	return nil
}

// Filter returns a go-kit logging filter from the level.
func (ll Level) Filter() level.Option {
	switch ll {
	case LevelDebug:
		return level.AllowDebug()
	case LevelInfo:
		return level.AllowInfo()
	case LevelWarn:
		return level.AllowWarn()
	case LevelError:
		return level.AllowError()
	default:
		return level.AllowAll()
	}
}

// Format represents a text format to use when writing logs.
type Format string

// Supported log formats.
const (
	FormatLogfmt Format = "logfmt"
	FormatJSON   Format = "json"

	FormatDefault = FormatLogfmt
)

var (
	_ encoding.TextMarshaler   = FormatDefault
	_ encoding.TextUnmarshaler = (*Format)(nil)
)

// MarshalText implements encoding.TextMarshaler.
func (ll Format) MarshalText() (text []byte, err error) {
	return []byte(ll), nil
}

// UnmarshalText implements encoding.TextUnmarshaler.
func (ll *Format) UnmarshalText(text []byte) error {
	switch Format(text) {
	case "":
		*ll = FormatDefault
	case FormatLogfmt, FormatJSON:
		*ll = Format(text)
	default:
		return fmt.Errorf("unrecognized log format %q", string(text))
	}
	return nil
}
