package models

import (
	"fmt"
	"time"

	"github.com/go-sourcemap/sourcemap"
	loki "github.com/prometheus/common/model"
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
	Type       string      `json:"type,omitempty"`
	Value      string      `json:"value,omitempty"`
	Stacktrace *Stacktrace `json:"stacktrace,omitempty"`
	Timestamp  time.Time   `json:"timestamp"`
}

// MapFrames converts a stacktrace with obfuscated frame sources, into
// a real path one using source maps
func (s Stacktrace) MapFrames(scm *sourcemap.Consumer) Stacktrace {
	var frames []Frame
	for _, frame := range s.Frames {
		file, fn, line, col, ok := scm.Source(frame.Lineno, frame.Colno)

		if ok {
			newFrame := Frame{
				Function: fn,
				Module:   frame.Module,
				Filename: file,
				Lineno:   line,
				Colno:    col,
			}

			frames = append(frames, newFrame)
		} else {
			frames = append(frames, frame)
		}

	}

	return Stacktrace{Frames: frames}
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

// LabelSet creates the labels required to export the exception
// int Loki
func (e Exception) LabelSet() loki.LabelSet {
	labels := make(loki.LabelSet, 3)

	labels["kind"] = "exception"
	labels["type"] = loki.LabelValue(e.Type)
	labels["value"] = loki.LabelValue(e.Value)

	return labels
}
