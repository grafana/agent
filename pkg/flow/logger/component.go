package logger

import (
	"io"

	"github.com/go-kit/log"
)

// Component is a logger passed to Grafana Agent Flow components. It implements
// the [log.Logger] interface.
type Component struct {
	componentID string

	orig log.Logger // Original logger before the component name was added.
	log  log.Logger // Logger with component name injected.
}

// NewComponentLogger creates a component logger from the provided logging
// sink.
func NewComponentLogger(sink *Sink, componentID string) *Component {
	if sink == nil {
		sink, _ = WriterSink(io.Discard, DefaultSinkOptions)
	}

	return &Component{
		componentID: fullID(sink.parentComponentID, componentID),

		orig: sink.l,
		log:  wrapWithComponentID(sink.l, sink.parentComponentID, componentID),
	}
}

// Log implements log.Logger.
func (c *Component) Log(kvps ...interface{}) error {
	return c.log.Log(kvps...)
}
