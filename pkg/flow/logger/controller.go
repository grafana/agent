package logger

import (
	"fmt"
	"io"
	"sync"

	"github.com/go-kit/log"
)

// Controller is a logger for Grafana Agent Flow controllers. It implements the
// [log.Logger] interface.
type Controller struct {
	sink *Sink

	parentComponentID string

	mut  sync.RWMutex
	orig log.Logger // Original logger before the component name was added.
	log  log.Logger // Logger with component name injected.
}

// NewControllerLogger creates a controller logger from the provided logging
// sink. If [WriterSink] is used, the resulting Controller may be updated by
// invoking [Controller.Update].
func NewControllerLogger(sink *Sink) *Controller {
	if sink == nil {
		sink, _ = WriterSink(io.Discard, DefaultSinkOptions)
	}

	return &Controller{
		sink: sink,

		orig: sink.l,
		log:  wrapWithComponentID(sink.l, sink.parentComponentID, ""),
	}
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

// Log implements log.Logger.
func (c *Controller) Log(kvps ...interface{}) error {
	c.mut.RLock()
	defer c.mut.RUnlock()
	return c.log.Log(kvps...)
}

// Update reconfigures the options used for the Logger. Update may only be
// called when a WriterSink was used to construct the Controller.
func (c *Controller) Update(o SinkOptions) error {
	if !c.sink.updatable {
		return fmt.Errorf("updating logging settings is not supported in this context")
	}

	newLogger, err := writerSinkLogger(c.sink.w, o)
	if err != nil {
		return err
	}

	l := newLogger
	if c.sink.parentComponentID != "" {
		l = log.With(l, "component", c.sink.parentComponentID)
	}

	c.mut.Lock()
	defer c.mut.Unlock()
	c.orig = newLogger
	c.log = l
	return nil
}
