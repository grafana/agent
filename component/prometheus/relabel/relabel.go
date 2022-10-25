package relabel

import (
	"context"
	"sync"

	"github.com/prometheus/prometheus/model/exemplar"
	"github.com/prometheus/prometheus/model/metadata"

	"github.com/prometheus/prometheus/storage"

	"github.com/grafana/agent/component"
	flow_relabel "github.com/grafana/agent/component/common/relabel"
	"github.com/grafana/agent/component/prometheus"
	prometheus_client "github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/prometheus/model/labels"

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
	ForwardTo []storage.Appendable `river:"forward_to,attr"`

	// The relabelling rules to apply to each metric before it's forwarded.
	MetricRelabelConfigs []*flow_relabel.Config `river:"rule,block,optional"`
}

// Exports holds values which are exported by the prometheus.relabel component.
type Exports struct {
	Receiver storage.Appendable `river:"receiver,attr"`
}

// Component implements the prometheus.relabel component.
type Component struct {
	mut              sync.RWMutex
	opts             component.Options
	mrc              []*relabel.Config
	forwardto        []storage.Appendable
	metricsProcessed prometheus_client.Counter

	cacheMut sync.RWMutex
	cache    map[uint64]*labelAndID
}

var (
	_ component.Component = (*Component)(nil)
	_ storage.Appendable  = (*Component)(nil)
	_ storage.Appender    = (*appender)(nil)
)

func (c *Component) Appender(ctx context.Context) storage.Appender {
	app := &appender{
		children: make([]storage.Appender, 0),
		relabel:  c.Relabel,
	}
	for _, forward := range c.forwardto {
		app.children = append(app.children, forward.Appender(ctx))
	}
	return app
}

// New creates a new prometheus.relabel component.
func New(o component.Options, args Arguments) (*Component, error) {
	c := &Component{
		opts:  o,
		cache: make(map[uint64]*labelAndID),
	}
	c.forwardto = args.ForwardTo
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
	c.opts.OnStateChange(Exports{Receiver: c})

	return nil
}

func (c *Component) Relabel(val float64, lbls labels.Labels) labels.Labels {
	c.mut.RLock()
	defer c.mut.RUnlock()

	globalRef := prometheus.GlobalRefMapping.GetOrAddGlobalRefID(lbls)
	var relabelled labels.Labels
	newLbls, found := c.getFromCache(globalRef)
	if found {
		// If newLbls is nil but cache entry was found then we want to keep the value nil, if it's not we want to reuse the labels
		if newLbls != nil {
			relabelled = newLbls.labels
		}
	} else {
		processedLabels := relabel.Process(lbls, c.mrc...)
		c.addToCache(globalRef, processedLabels)
	}

	// If stale remove from the cache, the reason we don't exit early is so the stale value can propagate.
	// TODO: (@mattdurham) This caching can leak and likely needs a timed eviction at some point, but this is simple.
	// In the future the global ref cache may have some hooks to allow notification of when caches should be evicted.
	if value.IsStaleNaN(val) {
		c.deleteFromCache(globalRef)
	}
	if relabelled == nil {
		return nil
	}
	return newLbls.labels
}

func (c *Component) getFromCache(id uint64) (*labelAndID, bool) {
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

	c.cache = make(map[uint64]*labelAndID)
}

func (c *Component) addToCache(originalID uint64, lbls labels.Labels) {
	c.cacheMut.Lock()
	defer c.cacheMut.Unlock()

	if lbls == nil {
		c.cache[originalID] = nil
		return
	}
	newGlobal := prometheus.GlobalRefMapping.GetOrAddGlobalRefID(lbls)
	c.cache[originalID] = &labelAndID{
		labels: lbls,
		id:     newGlobal,
	}
}

// labelAndID stores both the globalrefid for the label and the id itself. We store the id so that it doesnt have
// to be recalculated again.
type labelAndID struct {
	labels labels.Labels
	id     uint64
}

type appender struct {
	children []storage.Appender
	relabel  func(val float64, lbls labels.Labels) labels.Labels
}

// Append statisfies the appender interface.
func (a *appender) Append(ref storage.SeriesRef, l labels.Labels, t int64, v float64) (storage.SeriesRef, error) {
	newLbls := a.relabel(v, l)
	if newLbls == nil {
		return ref, nil
	}
	for _, x := range a.children {
		_, _ = x.Append(ref, newLbls, t, v)
	}
	return ref, nil
}

// Commit satisfies the appender interface.
func (a *appender) Commit() error {
	for _, x := range a.children {
		_ = x.Commit()
	}
	return nil
}

// Rollback satisfies the appender interface.
func (a *appender) Rollback() error {
	for _, x := range a.children {
		_ = x.Rollback()
	}
	return nil
}

// AppendExemplar satisfies the appender interface.
func (a *appender) AppendExemplar(ref storage.SeriesRef, l labels.Labels, e exemplar.Exemplar) (storage.SeriesRef, error) {
	// TODO implement me
	panic("implement me")
}

// UpdateMetadata satisfies the appender interface.
func (a *appender) UpdateMetadata(ref storage.SeriesRef, l labels.Labels, m metadata.Metadata) (storage.SeriesRef, error) {
	// TODO implement me
	panic("implement me")
}
