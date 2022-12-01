package relabel

import (
	"context"
	"reflect"
	"sync"

	"github.com/go-kit/log/level"
	"github.com/grafana/agent/component"
	"github.com/grafana/agent/component/common/loki"
	flow_relabel "github.com/grafana/agent/component/common/relabel"
	lru "github.com/hashicorp/golang-lru"
	prometheus_client "github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/model/labels"
	"github.com/prometheus/prometheus/model/relabel"
)

func init() {
	component.Register(component.Registration{
		Name:    "loki.relabel",
		Args:    Arguments{},
		Exports: Exports{},
		Build: func(opts component.Options, args component.Arguments) (component.Component, error) {
			return New(opts, args.(Arguments))
		},
	})
}

// Arguments holds values which are used to configure the loki.relabel
// component.
type Arguments struct {
	// Where the relabelled metrics should be forwarded to.
	ForwardTo []loki.LogsReceiver `river:"forward_to,attr"`

	// The relabelling rules to apply to each log entry before it's forwarded.
	RelabelConfigs []*flow_relabel.Config `river:"rule,block,optional"`
}

// Exports holds values which are exported by the loki.relabel component.
type Exports struct {
	Receiver loki.LogsReceiver `river:"receiver,attr"`
}

// Component implements the loki.relabel component.
type Component struct {
	mut      sync.RWMutex
	opts     component.Options
	rcs      []*relabel.Config
	receiver loki.LogsReceiver
	fanout   []loki.LogsReceiver

	entriesProcessed prometheus_client.Counter
	entriesOutgoing  prometheus_client.Counter
	cacheHits        prometheus_client.Counter
	cacheMisses      prometheus_client.Counter
	cacheSize        prometheus_client.Gauge

	cache *lru.Cache
}

var (
	_ component.Component = (*Component)(nil)

	maxCacheSize = 10_000 // TODO(@tpaschalis) Do we make this configurable?
)

// New creates a new loki.relabel component.
func New(o component.Options, args Arguments) (*Component, error) {
	cache, err := lru.New(maxCacheSize)
	if err != nil {
		return nil, err
	}

	c := &Component{
		opts:  o,
		cache: cache,
	}
	c.entriesProcessed = prometheus_client.NewCounter(prometheus_client.CounterOpts{
		Name: "loki_relabel_entries_processed",
		Help: "Total number of log entries processed",
	})
	c.entriesOutgoing = prometheus_client.NewCounter(prometheus_client.CounterOpts{
		Name: "loki_relabel_entries_written",
		Help: "Total number of log entries forwarded",
	})
	c.cacheMisses = prometheus_client.NewCounter(prometheus_client.CounterOpts{
		Name: "loki_relabel_cache_misses",
		Help: "Total number of cache misses",
	})
	c.cacheHits = prometheus_client.NewCounter(prometheus_client.CounterOpts{
		Name: "loki_relabel_cache_hits",
		Help: "Total number of cache hits",
	})
	c.cacheSize = prometheus_client.NewGauge(prometheus_client.GaugeOpts{
		Name: "loki_relabel_cache_size",
		Help: "Total size of relabel cache",
	})

	for _, metric := range []prometheus_client.Collector{c.entriesProcessed, c.entriesOutgoing, c.cacheMisses, c.cacheHits, c.cacheSize} {
		err := o.Registerer.Register(metric)
		if err != nil {
			return nil, err
		}
	}

	// Create and immediately export the receiver which remains the same for
	// the component's lifetime.
	c.receiver = make(loki.LogsReceiver)
	o.OnStateChange(Exports{Receiver: c.receiver})

	// Call to Update() to set the relabelling rules once at the start.
	if err := c.Update(args); err != nil {
		return nil, err
	}

	return c, nil
}

// Run implements component.Component.
func (c *Component) Run(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		case entry := <-c.receiver:
			lbls := c.relabel(entry)
			if len(lbls) == 0 {
				level.Debug(c.opts.Logger).Log("msg", "dropping entry after relabeling", "labels", entry.Labels.String())
				continue
			}

			c.entriesOutgoing.Inc()
			entry.Labels = lbls
			for _, f := range c.fanout {
				select {
				case <-ctx.Done():
					return nil
				case f <- entry:
				}
			}
		}
	}
}

func (c *Component) relabel(e loki.Entry) model.LabelSet {
	c.entriesProcessed.Inc()
	hash := e.Labels.Fingerprint().String()
	found, ok := c.cache.Get(hash)
	if ok {
		c.cacheHits.Inc()
		return found.(model.LabelSet)
	}

	c.cacheMisses.Inc()

	// TODO(@tpaschalis) It's unfortunate how we have to cast back and forth
	// between model.LabelSet (map) and labels.Labels (slice). Promtail does
	// not have this issue as relabel config rules are only applied to targets.
	// Do we want to use labels.Labels in loki.Entry instead?
	var lbls labels.Labels
	for k, v := range e.Labels {
		lbls = append(lbls, labels.Label{
			Name:  string(k),
			Value: string(v),
		})
	}
	lbls = relabel.Process(lbls, c.rcs...)

	relabelled := make(model.LabelSet, len(lbls))
	for i := range lbls {
		relabelled[model.LabelName(lbls[i].Name)] = model.LabelValue(lbls[i].Value)
	}

	c.cache.Add(hash, relabelled)
	c.cacheSize.Set(float64(c.cache.Len()))

	return relabelled
}

// Update implements component.Component.
func (c *Component) Update(args component.Arguments) error {
	c.mut.Lock()
	defer c.mut.Unlock()

	newArgs := args.(Arguments)
	newRCS := flow_relabel.ComponentToPromRelabelConfigs(newArgs.RelabelConfigs)
	if relabelingChanged(c.rcs, newRCS) {
		level.Debug(c.opts.Logger).Log("msg", "received new relabel configs, purging cache")
		c.cache.Purge()
		c.cacheSize.Set(0)
	}
	c.rcs = newRCS
	c.cache.Purge()
	c.fanout = newArgs.ForwardTo

	return nil
}

// TODO(@tpaschalis) This is an attempt to not purge the cache if the
// relabeling rules are the same. One way to go about it would be to use the
// fingerprinting idea on Flow's relabel structs, but this seemed more
// straightforward. Are we okay with importing reflect here?
func relabelingChanged(prev, next []*relabel.Config) bool {
	if len(prev) != len(next) {
		return true
	}
	for i := range prev {
		if !reflect.DeepEqual(prev[i], next[i]) {
			return true
		}
	}
	return false
}
