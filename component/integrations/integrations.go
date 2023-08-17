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

type Component struct{}

func (c *Component) Run(ctx context.Context) error {
	return nil
}

func (c *Component) Update(args component.Arguments) error {
	return nil
}

func New(o component.Options, args Arguments) (*Component, error) {
	c := &Component{}

	var integrations []internal.Integration

	for _, config := range args.Integrations {
		i, err := config.NewIntegration(o.Logger, args.Globals)
		if err != nil {
			return nil, err
		}
		integrations = append(integrations, i)
	}

	_, err := internal.NewSubsystem(o.Logger, args.Globals)
	if err != nil {
		return nil, err
	}

	return c, nil
}
