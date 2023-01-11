//go:build !windows

package windowsevent

import (
	"context"

	"github.com/go-kit/log/level"

	"github.com/grafana/agent/component"
)

func init() {
	component.Register(component.Registration{
		Name: "loki.source.windowsevent",
		Args: Arguments{},

		Build: func(opts component.Options, args component.Arguments) (component.Component, error) {
			level.Info(opts.Logger).Log("msg", "loki.source.windowsevent only works on windows platforms")
			return &FakeComponent{}, nil
		},
	})
}

var (
	_ component.Component = (*FakeComponent)(nil)
)

// FakeComponent implements the loki.source.windowsevent component for non-windows environments.
type FakeComponent struct {
}

func (f *FakeComponent) Run(ctx context.Context) error {
	<-ctx.Done()
	return nil
}

func (f *FakeComponent) Update(_ component.Arguments) error {
	return nil
}
