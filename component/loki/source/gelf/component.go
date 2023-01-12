package gelf

import (
	"context"
	"sync"

	"github.com/prometheus/common/model"

	"github.com/grafana/agent/component"
	"github.com/grafana/agent/component/common/loki"
	"github.com/grafana/loki/clients/pkg/promtail/api"
	"github.com/grafana/loki/clients/pkg/promtail/scrapeconfig"
	"github.com/grafana/loki/clients/pkg/promtail/targets/gelf"
)

func init() {
	component.Register(component.Registration{
		Name: "loki.source.gelf",
		Args: Arguments{},

		Build: func(opts component.Options, args component.Arguments) (component.Component, error) {
			return New(opts, args.(Arguments))
		},
	})
}

var _ component.Component = (*Component)(nil)

// Component is a receiver for graylog formatted log files.
type Component struct {
	mut       sync.RWMutex
	target    *gelf.Target
	o         component.Options
	metrics   *gelf.Metrics
	handler   *handler
	receivers []loki.LogsReceiver
}

// Run starts the component.
func (c *Component) Run(ctx context.Context) error {
	defer func() {
		c.target.Stop()
	}()
	for {
		select {
		case <-ctx.Done():
			return nil
		case entry := <-c.handler.c:
			c.mut.RLock()
			lokiEntry := loki.Entry{
				Labels: entry.Labels,
				Entry:  entry.Entry,
			}
			lokiEntry.Labels["source"] = model.LabelValue(c.o.ID)
			for _, r := range c.receivers {
				r <- lokiEntry
			}
			c.mut.RUnlock()
		}
	}
}

// Update updates the fields of the component.
func (c *Component) Update(args component.Arguments) error {
	newArgs := args.(Arguments)

	c.mut.Lock()
	defer c.mut.Unlock()

	if c.target != nil {
		c.target.Stop()
	}
	c.receivers = newArgs.Receivers
	t, err := gelf.NewTarget(c.metrics, c.o.Logger, c.handler, nil, convertConfig(newArgs))
	if err != nil {
		return err
	}
	c.target = t
	return nil
}

// Arguments are the arguments for the component.
type Arguments struct {
	// ListenAddress only supports UDP.
	ListenAddress        string              `river:"listen_address,attr,optional"`
	UseIncomingTimestamp bool                `river:"use_incoming_timestamp,attr,optional"`
	Receivers            []loki.LogsReceiver `river:"forward_to,attr"`
}

func defaultArgs() Arguments {
	return Arguments{
		ListenAddress:        "0.0.0.0:12201",
		UseIncomingTimestamp: false,
	}
}

// UnmarshalRiver implements river.Unmarshaler.
func (r *Arguments) UnmarshalRiver(f func(v interface{}) error) error {
	*r = defaultArgs()

	type arguments Arguments
	if err := f((*arguments)(r)); err != nil {
		return err
	}

	return nil
}

func convertConfig(a Arguments) *scrapeconfig.GelfTargetConfig {
	return &scrapeconfig.GelfTargetConfig{
		ListenAddress:        a.ListenAddress,
		Labels:               nil,
		UseIncomingTimestamp: a.UseIncomingTimestamp,
	}
}

// New creates a new gelf component.
func New(o component.Options, args Arguments) (component.Component, error) {
	metrics := gelf.NewMetrics(o.Registerer)
	c := &Component{
		o:       o,
		metrics: metrics,
		handler: &handler{c: make(chan api.Entry)},
	}
	// Call to Update() to start readers and set receivers once at the start.
	if err := c.Update(args); err != nil {
		return nil, err
	}
	return c, nil
}

type handler struct {
	c chan api.Entry
}

func (h *handler) Chan() chan<- api.Entry {
	return h.c
}
func (handler) Stop() {
	// noop.
}
