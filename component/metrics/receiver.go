package metrics

import (
	"github.com/prometheus/prometheus/model/labels"
	promrelabel "github.com/prometheus/prometheus/model/relabel"
)

// Receiver is used to pass an array of metrics to another receiver
type Receiver struct {
	// metrics should be considered immutable
	Receive func(timestamp int64, metrics []*FlowMetric)
}

// RiverCapsule marks receivers as a capsule.
func (r Receiver) RiverCapsule() {}

// FlowMetric is a wrapper around a single metric without the timestamp.
type FlowMetric struct {
	globalRefID uint64
	labels      labels.Labels
	value       float64
}

// NewFlowMetric instantiates a new flow metric
func NewFlowMetric(globalRefID uint64, lbls labels.Labels, value float64) *FlowMetric {
	// Always ensure we have a valid global ref id
	if globalRefID == 0 {
		globalRefID = GlobalRefMapping.GetOrAddGlobalRefID(lbls)
	}
	return &FlowMetric{
		globalRefID: globalRefID,
		labels:      lbls,
		value:       value,
	}
}

// GlobalRefID Retrieves the GlobalRefID
func (fw *FlowMetric) GlobalRefID() uint64 { return fw.globalRefID }

// Value returns the value
func (fw *FlowMetric) Value() float64 { return fw.value }

// LabelsCopy returns a copy of the labels structure
func (fw *FlowMetric) LabelsCopy() labels.Labels {
	return fw.labels.Copy()
}

// RawLabels returns the actual underlying labels that SHOULD be treated as immutable. Usage of this
// must be very careful to ensure that nothing that consume this mutates labels in anyway.
func (fw *FlowMetric) RawLabels() labels.Labels {
	return fw.labels
}

// Relabel applies normal prometheus relabel rules and returns a flow metric. NOTE this may return itself.
func (fw *FlowMetric) Relabel(cfgs ...*promrelabel.Config) *FlowMetric {
	retLbls := promrelabel.Process(fw.labels)
	if retLbls == nil {
		return nil
	}
	if retLbls.Hash() == fw.labels.Hash() {
		return fw
	}
	return NewFlowMetric(0, retLbls, fw.value)
}
