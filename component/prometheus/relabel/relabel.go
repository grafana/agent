package relabel

import (
	"context"
	"sync"

	"github.com/grafana/agent/component"
	flow_relabel "github.com/grafana/agent/component/common/relabel"
	"github.com/grafana/agent/component/prometheus"
	prometheus_client "github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/prometheus/model/relabel"
	"github.com/prometheus/prometheus/model/value"
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
	ForwardTo []*prometheus.Receiver `river:"forward_to,attr"`

	// The relabelling rules to apply to each metric before it's forwarded.
	MetricRelabelConfigs []*flow_relabel.Config `river:"rule,block,optional"`
}

// Exports holds values which are exported by the prometheus.relabel component.
type Exports struct {
	Receiver *prometheus.Receiver `river:"receiver,attr"`
}

// Component implements the prometheus.relabel component.
type Component struct {
	mut              sync.RWMutex
	opts             component.Options
	mrc              []*relabel.Config
	forwardto        []*prometheus.Receiver
	receiver         *prometheus.Receiver
	metricsProcessed prometheus_client.Counter

	cacheMut sync.RWMutex
	cache    map[uint64]*prometheus.FlowMetric
}

var _ component.Component = (*Component)(nil)

// New creates a new prometheus.relabel component.
func New(o component.Options, args Arguments) (*Component, error) {
	c := &Component{
		opts:  o,
		cache: make(map[uint64]*prometheus.FlowMetric),
	}
	c.receiver = &prometheus.Receiver{Receive: c.Receive}
	c.metricsProcessed = prometheus_client.NewCounter(prometheus_client.CounterOpts{
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
	c.clearCache()
	c.mrc = flow_relabel.ComponentToPromRelabelConfigs(newArgs.MetricRelabelConfigs)
	c.forwardto = newArgs.ForwardTo
	c.opts.OnStateChange(Exports{Receiver: c.receiver})

	return nil
}

// Receive implements the receiver.Receive func that allows an array of metrics
// to be passed around.
func (c *Component) Receive(ts int64, metricArr []*prometheus.FlowMetric) {
	c.mut.RLock()
	defer c.mut.RUnlock()

	relabelledMetrics := make([]*prometheus.FlowMetric, 0)
	for _, m := range metricArr {
		// Relabel may return the original flowmetric if no changes applied, nil if everything was removed or an entirely new flowmetric.
		var relabelledFm *prometheus.FlowMetric
		fm, found := c.getFromCache(m.GlobalRefID())
		if found {
			relabelledFm = fm
		} else {
			relabelledFm = m.Relabel(c.mrc...)
			c.addToCache(m.GlobalRefID(), relabelledFm)
		}

		// If stale remove from the cache, the reason we don't exit early is so the stale value can propagate.
		// TODO: (@mattdurham) This caching can leak and likely needs a timed eviction at some point, but this is simple.
		// In the future the global ref cache may have some hooks to allow notification of when caches should be evicted.
		if value.IsStaleNaN(m.Value()) {
			c.deleteFromCache(m.GlobalRefID())
		}
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

func (c *Component) getFromCache(id uint64) (*prometheus.FlowMetric, bool) {
	c.cacheMut.RLock()
	defer c.cacheMut.RUnlock()

	fm, found := c.cache[id]
	return fm, found
}

func (c *Component) deleteFromCache(id uint64) {
	c.cacheMut.Lock()
	defer c.cacheMut.Unlock()

	delete(c.cache, id)
}

func (c *Component) clearCache() {
	c.cacheMut.Lock()
	defer c.cacheMut.Unlock()

	c.cache = make(map[uint64]*prometheus.FlowMetric)
}

func (c *Component) addToCache(originalID uint64, fm *prometheus.FlowMetric) {
	c.cacheMut.Lock()
	defer c.cacheMut.Unlock()

	c.cache[originalID] = fm
}
