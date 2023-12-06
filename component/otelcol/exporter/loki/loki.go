// Package loki provides an otelcol.exporter.loki component.
package loki

import (
	"context"
	"fmt"
	"time"

	"github.com/go-kit/log"
	"github.com/grafana/agent/component"
	"github.com/grafana/agent/component/common/loki"
	"github.com/grafana/agent/component/otelcol"
	"github.com/grafana/agent/component/otelcol/exporter/loki/internal/convert"
	"github.com/grafana/agent/component/otelcol/internal/lazyconsumer"
)

func init() {
	component.Register(component.Registration{
		Name:    "otelcol.exporter.loki",
		Args:    Arguments{},
		Exports: otelcol.ConsumerExports{},

		Build: func(o component.Options, a component.Arguments) (component.Component, error) {
			return New(o, a.(Arguments))
		},
	})
}

// Arguments configures the otelcol.exporter.loki component.
type Arguments struct {
	ForwardTo []loki.LogsReceiver `river:"forward_to,attr"`
}

// Component is the otelcol.exporter.loki component.
type Component struct {
	log  log.Logger
	opts component.Options

	converter               *convert.Converter
	logsReceiverStreamDebug *logsReceiverStreamDebug
}

var (
	_ component.Component   = (*Component)(nil)
	_ component.DebugStream = (*Component)(nil)
)

// New creates a new otelcol.exporter.loki component.
func New(o component.Options, c Arguments) (*Component, error) {
	converter := convert.New(o.Logger, o.Registerer, c.ForwardTo)

	res := &Component{
		log:  o.Logger,
		opts: o,

		converter: converter,
		logsReceiverStreamDebug: &logsReceiverStreamDebug{
			entries:             make(chan loki.Entry),
			debugStreamCallback: func(func() string) {},
		},
	}
	if err := res.Update(c); err != nil {
		return nil, err
	}

	// Construct a consumer based on our converter and export it. This will
	// remain the same throughout the component's lifetime, so we do this
	// during component construction.
	export := lazyconsumer.New(context.Background())
	export.SetConsumers(nil, nil, converter)
	o.OnStateChange(otelcol.ConsumerExports{Input: export})

	return res, nil
}

// Run implements Component.
func (c *Component) Run(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		case entry := <-c.logsReceiverStreamDebug.Chan():
			c.logsReceiverStreamDebug.debugStreamCallback(func() string {
				return fmt.Sprintf("ts=%s, labels=%s, entry=%s", entry.Timestamp.Format(time.RFC3339Nano), entry.Labels.String(), entry.Line)
			})
		}
	}
}

// Update implements Component.
func (c *Component) Update(newConfig component.Arguments) error {
	cfg := newConfig.(Arguments)
	c.converter.UpdateFanout(append(cfg.ForwardTo, c.logsReceiverStreamDebug))
	return nil
}

type logsReceiverStreamDebug struct {
	entries             chan loki.Entry
	debugStreamCallback func(computeDataFunc func() string)
}

func (l *logsReceiverStreamDebug) Chan() chan loki.Entry {
	return l.entries
}

func (c *Component) HookDebugStream(active bool, debugStreamCallback func(computeDataFunc func() string)) {
	c.logsReceiverStreamDebug.debugStreamCallback = debugStreamCallback
}
