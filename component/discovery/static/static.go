package static

import (
	"context"

	"github.com/davecgh/go-spew/spew"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/grafana/agent/component"
	metricsscraper "github.com/grafana/agent/component/metrics-scraper"
)

func init() {
	component.Register(component.Registration[Config]{
		Name: "discovery.static",
		BuildComponent: func(l log.Logger, c Config) (component.Component[Config], error) {
			return NewComponent(l, c)
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
	log log.Logger
}

// NewComponent creates a new discovery.static component.
func NewComponent(l log.Logger, c Config) (*Component, error) {
	spew.Dump(c)
	return &Component{log: l}, nil
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
	spew.Dump(cfg)
	return nil
}

// CurrentState implements Component.
func (c *Component) CurrentState() interface{} {
	return State{}
}

// Config implements Component.
func (c *Component) Config() Config {
	return Config{}
}
