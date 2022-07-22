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
	ForwardTo []*metrics.Receiver `river:"forward_to,attr"`

	// The relabelling steps to apply to each metric before it's forwarded.
	MetricRelabelConfigs []*flow_relabel.Config `river:"metric_relabel_config,block,optional"`
}

// Exports holds values which are exported by the metrics.mutate component.
type Exports struct {
	Receiver *metrics.Receiver `river:"receiver,attr"`
}

// Component implements the metrics.mutate component.
type Component struct {
	opts component.Options
	mrc  []*relabel.Config

	appendable *fa.FlowAppendable
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

	c.mrc = flow_relabel.ComponentToPromRelabelConfigs(newArgs.MetricRelabelConfigs)
	c.appendable.SetReceivers(newArgs.ForwardTo)
	c.opts.OnStateChange(Exports{Receiver: c.receiver})

	return nil
}

// Receive implements the receiver.Receive func that allows an array of metrics
// to be passed around.
// TODO (@tpaschalis) The relabelling process will run _every_ time, for all
// metrics, resulting in some serious CPU overhead. We should be caching the
// relabeling results per refID and clearing entries for dropped or stale
// series. This is a blocker for releasing a production-grade  of the metrics.mutate
// component.
func (c *Component) Receive(ts int64, metricArr []*metrics.FlowMetric) {
	app := c.appendable.Appender(context.Background())
	for _, m := range metricArr {
		m.Labels = relabel.Process(m.Labels, c.mrc...)
		if m.Labels == nil {
			continue
		}
		_, err := app.Append(storage.SeriesRef(m.GlobalRefID), m.Labels, ts, m.Value)
		if err != nil {
			level.Error(c.opts.Logger).Log("msg", "failed to forward sample from metrics.mutate component", "err", err, "componentID", c.opts.ID)
		}
	}
	err := app.Commit()
	if err != nil {
		level.Error(c.opts.Logger).Log("msg", "failed to commit after relabelling metrics", "err", err)
	}
}
