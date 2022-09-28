package appendable

import (
	"context"
	"sync"

	"github.com/grafana/agent/component/prometheus"
	"github.com/prometheus/prometheus/model/exemplar"
	"github.com/prometheus/prometheus/model/labels"
	"github.com/prometheus/prometheus/model/metadata"
	"github.com/prometheus/prometheus/model/value"
	"github.com/prometheus/prometheus/storage"
)

// FlowMetric is a wrapper around a single sample without the timestamp.
type FlowMetric struct {
	Labels labels.Labels
	Value  float64
}

// FlowAppendable is a flow-specific implementation of an Appender.
type FlowAppendable struct {
	mut       sync.RWMutex
	receivers []*prometheus.Receiver
}

// NewFlowAppendable initializes the appendable.
func NewFlowAppendable(receivers ...*prometheus.Receiver) *FlowAppendable {
	return &FlowAppendable{
		receivers: receivers,
	}
}

type flowAppender struct {
	buffer    map[int64][]*prometheus.FlowMetric // Though mostly a map of 1 item, this allows it to work if more than one TS gets added
	receivers []*prometheus.Receiver
}

// Appender implements the Prometheus Appendable interface.
func (app *FlowAppendable) Appender(_ context.Context) storage.Appender {
	app.mut.RLock()
	defer app.mut.RUnlock()

	return &flowAppender{
		buffer:    make(map[int64][]*prometheus.FlowMetric),
		receivers: app.receivers,
	}
}

// SetReceivers defines the list of receivers for this appendable.
func (app *FlowAppendable) SetReceivers(receivers []*prometheus.Receiver) {
	app.mut.Lock()
	app.receivers = receivers
	app.mut.Unlock()
}

// ListReceivers is a test method for exposing the Appender's receivers.
func (app *FlowAppendable) ListReceivers() []*prometheus.Receiver {
	app.mut.RLock()
	defer app.mut.RUnlock()
	return app.receivers
}

func (app *flowAppender) Append(ref storage.SeriesRef, l labels.Labels, t int64, v float64) (storage.SeriesRef, error) {
	if len(app.receivers) == 0 {
		return 0, nil
	}
	_, found := app.buffer[t]
	if !found {
		set := make([]*prometheus.FlowMetric, 0)
		app.buffer[t] = set
	}
	// If ref is 0 then lets grab a global id
	if ref == 0 {
		ref = storage.SeriesRef(prometheus.GlobalRefMapping.GetOrAddGlobalRefID(l))
	}
	// If it is stale then we can remove it
	if value.IsStaleNaN(v) {
		prometheus.GlobalRefMapping.AddStaleMarker(uint64(ref), l)
	} else {
		prometheus.GlobalRefMapping.RemoveStaleMarker(uint64(ref))
	}
	app.buffer[t] = append(app.buffer[t], prometheus.NewFlowMetric(uint64(ref), l, v))
	return ref, nil
}

func (app *flowAppender) AppendExemplar(ref storage.SeriesRef, l labels.Labels, e exemplar.Exemplar) (storage.SeriesRef, error) {
	return 0, nil
}

func (app *flowAppender) UpdateMetadata(ref storage.SeriesRef, l labels.Labels, m metadata.Metadata) (storage.SeriesRef, error) {
	return 0, nil
}

func (app *flowAppender) Commit() error {
	for _, r := range app.receivers {
		for ts, metrics := range app.buffer {
			if r == nil || r.Receive == nil {
				continue
			}
			r.Receive(ts, metrics)
		}
	}
	app.buffer = make(map[int64][]*prometheus.FlowMetric)
	return nil
}

func (app *flowAppender) Rollback() error {
	app.buffer = make(map[int64][]*prometheus.FlowMetric)
	return nil
}
