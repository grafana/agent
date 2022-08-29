package relabel

import (
	"context"
	"sync"

	"github.com/grafana/agent/component"
	flow_relabel "github.com/grafana/agent/component/common/relabel"
	flow_prometheus "github.com/grafana/agent/component/prometheus"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/prometheus/model/relabel"
)

func init() {
	component.Register(component.Registration{
		Name:    "prometheus.relabel",
		Args:    Arguments{},
		Exports: Exports{},
		Build: func(opts component.Options, args component.Arguments) (component.Component, error) {
			return New(opts, args.(Arguments))
		},
	})
}

// Arguments holds values which are used to configure the prometheus.relabel
// component.
type Arguments struct {
	// Where the relabelled metrics should be forwarded to.
	ForwardTo []*flow_prometheus.Receiver `river:"forward_to,attr"`

	// The relabelling steps to apply to each metric before it's forwarded.
	MetricRelabelConfigs []*flow_relabel.Config `river:"metric_relabel_config,block,optional"`
}

// Exports holds values which are exported by the prometheus.relabel component.
type Exports struct {
	Receiver *flow_prometheus.Receiver `river:"receiver,attr"`
}

// Component implements the prometheus.relabel component.
type Component struct {
	mut              sync.RWMutex
	opts             component.Options
	mrc              []*relabel.Config
	forwardto        []*flow_prometheus.Receiver
	receiver         *flow_prometheus.Receiver
	metricsProcessed prometheus.Counter
}

var (
	_ component.Component = (*Component)(nil)
)

// New creates a new prometheus.relabel component.
func New(o component.Options, args Arguments) (*Component, error) {
	c := &Component{opts: o}
	c.receiver = &flow_prometheus.Receiver{Receive: c.Receive}
	c.metricsProcessed = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "agent_prometheus_relabel_metrics_processed",
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

	c.mrc = flow_relabel.ComponentToPromRelabelConfigs(newArgs.MetricRelabelConfigs)
	c.forwardto = newArgs.ForwardTo
	c.opts.OnStateChange(Exports{Receiver: c.receiver})

	return nil
}

// Receive implements the receiver.Receive func that allows an array of metrics
// to be passed around.
// TODO (@tpaschalis) The relabelling process will run _every_ time, for all
// metrics, resulting in some serious CPU overhead. We should be caching the
// relabeling results per refID and clearing entries for dropped or stale
// series. This is a blocker for releasing a production-grade version of the
// prometheus.relabel component.
func (c *Component) Receive(ts int64, metricArr []*flow_prometheus.FlowMetric) {
	c.mut.RLock()
	defer c.mut.RUnlock()

	relabelledMetrics := make([]*flow_prometheus.FlowMetric, 0)
	for _, m := range metricArr {
		// Relabel may return the original flowmetric if no changes applied, nil if everything was removed or an entirely new flowmetric.
		relabelledFm := m.Relabel(c.mrc...)
		if relabelledFm == nil {
			continue
		}
		relabelledMetrics = append(relabelledMetrics, relabelledFm)
	}
	if len(relabelledMetrics) == 0 {
		return
	}
	for _, forward := range c.forwardto {
		forward.Receive(ts, relabelledMetrics)
	}
}
