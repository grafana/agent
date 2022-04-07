package models

import (
	"time"

	"github.com/grafana/agent/pkg/integrations/v2/app_o11y_receiver/utils"
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
	Message   string       `json:"message,omitempty"`
	LogLevel  LogLevel     `json:"level,omitempty"`
	Context   LogContext   `json:"context,omitempty"`
	Timestamp time.Time    `json:"timestamp"`
	Trace     TraceContext `json:"trace,omitempty"`
}

// KeyVal representation of a Log object
func (l Log) KeyVal() *utils.KeyVal {
	kv := utils.NewKeyVal()
	utils.KeyValAdd(kv, "timestamp", l.Timestamp.String())
	utils.KeyValAdd(kv, "kind", "log")
	utils.KeyValAdd(kv, "message", l.Message)
	utils.KeyValAdd(kv, "level", string(l.LogLevel))
	utils.MergeKeyValWithPrefix(kv, utils.KeyValFromMap(l.Context), "context_")
	utils.MergeKeyVal(kv, l.Trace.KeyVal())
	return kv
}
