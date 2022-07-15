package mutate

import (
	"context"

	"github.com/go-kit/log/level"
	"github.com/grafana/agent/component"
	fa "github.com/grafana/agent/component/common/appendable"
	flow_relabel "github.com/grafana/agent/component/common/relabel"
	"github.com/grafana/agent/component/metrics"
	"github.com/prometheus/prometheus/model/relabel"
	"github.com/prometheus/prometheus/storage"
)

func init() {
	component.Register(component.Registration{
		Name:    "metrics.mutate",
		Args:    Arguments{},
		Exports: Exports{},

		Build: func(opts component.Options, args component.Arguments) (component.Component, error) {
			return New(opts, args.(Arguments))
		},
	})
}

// Arguments holds values which are used to configure the metrics.mutate
// component.
type Arguments struct {
	// Where the relabelled metrics should be forwarded to.
	ForwardTo []*metrics.Receiver `hcl:"forward_to"`

	// The relabelling steps to apply to each metric before it's forwarded.
	MetricRelabelConfigs []*flow_relabel.RelabelConfig `hcl:"metric_relabel_config,block"`
}

// Exports holds values which are exported by the metrics.mutate component.
type Exports struct {
	Receiver *metrics.Receiver `hcl:"receiver"`
}

// Component implements the metrics.mutate component.
type Component struct {
	opts component.Options
	mrc  []*relabel.Config

	appendable *appendable.FlowAppendable
	receiver   *metrics.Receiver
}

var (
	_ component.Component = (*Component)(nil)
)

// New creates a new metrics.mutate component.
func New(o component.Options, args Arguments) (*Component, error) {
	c := &Component{opts: o}
	c.appendable = fa.NewFlowAppendable(args.ForwardTo...)
	c.receiver = &metrics.Receiver{Receive: c.Receive}

	// Call to Update() to set the relabelling rules once at the start.
	if err := c.Update(args); err != nil {
		return nil, err
	}

	return c, nil
}

// Run implements component.Component.
func (c *Component) Run(ctx context.Context) error {
	<-ctx.Done()
	return nil
}

// Update implements component.Component.
func (c *Component) Update(args component.Arguments) error {
	newArgs := args.(Arguments)

	c.mrc = flow_relabel.HCLToPromRelabelConfigs(newArgs.MetricRelabelConfigs)
	c.appendable.Receivers = newArgs.ForwardTo
	c.opts.OnStateChange(Exports{Receiver: c.receiver})

	return nil
}

// Receive implements the receiver.Receive func that allows an array of metrics
// to be passed around.
func (c *Component) Receive(ts int64, metricArr []*metrics.FlowMetric) {
	app := c.appendable.Appender(context.Background())
	for _, m := range metricArr {
		m.Labels = relabel.Process(m.Labels, c.mrc...)
		if m.Labels == nil {
			continue
		}
		app.Append(storage.SeriesRef(m.GlobalRefID), m.Labels, ts, m.Value)
	}
	err := app.Commit()
	if err != nil {
		level.Error(c.opts.Logger).Log("msg", "failed to commit after relabelling metrics", "err", err)
	}
}
