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

	mut sync.RWMutex
	l   log.Logger
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
		l:    sink.l,
	}
}

// Log implements log.Logger.
func (c *Controller) Log(kvps ...interface{}) error {
	c.mut.RLock()
	defer c.mut.RUnlock()
	return c.l.Log(kvps...)
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

	c.mut.Lock()
	defer c.mut.Unlock()
	c.l = newLogger
	return nil
}
