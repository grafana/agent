package metrics

import (
	"github.com/prometheus/prometheus/model/labels"
)

// Receiver is used to pass an array of metrics to another receiver
type Receiver struct {
	// metrics should be considered immutable
	Receive func(timestamp int64, metrics []*FlowMetric)
}

// RiverCapsule marks receivers as a capsule.
func (r Receiver) RiverCapsule() {}

// FlowMetric is a wrapper around a single metric without the timestamp, FlowMetric is *mostly* immutable
// the globalrefid can only be set if it is 0
type FlowMetric struct {
	globalRefID uint64
	labels      labels.Labels
	value       float64
}

// NewFlowMetric instantiates a new flow metric
func NewFlowMetric(globalRefID uint64, lbls labels.Labels, value float64) *FlowMetric {
	// Always ensure we have a valid global ref id
	if globalRefID == 0 {
		globalRefID = GlobalRefMapping.GetGlobalRefIDByLabels(lbls)
	}
	return &FlowMetric{
		globalRefID: globalRefID,
		labels:      lbls,
		value:       value,
	}
}

// GlobalRefID Retrieves the GlobalRefID
func (fw *FlowMetric) GlobalRefID() uint64 { return fw.globalRefID}

// Value returns the value
func (fw *FlowMetric) Value() float64 { return fw.value }

// LabelsCopy returns a copy of the labels structure
func (fw *FlowMetric) LabelsCopy() labels.Labels {
	return fw.labels.Copy()
}
