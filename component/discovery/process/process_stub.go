//go:build !linux

package process

import (
	"context"

	"github.com/grafana/agent/component"
	"github.com/grafana/agent/component/discovery"
	"github.com/grafana/agent/pkg/flow/logging/level"
)

func init() {
	component.Register(component.Registration{
		Name:    "discovery.process",
		Args:    Arguments{},
		Exports: discovery.Exports{},

		Build: func(opts component.Options, args component.Arguments) (component.Component, error) {
			return New(opts, args.(Arguments))
		},
	})
}

func New(opts component.Options, args Arguments) (*Component, error) {
	_ = level.Warn(opts.Logger).Log("msg", "the discovery.process component only works on linux; enabling it otherwise will do nothing")
	return &Component{}, nil
}

type Component struct {
}

func (c *Component) Run(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return nil
	}
}

func (c *Component) Update(args component.Arguments) error {
	return nil
}
