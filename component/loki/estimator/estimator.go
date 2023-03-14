package estimator

import (
	"context"
	"net/http"
	"path"

	"github.com/grafana/agent/component"
	"github.com/grafana/agent/component/common/loki"
	"github.com/grafana/agent/component/discovery"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/model"
)

func init() {
	component.Register(component.Registration{
		Name:    "loki.estimator",
		Args:    Arguments{},
		Exports: Exports{},

		Build: func(opts component.Options, args component.Arguments) (component.Component, error) {
			return New(opts, args.(Arguments))
		},
	})
}

type Component struct {
	opts            component.Options
	receiver        loki.LogsReceiver
	logBytesCounter prometheus.Counter
	logBytes        uint64
	reg             prometheus.Registry
	entries         chan loki.Entry
}

type Arguments struct{}

type Exports struct {
	Reciever loki.LogsReceiver  `river:"receiver,attr"`
	Targets  []discovery.Target `river:"targets,attr"`
}

func New(o component.Options, args Arguments) (*Component, error) {
	c := &Component{
		opts: o,
		logBytesCounter: prometheus.NewCounter(
			prometheus.CounterOpts{
				Name: "estimator_log_bytes",
				Help: "Count of log bytes received. Excludes labels.",
			},
		),
		reg:     *prometheus.NewRegistry(),
		entries: make(chan loki.Entry),
	}
	c.reg.MustRegister(c.logBytesCounter)
	c.receiver = make(loki.LogsReceiver)

	o.OnStateChange(
		Exports{
			Reciever: c.receiver,
			Targets: []discovery.Target{
				{
					model.AddressLabel:     o.HTTPListenAddr,
					model.SchemeLabel:      "http",
					model.MetricsPathLabel: path.Join(o.HTTPPath, "metrics"),
					"instance":             o.ID,
					"job":                  "loki/estimator",
				},
			},
		},
	)
	return c, nil
}

func (c *Component) Chan() chan<- loki.Entry {
	return c.entries
}

func (c *Component) Run(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		case entry := <-c.receiver:
			select {
			case <-ctx.Done():
				return nil
			case c.Chan() <- entry:
				// no-op
			}
		}
	}
}

func (c *Component) Update(newConfig component.Arguments) error {
	return nil
}

func (c *Component) currentLogBytesCounter() uint64 {
	for {
		select {
		case e, ok := <-c.entries:
			if !ok {
				return c.logBytes
			}
			eSize := e.Size()
			c.logBytes = c.logBytes + uint64(eSize)
			c.logBytesCounter.Add(float64(eSize))
			return c.logBytes
		default:
			return c.logBytes
		}
	}
}

func (c *Component) DebugInfo() interface{} {
	return debugInfo{
		LogBytes: c.currentLogBytesCounter(),
	}
}

type debugInfo struct {
	LogBytes uint64 `river:"log_bytes,attr"`
}

func (c *Component) Handler() http.Handler {
	return c
}

func (c *Component) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	c.currentLogBytesCounter() // Don't need to fetch the return value, because this function also updates the counter.
	if req.URL.Path == "/metrics" {
		promhttp.HandlerFor(&c.reg, promhttp.HandlerOpts{Registry: &c.reg}).ServeHTTP(w, req)
		return
	}
	w.WriteHeader(http.StatusNotFound)
}
