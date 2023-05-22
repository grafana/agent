package memory

import (
	"context"
	"sync"

	"github.com/grafana/agent/component"
	"github.com/prometheus/prometheus/tsdb/record"
	"github.com/prometheus/prometheus/tsdb/wlog"
)

func init() {
	component.Register(component.Registration{
		Name:      "prometheus.wal.memory",
		Singleton: false,
		Args:      Arguments{},
		Exports:   Exports{},
		Build: func(opts component.Options, args component.Arguments) (component.Component, error) {
			return NewComponent(opts, args.(Arguments))
		},
	})
}

type Component struct {
	mut  sync.Mutex
	args Arguments
}

func (c *Component) SetWriteTo(write wlog.WriteTo) {
	panic("not implemented") // TODO: Implement
}
func (c *Component) Start() {
	panic("not implemented") // TODO: Implement
}
func (c *Component) Stop() {
	panic("not implemented") // TODO: Implement
}

// Next two methods are intended for garbage-collection: first we call
// UpdateSeriesSegment on all current series
func (c *Component) UpdateSeriesSegment(_ []record.RefSeries, _ int) {
	panic("not implemented") // TODO: Implement
}

// Then SeriesReset is called to allow the deletion
// of all series created in a segment lower than the argument.
func (c *Component) SeriesReset(_ int) {
	panic("not implemented") // TODO: Implement
}

func NewComponent(o component.Options, c Arguments) (*Component, error) {
	return &Component{
		args: c,
	}, nil
}

func (c *Component) Run(ctx context.Context) error {
	return nil
}

func (c *Component) Update(args component.Arguments) error {
	c.args = args.(Arguments)
	return nil
}

// Append and AppendExemplar should block until the samples are fully accepted,
// whether enqueued in memory or successfully written to it's final destination.
// Once returned, the WAL Watcher will not attempt to pass that data again.
func (c *Component) Append(_ []record.RefSample) (_ bool) {
	panic("not implemented") // TODO: Implement
}
func (c *Component) AppendExemplars(_ []record.RefExemplar) (_ bool) {
	panic("not implemented") // TODO: Implement
}
func (c *Component) AppendHistograms(_ []record.RefHistogramSample) (_ bool) {
	panic("not implemented") // TODO: Implement
}
func (c *Component) AppendFloatHistograms(_ []record.RefFloatHistogramSample) (_ bool) {
	panic("not implemented") // TODO: Implement
}
func (c *Component) StoreSeries(_ []record.RefSeries, _ int) {
	panic("not implemented") // TODO: Implement
}

type Arguments struct{}

type Exports struct{}
