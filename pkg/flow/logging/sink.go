package logging

import (
	"fmt"
	"io"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
)

// Sink is where a Controller logger will send log lines to.
type Sink struct {
	w         io.Writer // Raw writer to use
	updatable bool      // Whether the sink supports being updated.

	// parentComponentID is the ID of the parent component which generated the
	// sink. Empty if the sink is not associated with a component.
	parentComponentID string

	logger *lazyLogger // Constructed logger to use.
}

// WriterSink forwards logs to the provided [io.Writer]. WriterSinks support
// being updated.
func WriterSink(w io.Writer, o SinkOptions) (*Sink, error) {
	if w == nil {
		w = io.Discard
	}

	l, err := writerSinkLogger(w, o)
	if err != nil {
		return nil, err
	}

	return &Sink{
		w:         w,
		updatable: true,

		logger: &lazyLogger{inner: l},
	}, nil
}

// LoggerSink forwards logs to the provided Logger. The component ID from the
// provided Logger will be propagated to any new Loggers created using this
// Sink. LoggerSink does not support being updated.
func LoggerSink(c *Logger) *Sink {
	return &Sink{
		parentComponentID: fullID(c.parentComponentID, c.componentID),

		w:      io.Discard,
		logger: &lazyLogger{inner: c.orig},
	}
}

// Update reconfigures the options used for the Sink. Update will return an
// error if the options are invalid or if the Sink doesn't support being given
// SinkOptions.
func (s *Sink) Update(o SinkOptions) error {
	if !s.updatable {
		return fmt.Errorf("logging options cannot be updated in this context")
	}

	l, err := writerSinkLogger(s.w, o)
	if err != nil {
		return err
	}

	s.logger.UpdateInner(l)
	return nil
}

func writerSinkLogger(w io.Writer, o SinkOptions) (log.Logger, error) {
	var l log.Logger

	switch o.Format {
	case FormatLogfmt:
		l = log.NewLogfmtLogger(log.NewSyncWriter(w))
	case FormatJSON:
		l = log.NewJSONLogger(log.NewSyncWriter(w))
	default:
		return nil, fmt.Errorf("unrecognized log format %q", o.Format)
	}

	l = level.NewFilter(l, o.Level.Filter())

	if o.IncludeTimestamps {
		l = log.With(l, "ts", log.DefaultTimestampUTC)
	}
	return l, nil
}
