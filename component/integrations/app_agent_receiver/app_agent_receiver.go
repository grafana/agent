package app_agent_receiver

import (
	"context"

	"github.com/grafana/agent/component"
	internal "github.com/grafana/agent/pkg/integrations/v2/app_agent_receiver"
)

func init() {
	component.Register(component.Registration{
		Name:    "integrations.v2.app_agent_receiver",
		Args:    Arguments{},
		Exports: Exports{},

		Build: func(opts component.Options, args component.Arguments) (component.Component, error) {
			return New(opts, args.(Arguments))
		},
	})
}

type Exports struct {
	Config internal.Config `river:"self,attr"`
}

type Component struct{}

func New(o component.Options, args Arguments) (*Component, error) {
	c := &Component{}

	if err := c.Update(args); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *Component) Run(ctx context.Context) error {
	// TODO(rfratto):
	//
	// * Start HTTP server for collecting telemetry from Faro clients.

	return nil
}

func (c *Component) Update(args component.Arguments) error {
	// TODO(rfratto):
	//
	// * Ensure server gets restarted with new settings.

	return nil
}
