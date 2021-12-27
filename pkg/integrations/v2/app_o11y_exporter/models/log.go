package models

import (
	"fmt"
	"time"

	loki "github.com/prometheus/common/model"
)

type LogLevel string

const (
	LogLevelTrace   LogLevel = "trace"
	LogLevelDebug   LogLevel = "debug"
	LogLevelInfo    LogLevel = "info"
	LogLevelWarning LogLevel = "warning"
	LogLevelError   LogLevel = "error"
)

type LogContext map[string]string

type Log struct {
	Message   string     `json:"message,omitempty"`
	LogLevel  LogLevel   `json:"level,omitempty"`
	Context   LogContext `json:"context,omitempty"`
	Timestamp time.Time  `json:"timestamp"`
}

func (l Log) LabelSet() loki.LabelSet {
	labels := make(loki.LabelSet, len(l.Context)+1)

	for k, v := range l.Context {
		labels[loki.LabelName(fmt.Sprintf("context_%s", k))] = loki.LabelValue(v)
	}

	labels["level"] = loki.LabelValue(l.LogLevel)
	return labels
}
