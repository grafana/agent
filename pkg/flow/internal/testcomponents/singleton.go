package testcomponents

import (
	"context"

	"github.com/go-kit/log"
	"github.com/grafana/agent/component"
)

func init() {
	component.Register(component.Registration{
		Name:      "testcomponents.singleton",
		Args:      SingletonArguments{},
		Exports:   SingletonExports{},
		Singleton: true,

		Build: func(opts component.Options, args component.Arguments) (component.Component, error) {
			return NewSingleton(opts, args.(SingletonArguments))
		},
	})
}

// SingletonArguments configures the testcomponents.singleton component.
type SingletonArguments struct{}

// SingletonExports describes exported fields for the
// testcomponents.singleton component.
type SingletonExports struct{}

// Singleton implements the testcomponents.singleton component, which is a
// no-op component.
type Singleton struct {
	opts component.Options
	log  log.Logger
}

// NewSingleton creates a new singleton component.
func NewSingleton(o component.Options, cfg SingletonArguments) (*Singleton, error) {
	t := &Singleton{opts: o, log: o.Logger}
	if err := t.Update(cfg); err != nil {
		return nil, err
	}
	return t, nil
}

var (
	_ component.Component = (*Passthrough)(nil)
)

// Run implements Component.
func (t *Singleton) Run(ctx context.Context) error {
	<-ctx.Done()
	return nil
}

// Update implements Component.
func (t *Singleton) Update(args component.Arguments) error {
	return nil
}
