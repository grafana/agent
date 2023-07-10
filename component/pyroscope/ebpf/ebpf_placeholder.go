//go:build !linux

package ebpf

import (
	"context"

	"github.com/go-kit/log/level"
	"github.com/grafana/agent/component"
)

func init() {
	component.Register(component.Registration{
		Name: "pyroscope.ebpf",
		Args: Arguments{},

		Build: func(opts component.Options, args component.Arguments) (component.Component, error) {
			arguments := args.(Arguments)
			return New(opts, arguments)
		},
	})
}

// Component is a noop placeholder to print a warning when the ebpf component is used but the OS is not linux.
type Component struct {
}

func New(opts component.Options, args Arguments) (component.Component, error) {
	level.Warn(opts.Logger).Log("msg", "the pyroscope.ebpf component only works on linux; enabling it otherwise will do nothing")
	return &Component{}, nil
}

func (i *Component) Run(ctx context.Context) error {
	<-ctx.Done()
	return ctx.Err()
}

func (c *Component) Update(args component.Arguments) error {
	return nil
}
