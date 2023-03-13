package logger

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

	l log.Logger // Constructed logger to use.
}

// WriterSink forwards logs to the provided [io.Writer].
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

		l: l,
	}, nil
}

// ControllerSink forwards logs to the provided Controller logger.
func ControllerSink(c *Controller) *Sink {
	return &Sink{
		w: io.Discard,
		l: c,
	}
}

// ComponentSink forwards logs to the provided Component logger. The component
// label from c is dropped.
func ComponentSink(c *Component) *Sink {
	return &Sink{
		w: io.Discard,

		// Send logs to the original logger the Component uses so the component ID
		// gets dropped.
		l: c.orig,
	}
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
