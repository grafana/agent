package estimator

import (
	"context"
	"net/http"
	"path"

	prom_client "github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/model"

	"github.com/grafana/agent/component"
	"github.com/grafana/agent/component/discovery"
	"github.com/grafana/agent/component/prometheus"
	"github.com/prometheus/prometheus/model/labels"
	"github.com/prometheus/prometheus/storage"
)

func init() {
	component.Register(component.Registration{
		Name:    "prometheus.estimator",
		Args:    Arguments{},
		Exports: Exports{},

		Build: func(opts component.Options, args component.Arguments) (component.Component, error) {
			return New(opts, args.(Arguments))
		},
	})
}

type Component struct {
	opts     component.Options
	receiver *prometheus.Interceptor
	// This is maybe a hacky way to track unique combinations of seriesRef (map key), and label hash (map value).
	// Could consume less memory by combining those two values into a single unique hash, but I'm not sure it's worth the additional effort ¯\_(ツ)_/¯
	// Also, we may end up storing *actual* values for these, should we wish to report on them individually. I.E. There are x series with y label(s)
	activeSeries      map[storage.SeriesRef]map[uint64]struct{}
	activeSeriesGauge prom_client.Gauge
	reg               prom_client.Registry
}

type Arguments struct {
	Foo string `river:"foo,attr,optional"`
}

type Exports struct {
	Receiver storage.Appendable `river:"receiver,attr"`
	Targets  []discovery.Target `river:"targets,attr"`
}

func New(o component.Options, args Arguments) (*Component, error) {
	c := &Component{
		opts:         o,
		activeSeries: make(map[storage.SeriesRef]map[uint64]struct{}),
		activeSeriesGauge: prom_client.NewGauge(
			prom_client.GaugeOpts{
				Name: "estimator_active_series",
				Help: "The last count of active series being sent to the estimator",
			},
		),
		reg: *prom_client.NewRegistry(),
	}

	c.reg.MustRegister(c.activeSeriesGauge)

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

	c.receiver = interceptor
	o.OnStateChange(
		Exports{
			Receiver: c.receiver,
			Targets: []discovery.Target{
				{
					model.AddressLabel:     o.HTTPListenAddr,
					model.SchemeLabel:      "http",
					model.MetricsPathLabel: path.Join(o.HTTPPath, "metrics"),
					"instance":             o.ID,
					"job":                  "prometheus/estimator",
				},
			},
		})

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

func (c *Component) currentActiveSeriesCount() uint64 {
	series := 0
	for _, labels := range c.activeSeries {
		series = series + len(labels)
	}
	return uint64(series)
}

func (c *Component) DebugInfo() interface{} {
	return debugInfo{
		ActiveSeries: c.currentActiveSeriesCount(),
	}
}

type debugInfo struct {
	ActiveSeries uint64 `river:"active_series,attr"`
}

func (c *Component) Handler() http.Handler {
	return c
}

func (c *Component) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	series := c.currentActiveSeriesCount()
	c.activeSeriesGauge.Set(float64(series))
	if req.URL.Path == "/metrics" {
		promhttp.HandlerFor(&c.reg, promhttp.HandlerOpts{Registry: &c.reg}).ServeHTTP(w, req)
		return
	}
	w.WriteHeader(http.StatusNotFound)
}
