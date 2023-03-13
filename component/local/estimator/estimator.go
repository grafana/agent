package estimator

import (
	"context"

	"github.com/grafana/agent/component"
	"github.com/grafana/agent/component/prometheus"
	"github.com/prometheus/prometheus/model/labels"
	"github.com/prometheus/prometheus/storage"
)

func init() {
	component.Register(component.Registration{
		Name:    "local.estimator",
		Args:    Arguments{},
		Exports: Exports{},

		Build: func(opts component.Options, args component.Arguments) (component.Component, error) {
			return New(opts, args.(Arguments))
		},
	})
}

type Component struct {
	opts            component.Options
	metricsReceiver *prometheus.Interceptor
	// This is maybe a hacky way to track unique combinations of seriesRef (map key), and label hash (map value).
	// Could consume less memory by combining those two values into a single unique hash, but I'm not sure it's worth the additional effort ¯\_(ツ)_/¯
	// Also, we may end up storing *actual* values for these, should we wish to report on them individually. I.E. There are x series with y label(s)
	activeSeries map[storage.SeriesRef]map[uint64]struct{}
}

type Arguments struct {
}

type Exports struct {
	metricsReceiver *prometheus.Interceptor `river:metrics_reciever,attr`
}

func New(o component.Options, args Arguments) (*Component, error) {
	c := &Component{
		opts:         o,
		activeSeries: make(map[storage.SeriesRef]map[uint64]struct{}),
	}

	interceptor := prometheus.NewInterceptor(
		nil,
		prometheus.WithAppendHook(func(globalRef storage.SeriesRef, l labels.Labels, t int64, v float64, next storage.Appender) (storage.SeriesRef, error) {

			// TODO: Is `prometheus.GlobalRefMapping` required here?
			_, ok := c.activeSeries[globalRef]
			if !ok {
				c.activeSeries[globalRef] = make(map[uint64]struct{})
			}

			_, ok = c.activeSeries[globalRef][l.Hash()]
			if !ok {
				c.activeSeries[globalRef][l.Hash()] = struct{}{}
			}
			return globalRef, nil
		}),
	)

	c.metricsReceiver = interceptor
	o.OnStateChange(Exports{metricsReceiver: c.metricsReceiver})

	return c, nil
}

func (c *Component) Run(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		}
	}
}

func (c *Component) Update(newConfig component.Arguments) error {
	// Reset the tracked active series
	c.activeSeries = make(map[storage.SeriesRef]map[uint64]struct{})
	return nil
}
