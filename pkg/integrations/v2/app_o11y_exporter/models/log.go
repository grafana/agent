package models

import (
	"fmt"
	"time"

	loki "github.com/prometheus/common/model"
)

// LogLevel is an alias for string
type LogLevel string

const (
	// LogLevelTrace is "trace"
	LogLevelTrace LogLevel = "trace"
	// LogLevelDebug is "debug"
	LogLevelDebug LogLevel = "debug"
	// LogLevelInfo is "info"
	LogLevelInfo LogLevel = "info"
	// LogLevelWarning is "warning"
	LogLevelWarning LogLevel = "warning"
	// LogLevelError is "error"
	LogLevelError LogLevel = "error"
)

// LogContext is a string to string map structure that
// represents the context of a log message
type LogContext map[string]string

// Log struct controls the data that come into a Log message
type Log struct {
	Message   string     `json:"message,omitempty"`
	LogLevel  LogLevel   `json:"level,omitempty"`
	Context   LogContext `json:"context,omitempty"`
	Timestamp time.Time  `json:"timestamp"`
}

// LabelSet creates the collection of labels required to export
// the Log into Loki
func (l Log) LabelSet() loki.LabelSet {
	labels := make(loki.LabelSet, len(l.Context)+1)

	for k, v := range l.Context {
		labels[loki.LabelName(fmt.Sprintf("context_%s", k))] = loki.LabelValue(v)
	}

	labels["level"] = loki.LabelValue(l.LogLevel)
	labels["kind"] = "logs"
	return labels
}
