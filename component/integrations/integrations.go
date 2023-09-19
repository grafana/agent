package integrations

import (
	"context"

	"github.com/grafana/agent/component"
	"github.com/grafana/agent/pkg/integrations/v2"
	internal "github.com/grafana/agent/pkg/integrations/v2"
)

func init() {
	component.Register(component.Registration{
		Name:    "integrations.v2",
		Args:    Arguments{},
		Exports: Exports{},

		Build: func(opts component.Options, args component.Arguments) (component.Component, error) {
			return New(opts, args.(Arguments))
		},
	})
}

type Arguments struct {
	Globals      integrations.Globals `river:"globals,attr"`
	Integrations []internal.Config    `river:"integrations,attr"`
}

type Exports struct{}

type Component struct {
	// SubSystem instance
	subsystem *internal.Subsystem
}

func (c *Component) Run(ctx context.Context) error {
	return nil
}

func (c *Component) Update(args component.Arguments) error {
	return nil
}

func New(o component.Options, args Arguments) (*Component, error) {

	subsystem, err := internal.NewSubsystem(o.Logger, args.Globals)
	if err != nil {
		return nil, err
	}

	return &Component{
		subsystem: subsystem,
	}, nil
}
