package string

import (
	"context"

	"github.com/grafana/agent/component"
	"github.com/grafana/agent/component/module"
	"github.com/grafana/agent/service/cluster"
	"github.com/grafana/agent/service/http"
	otel_service "github.com/grafana/agent/service/otel"
	"github.com/grafana/river/rivertypes"
)

func init() {
	component.Register(component.Registration{
		Name:          "module.string",
		Args:          Arguments{},
		Exports:       module.Exports{},
		NeedsServices: []string{http.ServiceName, cluster.ServiceName, otel_service.ServiceName},

		Build: func(opts component.Options, args component.Arguments) (component.Component, error) {
			return New(opts, args.(Arguments))
		},
	})
}

// Arguments holds values which are used to configure the module.string
// component.
type Arguments struct {
	// Content to load for the module.
	Content rivertypes.OptionalSecret `river:"content,attr"`

	// Arguments to pass into the module.
	Arguments map[string]any `river:"arguments,block,optional"`
}

// Component implements the module.string component.
type Component struct {
	mod *module.ModuleComponent
}

var (
	_ component.Component       = (*Component)(nil)
	_ component.HealthComponent = (*Component)(nil)
)

// New creates a new module.string component.
func New(o component.Options, args Arguments) (*Component, error) {
	m, err := module.NewModuleComponent(o)
	if err != nil {
		return nil, err
	}
	c := &Component{
		mod: m,
	}

	if err := c.Update(args); err != nil {
		return nil, err
	}
	return c, nil
}

// Run implements component.Component.
func (c *Component) Run(ctx context.Context) error {
	c.mod.RunFlowController(ctx)
	return nil
}

// Update implements component.Component.
func (c *Component) Update(args component.Arguments) error {
	newArgs := args.(Arguments)

	return c.mod.LoadFlowSource(newArgs.Arguments, newArgs.Content.Value)
}

// CurrentHealth implements component.HealthComponent.
func (c *Component) CurrentHealth() component.Health {
	return c.mod.CurrentHealth()
}
