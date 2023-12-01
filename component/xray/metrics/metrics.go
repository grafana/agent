package metrics

import (
	"context"
	"fmt"
	"sync"

	"go.uber.org/atomic"

	"github.com/prometheus/prometheus/storage"

	"github.com/grafana/agent/component"
	"github.com/grafana/agent/component/prometheus"
	"github.com/grafana/agent/service/labelstore"
	"github.com/prometheus/prometheus/model/exemplar"
	"github.com/prometheus/prometheus/model/histogram"
	"github.com/prometheus/prometheus/model/labels"
	"github.com/prometheus/prometheus/model/metadata"
)

func init() {
	component.Register(component.Registration{
		Name:    "xray.metrics",
		Args:    Arguments{},
		Exports: Exports{},

		Build: func(o component.Options, c component.Arguments) (component.Component, error) {
			return New(o, c.(Arguments))
		},
	})
}

type Arguments struct {
}

type Exports struct {
	Receiver storage.Appendable `river:"receiver,attr"`
}

// Component implements the prometheus.relabel component.
type Component struct {
	mut      sync.RWMutex
	opts     component.Options
	receiver *prometheus.Interceptor

	fanout *prometheus.Fanout
	exited atomic.Bool
	ls     labelstore.LabelStore

	seriesByName map[string]*SeriesCounter
	seriesByJob  map[string]*SeriesCounter

	cacheMut sync.RWMutex
}

// count number of uniquie series
// should probably do rolling time buckets. For now just collect all
type SeriesCounter struct {
	series map[string]struct{}
}

func NewSeriesCounter() *SeriesCounter {
	return &SeriesCounter{
		series: map[string]struct{}{},
	}
}

func (s *SeriesCounter) Add(l labels.Labels) {
	s.series[l.String()] = struct{}{}
}

var (
	_ component.Component = (*Component)(nil)
)

func New(o component.Options, args Arguments) (*Component, error) {

	data, err := o.GetServiceData(labelstore.ServiceName)
	if err != nil {
		return nil, err
	}
	c := &Component{
		opts:         o,
		ls:           data.(labelstore.LabelStore),
		seriesByName: map[string]*SeriesCounter{},
		seriesByJob:  map[string]*SeriesCounter{},
	}

	c.fanout = prometheus.NewFanout(nil, o.ID, o.Registerer, c.ls)
	c.receiver = prometheus.NewInterceptor(
		c.fanout,
		c.ls,
		prometheus.WithAppendHook(func(_ storage.SeriesRef, l labels.Labels, t int64, v float64, next storage.Appender) (storage.SeriesRef, error) {
			if c.exited.Load() {
				return 0, fmt.Errorf("%s has exited", o.ID)
			}
			c.cacheMut.Lock()
			name := l.Get("__name__")
			job := l.Get("job")

			scount := c.seriesByName[name]
			if scount == nil {
				scount = NewSeriesCounter()
				c.seriesByName[name] = scount
			}
			scount.Add(l)

			scount = c.seriesByJob[job]
			if scount == nil {
				scount = NewSeriesCounter()
				c.seriesByJob[job] = scount
			}
			scount.Add(l)

			c.cacheMut.Unlock()
			return 0, nil
		}),
		// TODO: these do nothing
		prometheus.WithExemplarHook(func(_ storage.SeriesRef, l labels.Labels, e exemplar.Exemplar, next storage.Appender) (storage.SeriesRef, error) {
			if c.exited.Load() {
				return 0, fmt.Errorf("%s has exited", o.ID)
			}
			return 0, nil
		}),
		prometheus.WithMetadataHook(func(_ storage.SeriesRef, l labels.Labels, m metadata.Metadata, next storage.Appender) (storage.SeriesRef, error) {
			if c.exited.Load() {
				return 0, fmt.Errorf("%s has exited", o.ID)
			}
			return 0, nil
		}),
		prometheus.WithHistogramHook(func(_ storage.SeriesRef, l labels.Labels, t int64, h *histogram.Histogram, fh *histogram.FloatHistogram, next storage.Appender) (storage.SeriesRef, error) {
			if c.exited.Load() {
				return 0, fmt.Errorf("%s has exited", o.ID)
			}
			return 0, nil
		}),
	)

	// Immediately export the receiver which remains the same for the component
	// lifetime.
	o.OnStateChange(Exports{Receiver: c.receiver})

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

	//c.clearCache(100_000)

	return nil
}

// ScraperStatus reports the status of the scraper's jobs.
type debugInfo struct {
	Metrics map[string]int `river:"metrics,attr"`
	Jobs    map[string]int `river:"jobs,attr"`
}

// DebugInfo implements component.DebugComponent
func (c *Component) DebugInfo() interface{} {
	di := &debugInfo{
		Metrics: map[string]int{},
		Jobs:    map[string]int{},
	}
	c.cacheMut.RLock()
	for m, scount := range c.seriesByName {
		di.Metrics[m] = len(scount.series)
	}
	for m, scount := range c.seriesByJob {
		di.Jobs[m] = len(scount.series)
	}
	c.cacheMut.RUnlock()
	return di
}
