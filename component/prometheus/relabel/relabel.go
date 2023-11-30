package relabel

import (
	"context"
	"fmt"
	"sync"

	"go.uber.org/atomic"

	"github.com/prometheus/prometheus/storage"

	"github.com/grafana/agent/component"
	flow_relabel "github.com/grafana/agent/component/common/relabel"
	"github.com/grafana/agent/component/prometheus"
	"github.com/grafana/agent/service/labelstore"
	prometheus_client "github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/prometheus/model/exemplar"
	"github.com/prometheus/prometheus/model/histogram"
	"github.com/prometheus/prometheus/model/labels"
	"github.com/prometheus/prometheus/model/metadata"

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
	ForwardTo []storage.Appendable `river:"forward_to,attr"`

	// The relabelling rules to apply to each metric before it's forwarded.
	MetricRelabelConfigs []*flow_relabel.Config `river:"rule,block,optional"`

	// Cache size to use for LRU cache.
	//CacheSize int `river:"cache_size,attr,optional"`
}

// SetToDefault implements river.Defaulter.
/*func (arg *Arguments) SetToDefault() {
	*arg = Arguments{
		CacheSize: 500_000,
	}
}*/

// Exports holds values which are exported by the prometheus.relabel component.
type Exports struct {
	Receiver storage.Appendable `river:"receiver,attr"`
	Rules    flow_relabel.Rules `river:"rules,attr"`
}

// Component implements the prometheus.relabel component.
type Component struct {
	mut              sync.RWMutex
	opts             component.Options
	mrc              []*relabel.Config
	receiver         *prometheus.Interceptor
	metricsProcessed prometheus_client.Counter
	metricsOutgoing  prometheus_client.Counter
	fanout           *prometheus.Fanout
	exited           atomic.Bool
	ls               labelstore.LabelStore
	cache            *flow_relabel.Cache
}

var (
	_ component.Component = (*Component)(nil)
)

// New creates a new prometheus.relabel component.
func New(o component.Options, args Arguments) (*Component, error) {
	data, err := o.GetServiceData(labelstore.ServiceName)
	if err != nil {
		return nil, err
	}
	cache, err := flow_relabel.NewCache(data.(labelstore.LabelStore), 100_000, "prometheus", o.Registerer)
	if err != nil {
		return nil, err
	}

	c := &Component{
		opts:  o,
		ls:    data.(labelstore.LabelStore),
		cache: cache,
	}
	c.metricsProcessed = prometheus_client.NewCounter(prometheus_client.CounterOpts{
		Name: "agent_prometheus_relabel_metrics_processed",
		Help: "Total number of metrics processed",
	})
	c.metricsOutgoing = prometheus_client.NewCounter(prometheus_client.CounterOpts{
		Name: "agent_prometheus_relabel_metrics_written",
		Help: "Total number of metrics written",
	})

	for _, metric := range []prometheus_client.Collector{c.metricsProcessed, c.metricsOutgoing} {
		err = o.Registerer.Register(metric)
		if err != nil {
			return nil, err
		}
	}

	c.fanout = prometheus.NewFanout(args.ForwardTo, o.ID, o.Registerer, c.ls)
	c.receiver = prometheus.NewInterceptor(
		c.fanout,
		c.ls,
		prometheus.WithAppendHook(func(global storage.SeriesRef, l labels.Labels, t int64, v float64, next storage.Appender) (storage.SeriesRef, error) {
			if c.exited.Load() {
				return 0, fmt.Errorf("%s has exited", o.ID)
			}

			newLbl := c.cache.Relabel(v, uint64(global), l, c.mrc)
			if newLbl.IsEmpty() {
				return 0, nil
			}
			c.metricsOutgoing.Inc()
			return next.Append(0, newLbl, t, v)
		}),
		prometheus.WithExemplarHook(func(global storage.SeriesRef, l labels.Labels, e exemplar.Exemplar, next storage.Appender) (storage.SeriesRef, error) {
			if c.exited.Load() {
				return 0, fmt.Errorf("%s has exited", o.ID)
			}

			newLbl := c.cache.Relabel(0, uint64(global), l, c.mrc)
			if newLbl.IsEmpty() {
				return 0, nil
			}
			return next.AppendExemplar(0, newLbl, e)
		}),
		prometheus.WithMetadataHook(func(global storage.SeriesRef, l labels.Labels, m metadata.Metadata, next storage.Appender) (storage.SeriesRef, error) {
			if c.exited.Load() {
				return 0, fmt.Errorf("%s has exited", o.ID)
			}

			newLbl := c.cache.Relabel(0, uint64(global), l, c.mrc)
			if newLbl.IsEmpty() {
				return 0, nil
			}
			return next.UpdateMetadata(0, newLbl, m)
		}),
		prometheus.WithHistogramHook(func(global storage.SeriesRef, l labels.Labels, t int64, h *histogram.Histogram, fh *histogram.FloatHistogram, next storage.Appender) (storage.SeriesRef, error) {
			if c.exited.Load() {
				return 0, fmt.Errorf("%s has exited", o.ID)
			}

			newLbl := c.cache.Relabel(0, uint64(global), l, c.mrc)
			if newLbl.IsEmpty() {
				return 0, nil
			}
			return next.AppendHistogram(0, newLbl, t, h, fh)
		}),
	)

	// Immediately export the receiver which remains the same for the component
	// lifetime.
	o.OnStateChange(Exports{Receiver: c.receiver, Rules: args.MetricRelabelConfigs})

	// Call to Update() to set the relabelling rules once at the start.
	if err = c.Update(args); err != nil {
		return nil, err
	}

	return c, nil
}

// Run implements component.Component.
func (c *Component) Run(ctx context.Context) error {
	defer c.exited.Store(true)

	<-ctx.Done()
	return nil
}

// Update implements component.Component.
func (c *Component) Update(args component.Arguments) error {
	c.mut.Lock()
	defer c.mut.Unlock()

	newArgs := args.(Arguments)
	c.cache.ClearCache(100_000)
	c.mrc = flow_relabel.ComponentToPromRelabelConfigs(newArgs.MetricRelabelConfigs)
	c.fanout.UpdateChildren(newArgs.ForwardTo)

	c.opts.OnStateChange(Exports{Receiver: c.receiver, Rules: newArgs.MetricRelabelConfigs})

	return nil
}
