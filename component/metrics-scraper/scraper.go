package metricsscraper

import (
	"context"
	"sync"

	"github.com/davecgh/go-spew/spew"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/grafana/agent/component"
	metricsforwarder "github.com/grafana/agent/component/metrics-forwarder"
)

func init() {
	component.Register(component.Registration[Config]{
		Name: "metrics_scraper",
		BuildComponent: func(o component.Options, c Config) (component.Component[Config], error) {
			return NewComponent(o, c)
		},
	})
}

// Config represents the input state of the metrics_scraper component.
type Config struct {
	ScrapeInterval string                            `hcl:"scrape_interval,optional"`
	ScrapeTimeout  string                            `hcl:"scrape_timeout,optional"`
	Targets        []TargetGroup                     `hcl:"targets"`
	SendTo         *metricsforwarder.MetricsReceiver `hcl:"send_to"`
}

// TargetGroup is a set of targets that share a common set of labels.
type TargetGroup struct {
	Targets []LabelSet `hcl:"targets"`
	Labels  LabelSet   `hcl:"labels,optional"`
}

// LabelSet is a map of label names to values.
type LabelSet map[string]string

// Component is the metrics_scraper component.
type Component struct {
	log log.Logger

	mut sync.RWMutex
	cfg Config
}

// NewComponent creates a new metrics_scraper component.
func NewComponent(o component.Options, c Config) (*Component, error) {
	res := &Component{log: o.Logger, cfg: c}
	if err := res.Update(c); err != nil {
		return nil, err
	}
	return res, nil
}

var _ component.Component[Config] = (*Component)(nil)

// Run implements Component.
func (c *Component) Run(ctx context.Context, onStateChange func()) error {
	level.Info(c.log).Log("msg", "component starting")
	defer level.Info(c.log).Log("msg", "component shutting down")

	<-ctx.Done()
	return nil
}

// Update implements UpdatableComponent.
func (c *Component) Update(cfg Config) error {
	c.mut.Lock()
	defer c.mut.Unlock()
	spew.Dump(cfg)
	c.cfg = cfg
	return nil
}

// CurrentState implements Component.
func (c *Component) CurrentState() interface{} {
	return nil
}

// Config implements Component.
func (c *Component) Config() Config {
	c.mut.RLock()
	defer c.mut.RUnlock()
	return c.cfg
}
