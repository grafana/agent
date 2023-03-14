package logging

import (
	"io"

	"github.com/go-kit/log"
)

// Logger is a logger for Grafana Agent Flow components and controllers. It
// implements the [log.Logger] interface.
type Logger struct {
	sink *Sink

	parentComponentID string
	componentID       string

	orig log.Logger // Original logger before the component name was added.
	log  log.Logger // Logger with component name injected.
}

// New creates a new Logger from the provided logging Sink.
func New(sink *Sink, opts ...LoggerOption) *Logger {
	if sink == nil {
		sink, _ = WriterSink(io.Discard, DefaultSinkOptions)
	}

	l := &Logger{
		sink:              sink,
		parentComponentID: sink.parentComponentID,
		orig:              sink.logger,
	}
	for _, opt := range opts {
		opt(l)
	}

	// Build the final logger.
	l.log = wrapWithComponentID(sink.logger, sink.parentComponentID, l.componentID)

	return l
}

// LoggerOption is passed to New to customize the constructed Logger.
type LoggerOption func(*Logger)

// WithComponentID provides a component ID to the Logger.
func WithComponentID(id string) LoggerOption {
	return func(l *Logger) {
		l.componentID = id
	}
}

// Log implements log.Logger.
func (c *Logger) Log(kvps ...interface{}) error {
	return c.log.Log(kvps...)
}

func wrapWithComponentID(l log.Logger, parentID, componentID string) log.Logger {
	id := fullID(parentID, componentID)
	if id == "" {
		return l
	}
	return log.With(l, "component", id)
}

func fullID(parentID, componentID string) string {
	switch {
	case componentID == "":
		return parentID
	case parentID == "":
		return componentID
	default:
		return parentID + "/" + componentID
	}
}
