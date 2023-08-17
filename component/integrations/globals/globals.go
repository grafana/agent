package globals

import (
	"context"

	"github.com/grafana/agent/component"
	internal "github.com/grafana/agent/pkg/integrations/v2"
)

func init() {
	component.Register(component.Registration{
		Name:    "integrations.globals",
		Args:    Arguments{},
		Exports: Exports{},

		Build: func(opts component.Options, args component.Arguments) (component.Component, error) {
			return New(opts, args.(Arguments))
		},
	})
}

type Arguments struct {
	agentIdentifier string `river:"agent_identifier,string"`
}

type Exports struct {
	Self internal.Globals `river:"self,attr"`
}

type Component struct{}

func (c *Component) Run(ctx context.Context) error {
	return nil
}

func (c *Component) Update(args component.Arguments) error {
	return nil
}

func New(o component.Options, args Arguments) (*Component, error) {
	c := &Component{}

	return c, nil
}
