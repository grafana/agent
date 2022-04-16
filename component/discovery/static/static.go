package static

import (
	"context"
	"sync"

	"github.com/go-kit/log"
	"github.com/grafana/agent/component"
	metricsscraper "github.com/grafana/agent/component/metrics-scraper"
	"github.com/prometheus/common/model"
)

func init() {
	component.Register(component.Registration{
		Name:   "discovery.static",
		Config: Config{},
		BuildComponent: func(o component.Options, c component.Config) (component.Component, error) {
			return NewComponent(o, c.(Config))
		},
	})
}

// Config represents the input state of the discovery.static component.
type Config struct {
	Hosts  []string          `hcl:"hosts"`
	Labels map[string]string `hcl:"labels,optional"`
}

// State represents the output sate of the discovery.static component.
type State struct {
	Targets []metricsscraper.TargetGroup `hcl:"targets"`
}

// Component is the discovery.static component.
type Component struct {
	log  log.Logger
	opts component.Options

	mut   sync.RWMutex
	cfg   Config
	state State
}

// NewComponent creates a new discovery.static component.
func NewComponent(o component.Options, c Config) (*Component, error) {
	res := &Component{
		log:  o.Logger,
		opts: o,
	}
	if err := res.Update(c); err != nil {
		return nil, err
	}
	return res, nil
}

var _ component.Component = (*Component)(nil)

// Run implements Component.
func (c *Component) Run(ctx context.Context) error {
	<-ctx.Done()
	return nil
}

// Update implements UpdatableComponent.
func (c *Component) Update(newConfig component.Config) error {
	cfg := newConfig.(Config)

	c.mut.Lock()
	defer c.mut.Unlock()

	// Recalculate groups
	var group metricsscraper.TargetGroup
	for _, host := range cfg.Hosts {
		group.Targets = append(group.Targets, metricsscraper.LabelSet{
			model.AddressLabel: host,
		})
	}

	group.Labels = make(metricsscraper.LabelSet)
	for key, value := range cfg.Labels {
		group.Labels[key] = value
	}

	c.state.Targets = []metricsscraper.TargetGroup{group}

	c.opts.OnStateChange()
	c.cfg = cfg
	return nil
}

// CurrentState implements Component.
func (c *Component) CurrentState() interface{} {
	c.mut.RLock()
	defer c.mut.RUnlock()
	return c.state
}

// Config implements Component.
func (c *Component) Config() Config {
	c.mut.RLock()
	defer c.mut.RUnlock()
	return c.cfg
}
