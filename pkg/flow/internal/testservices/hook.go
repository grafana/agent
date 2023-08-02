package testservices

import (
	"context"
	"fmt"

	"github.com/grafana/agent/component"
	"github.com/grafana/agent/service"
)

// The Hook service allows components to invoke hooks used for testing.
type Hook struct {
	hooks Hooks
}

// NewHook creates a new Hook service.
func NewHook(hooks Hooks) *Hook {
	return &Hook{hooks: hooks}
}

// Hooks is a set of hooks that are exposed to a component using the hook
// service.
type Hooks struct {
	OnComponentCreate func(o component.Options, args component.Arguments)
}

var _ service.Service = (*Hook)(nil)

// Definition implements [service.Serivce]. The name of the returned service is
// "test".
func (h *Hook) Definition() service.Definition {
	return service.Definition{
		Name:       "hook",
		ConfigType: nil,
		DependsOn:  nil,
	}
}

// Run implements [service.Service].
func (h *Hook) Run(ctx context.Context, host service.Host) error {
	<-ctx.Done()
	return nil
}

// Update implements [service.Service].
func (h *Hook) Update(newConfig any) error {
	return fmt.Errorf("Update should not have been called")
}

// Data implements [service.Service]. It returns a *testing.T.
func (h *Hook) Data() any { return h.hooks }
