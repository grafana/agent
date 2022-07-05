package scrape

import (
	"context"
	"fmt"

	"github.com/prometheus/prometheus/model/exemplar"
	"github.com/prometheus/prometheus/model/labels"
	"github.com/prometheus/prometheus/storage"
)

// FlowMetric is a wrapper around a single sample without the timestamp.
type FlowMetric struct {
	refID  uint64
	Labels labels.Labels
	Value  float64
}

// Receiver is used to pass an array of metrics to another receiver
type Receiver struct {
	Receive func(timestamp int64, metrics []FlowMetric) `hcl:"receiver"`
}

type flowAppendable struct {
	Receivers []Receiver
}

func (app *flowAppendable) Appender(_ context.Context) storage.Appender {
	return app
}

func (app *flowAppendable) Append(ref storage.SeriesRef, l labels.Labels, t int64, v float64) (storage.SeriesRef, error) {
	fmt.Println("Called to Append with", ref, l, t, v)
	return 0, nil
}
func (app *flowAppendable) Commit() error   { return nil }
func (app *flowAppendable) Rollback() error { return nil }
func (app *flowAppendable) AppendExemplar(ref storage.SeriesRef, l labels.Labels, e exemplar.Exemplar) (storage.SeriesRef, error) {
	return 0, nil
}

func newFlowAppendable(receivers ...Receiver) *flowAppendable {
	return &flowAppendable{
		Receivers: receivers,
	}
}
