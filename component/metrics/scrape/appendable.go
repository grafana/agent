package scrape

import (
	"context"
	"sync"

	"github.com/grafana/agent/component/metrics"
	"github.com/prometheus/prometheus/model/exemplar"
	"github.com/prometheus/prometheus/model/labels"
	"github.com/prometheus/prometheus/model/value"
	"github.com/prometheus/prometheus/storage"
)

// FlowMetric is a wrapper around a single sample without the timestamp.
type FlowMetric struct {
	refID  uint64
	Labels labels.Labels
	Value  float64
}

type flowAppendable struct {
	mut       sync.Mutex
	buffer    map[int64][]*metrics.FlowMetric // Though mostly a map of 1 item, this allows it to work if more than one TS gets added
	receivers []*metrics.Receiver
}

func newFlowAppendable(receivers ...*metrics.Receiver) *flowAppendable {
	return &flowAppendable{
		receivers: receivers,
	}
}

func (app *flowAppendable) Appender(_ context.Context) storage.Appender {
	return app
}

func (app *flowAppendable) Append(ref storage.SeriesRef, l labels.Labels, t int64, v float64) (storage.SeriesRef, error) {
	app.mut.Lock()
	defer app.mut.Unlock()

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

func (app *flowAppendable) AppendExemplar(ref storage.SeriesRef, l labels.Labels, e exemplar.Exemplar) (storage.SeriesRef, error) {
	return 0, nil
}

func (app *flowAppendable) Commit() error {
	app.mut.Lock()
	defer app.mut.Unlock()
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

func (app *flowAppendable) Rollback() error {
	app.buffer = make(map[int64][]*metrics.FlowMetric)
	return nil
}
