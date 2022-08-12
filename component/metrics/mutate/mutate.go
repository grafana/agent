package mutate

import (
	"context"
	"sync"

	"github.com/grafana/agent/component"
	flow_relabel "github.com/grafana/agent/component/common/relabel"
	"github.com/grafana/agent/component/metrics"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/prometheus/model/relabel"
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
	mut       sync.RWMutex
	opts      component.Options
	mrc       []*relabel.Config
	forwardto []*metrics.Receiver

	receiver         *metrics.Receiver
	metricsProcessed prometheus.Counter
}

var (
	_ component.Component = (*Component)(nil)
)

// New creates a new metrics.mutate component.
func New(o component.Options, args Arguments) (*Component, error) {
	c := &Component{opts: o}

	c.receiver = &metrics.Receiver{Receive: c.Receive}
	c.metricsProcessed = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "agent_metrics_mutate_metrics_processed",
		Help: "Total number of metrics processed",
	})

	err := o.Registerer.Register(c.metricsProcessed)
	if err != nil {
		return nil, err
	}
	// Call to Update() to set the relabelling rules once at the start.
	if err = c.Update(args); err != nil {
		return nil, err
	}

	return c, nil
}

// Run implements component.Component.
func (c *Component) Run(ctx context.Context) error {
	<-ctx.Done()
	c.opts.Registerer.Unregister(c.metricsProcessed)
	return nil
}

// Update implements component.Component.
func (c *Component) Update(args component.Arguments) error {
	c.mut.Lock()
	defer c.mut.Unlock()

	newArgs := args.(Arguments)
	c.forwardto = newArgs.ForwardTo
	c.mrc = flow_relabel.ComponentToPromRelabelConfigs(newArgs.MetricRelabelConfigs)
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
	c.mut.RLock()
	defer c.mut.RLocker()

	relabelledMetrics := make([]*metrics.FlowMetric, 0, len(metricArr))
	for _, m := range metricArr {
		// Note relabel may return itself if no changes needed
		fm := m.Relabel(c.mrc...)
		if fm == nil {
			continue
		}
		relabelledMetrics = append(relabelledMetrics, fm)
	}
	if len(relabelledMetrics) == 0 {
		return
	}
	for _, forward := range c.forwardto {
		forward.Receive(ts, metricArr)
	}
}
