//go:build !windows
// +build !windows

package windowsevents

import (
	"context"
	"github.com/grafana/agent/component"
)

func init() {
	component.Register(component.Registration{
		Name: "loki.source.windowsevents",
		Args: Arguments{},

		Build: func(opts component.Options, args component.Arguments) (component.Component, error) {
			return &FakeComponent{}, nil
		},
	})
}

var (
	_ component.Component = (*FakeComponent)(nil)
)

// FakeComponent implements the loki.source.windowsevents component for non-windows environments.
type FakeComponent struct {
}

func (f *FakeComponent) Run(ctx context.Context) error {
	<-ctx.Done()
	return nil
}

func (f *FakeComponent) Update(args component.Arguments) error {
	return nil
}
