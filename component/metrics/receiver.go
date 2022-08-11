package metrics

import (
	"github.com/panjf2000/ants/v2"

	"github.com/prometheus/prometheus/model/labels"
)

var pool, _ = ants.NewPool(1000)

// Receive is a func signature for something to receive metrics
type Receive func(timestamp int64, metrics []*FlowMetric)

// Receiver is used to pass an array of metrics to another receiver
type Receiver struct {
	rec Receive
}

func NewReceiver(recFunc Receive) *Receiver {
	return &Receiver{
		rec: recFunc,
	}
}

func (r *Receiver) Send(timestamp int64, metrics []*FlowMetric) {
	ants.Submit(func() {
		r.rec(timestamp, metrics)
	})
}

// RiverCapsule marks receivers as a capsule.
func (r Receiver) RiverCapsule() {}

// FlowMetric is a wrapper around a single metric without the timestamp.
type FlowMetric struct {
	globalRefID RefID
	labels      labels.Labels
	value       float64
}

// RefID wraps uint64 and used for a globally unique value
type RefID uint64

// NewFlowMetric instantiates a new flow metric
func NewFlowMetric(globalRefID RefID, lbls labels.Labels, value float64) *FlowMetric {
	// Always ensure we have a valid global ref id
	if globalRefID == 0 {
		globalRefID = GlobalRefMapping.getGlobalRefIDByLabels(lbls)
	}
	return &FlowMetric{
		globalRefID: globalRefID,
		labels:      lbls,
		value:       value,
	}
}

// GlobalRefID Retrieves the GlobalRefID
func (fw *FlowMetric) GlobalRefID() RefID { return fw.globalRefID }

// Value returns the value
func (fw *FlowMetric) Value() float64 { return fw.value }

// LabelsCopy returns a copy of the labels structure
func (fw *FlowMetric) LabelsCopy() labels.Labels {
	return fw.labels.Copy()
}
