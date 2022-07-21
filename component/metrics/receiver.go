package metrics

import (
	"github.com/prometheus/prometheus/model/labels"
)

// Receiver is used to pass an array of metrics to another receiver
type Receiver struct {
	// metrics should be considered immutable
	Receive func(timestamp int64, metrics []*FlowMetric) `hcl:"receiver"`
}

// FlowMetric is a wrapper around a single metric without the timestamp
type FlowMetric struct {
	GlobalRefID uint64
	Labels      labels.Labels
	Value       float64
}
