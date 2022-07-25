package appendable

import (
	"context"

	"github.com/grafana/agent/component/metrics"
	"github.com/prometheus/prometheus/model/exemplar"
	"github.com/prometheus/prometheus/model/labels"
	"github.com/prometheus/prometheus/model/value"
	"github.com/prometheus/prometheus/storage"
)

// FlowMetric is a wrapper around a single sample without the timestamp.
type FlowMetric struct {
	Labels labels.Labels
	Value  float64
}

// FlowAppendable is a flow-specific implementation of an Appender.
type FlowAppendable []*metrics.Receiver

// NewFlowAppendable initializes the appendable.
func NewFlowAppendable(receivers ...*metrics.Receiver) FlowAppendable {
	return receivers
}

type flowAppender struct {
	buffer    map[int64][]*metrics.FlowMetric // Though mostly a map of 1 item, this allows it to work if more than one TS gets added
	receivers []*metrics.Receiver
}

// Appender implements the Prometheus Appendable interface.
func (app FlowAppendable) Appender(_ context.Context) storage.Appender {
	return &flowAppender{
		buffer:    make(map[int64][]*metrics.FlowMetric),
		receivers: app,
	}
}

func (app *flowAppender) Append(ref storage.SeriesRef, l labels.Labels, t int64, v float64) (storage.SeriesRef, error) {
	if len(app.receivers) == 0 {
		return 0, nil
	}
	_, found := app.buffer[t]
	if !found {
		set := make([]*metrics.FlowMetric, 0)
		app.buffer[t] = set
	}
	// If ref is 0 then lets grab a global id
	if ref == 0 {
		ref = storage.SeriesRef(metrics.GlobalRefMapping.GetOrAddGlobalRefID(l))
	}
	// If it is stale then we can remove it
	if value.IsStaleNaN(v) {
		metrics.GlobalRefMapping.AddStaleMarker(uint64(ref), l)
	} else {
		metrics.GlobalRefMapping.RemoveStaleMarker(uint64(ref))
	}
	app.buffer[t] = append(app.buffer[t], &metrics.FlowMetric{
		GlobalRefID: uint64(ref),
		Labels:      l,
		Value:       v,
	})
	return ref, nil
}

func (app *flowAppender) AppendExemplar(ref storage.SeriesRef, l labels.Labels, e exemplar.Exemplar) (storage.SeriesRef, error) {
	return 0, nil
}

func (app *flowAppender) Commit() error {
	for _, r := range app.receivers {
		for ts, metrics := range app.buffer {
			if r.Receive == nil {
				continue
			}
			r.Receive(ts, metrics)
		}
	}
	app.buffer = make(map[int64][]*metrics.FlowMetric)
	return nil
}

func (app *flowAppender) Rollback() error {
	app.buffer = make(map[int64][]*metrics.FlowMetric)
	return nil
}
