package gelf

import (
	"context"
	"sync"

	"github.com/grafana/agent/component"
	"github.com/grafana/agent/component/common/loki"
	flow_relabel "github.com/grafana/agent/component/common/relabel"
	"github.com/grafana/agent/component/loki/source/gelf/internal/target"
	"github.com/grafana/loki/clients/pkg/promtail/scrapeconfig"
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/model/relabel"
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
	target    *target.Target
	o         component.Options
	metrics   *target.Metrics
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
			if lokiEntry.Labels["job"] == "" {
				lokiEntry.Labels["job"] = model.LabelValue(c.o.ID)
			}
			for _, r := range c.receivers {
				r.Chan() <- lokiEntry
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

	var rcs []*relabel.Config
	if newArgs.RelabelRules != nil && len(newArgs.RelabelRules) > 0 {
		rcs = flow_relabel.ComponentToPromRelabelConfigs(newArgs.RelabelRules)
	}

	t, err := target.NewTarget(c.metrics, c.o.Logger, c.handler, rcs, convertConfig(newArgs))
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
	RelabelRules         flow_relabel.Rules  `river:"relabel_rules,attr,optional"`
	Receivers            []loki.LogsReceiver `river:"forward_to,attr"`
}

func defaultArgs() Arguments {
	return Arguments{
		ListenAddress:        "0.0.0.0:12201",
		UseIncomingTimestamp: false,
	}
}

// SetToDefault implements river.Defaulter.
func (r *Arguments) SetToDefault() {
	*r = defaultArgs()
}

func convertConfig(a Arguments) *scrapeconfig.GelfTargetConfig {
	return &scrapeconfig.GelfTargetConfig{
		ListenAddress:        a.ListenAddress,
		Labels:               nil,
		UseIncomingTimestamp: a.UseIncomingTimestamp,
	}
}

// New creates a new gelf component.
func New(o component.Options, args Arguments) (*Component, error) {
	metrics := target.NewMetrics(o.Registerer)
	c := &Component{
		o:       o,
		metrics: metrics,
		handler: &handler{c: make(chan loki.Entry)},
	}
	// Call to Update() to start readers and set receivers once at the start.
	if err := c.Update(args); err != nil {
		return nil, err
	}
	return c, nil
}

type handler struct {
	c chan loki.Entry
}

func (h *handler) Chan() chan<- loki.Entry {
	return h.c
}
func (handler) Stop() {
	// noop.
}
