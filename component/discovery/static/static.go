package static

import (
	"context"
	"sync"

	"github.com/davecgh/go-spew/spew"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/grafana/agent/component"
	metricsscraper "github.com/grafana/agent/component/metrics-scraper"
	"github.com/prometheus/common/model"
)

func init() {
	component.Register(component.Registration[Config]{
		Name: "discovery.static",
		BuildComponent: func(o component.Options, c Config) (component.Component[Config], error) {
			return NewComponent(o, c)
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
	log     log.Logger
	updated chan struct{}

	mut   sync.RWMutex
	cfg   Config
	state State
}

// NewComponent creates a new discovery.static component.
func NewComponent(o component.Options, c Config) (*Component, error) {
	res := &Component{
		log:     o.Logger,
		updated: make(chan struct{}, 1),
	}
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

	for {
		select {
		case <-c.updated:
			onStateChange()
		case <-ctx.Done():
			return nil
		}
	}
}

// Update implements UpdatableComponent.
func (c *Component) Update(cfg Config) error {
	c.mut.Lock()
	defer c.mut.Unlock()
	spew.Dump(cfg)

	c.cfg = cfg

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

	// Enqueue an update so Run will invoke onStateChange
	select {
	case c.updated <- struct{}{}:
	default:
	}

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
