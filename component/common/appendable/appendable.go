package appendable

import (
	"context"
	"sync"

	"github.com/grafana/agent/component/metrics"
	"github.com/prometheus/prometheus/model/exemplar"
	"github.com/prometheus/prometheus/model/labels"
	"github.com/prometheus/prometheus/model/value"
	"github.com/prometheus/prometheus/storage"
)

// FlowAppendable is a flow-specific implementation of an Appender.
type FlowAppendable struct {
	mut       sync.RWMutex
	receivers []*metrics.Receiver
}

// NewFlowAppendable initializes the appendable.
func NewFlowAppendable(receivers ...*metrics.Receiver) *FlowAppendable {
	return &FlowAppendable{
		receivers: receivers,
	}
}

type flowAppender struct {
	buffer    map[int64][]*metrics.FlowMetric // Though mostly a map of 1 item, this allows it to work if more than one TS gets added
	receivers []*metrics.Receiver
}

// Appender implements the Prometheus Appendable interface.
func (app *FlowAppendable) Appender(_ context.Context) storage.Appender {
	app.mut.RLock()
	defer app.mut.RUnlock()

	return &flowAppender{
		buffer:    make(map[int64][]*metrics.FlowMetric),
		receivers: app.receivers,
	}
}

// SetReceivers defines the list of receivers for this appendable.
func (app *FlowAppendable) SetReceivers(receivers []*metrics.Receiver) {
	app.mut.Lock()
	app.receivers = receivers
	app.mut.Unlock()
}

// ListReceivers is a test method for exposing the Appender's receivers.
func (app *FlowAppendable) ListReceivers() []*metrics.Receiver {
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
		set := make([]*metrics.FlowMetric, 0)
		app.buffer[t] = set
	}
	// The incoming ref is a global refid. The refid might be 0, and in that case assigning it to a flow metric
	// will ensure that it is not 0.
	fm := metrics.NewFlowMetric(metrics.RefID(ref), l, v)
	// If it is stale then we can remove it.
	if value.IsStaleNaN(v) {
		metrics.GlobalRefMapping.AddStaleMarker(fm.GlobalRefID(), l)
	} else {
		metrics.GlobalRefMapping.RemoveStaleMarker(fm.GlobalRefID())
	}
	app.buffer[t] = append(app.buffer[t], fm)
	return ref, nil
}

func (app *flowAppender) AppendExemplar(
	ref storage.SeriesRef,
	l labels.Labels,
	e exemplar.Exemplar,
) (storage.SeriesRef, error) {
	// TODO actually implement this before flow production.
	return 0, nil
}

func (app *flowAppender) Commit() error {
	for _, r := range app.receivers {
		for ts, metrics := range app.buffer {
			if r == nil {
				continue
			}

			r.Send(ts, metrics)
		}
	}
	app.buffer = make(map[int64][]*metrics.FlowMetric)
	return nil
}

func (app *flowAppender) Rollback() error {
	app.buffer = make(map[int64][]*metrics.FlowMetric)
	return nil
}
