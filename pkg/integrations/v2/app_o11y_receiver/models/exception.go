package models

import (
	"fmt"
	"time"

	"github.com/grafana/agent/pkg/integrations/v2/app_o11y_receiver/utils"
)

// Frame struct represents a single stacktrace frame
type Frame struct {
	Function string `json:"function,omitempty"`
	Module   string `json:"module,omitempty"`
	Filename string `json:"filename,omitempty"`
	Lineno   int    `json:"lineno,omitempty"`
	Colno    int    `json:"colno,omitempty"`
}

// String function converts a Frame into a human readable string
func (frame Frame) String() string {
	module := ""
	if len(frame.Module) > 0 {
		module = frame.Module + "|"
	}
	return fmt.Sprintf("\n  at %s (%s%s:%v:%v)", frame.Function, module, frame.Filename, frame.Lineno, frame.Colno)
}

// Stacktrace is a collection of Frames
type Stacktrace struct {
	Frames []Frame `json:"frames,omitempty"`
}

// Exception struct controls all the data regarding an exception
type Exception struct {
	Type       string       `json:"type,omitempty"`
	Value      string       `json:"value,omitempty"`
	Stacktrace *Stacktrace  `json:"stacktrace,omitempty"`
	Timestamp  time.Time    `json:"timestamp"`
	Trace      TraceContext `json:"trace,omitempty"`
}

// Message string is concatenating of the Exception.Type and Exception.Value
func (e Exception) Message() string {
	return fmt.Sprintf("%s: %s", e.Type, e.Value)
}

// String is the string representation of an Exception
func (e Exception) String() string {
	var stacktrace = e.Message()
	if e.Stacktrace != nil {
		for _, frame := range e.Stacktrace.Frames {
			stacktrace += frame.String()
		}
	}
	return stacktrace
}

// KeyVal representation of the exception object
func (e Exception) KeyVal() *utils.KeyVal {
	kv := utils.NewKeyVal()
	utils.KeyValAdd(kv, "timestamp", e.Timestamp.String())
	utils.KeyValAdd(kv, "kind", "exception")
	utils.KeyValAdd(kv, "type", e.Type)
	utils.KeyValAdd(kv, "value", e.Value)
	utils.KeyValAdd(kv, "stacktrace", e.String())
	utils.MergeKeyVal(kv, e.Trace.KeyVal())
	return kv
}
