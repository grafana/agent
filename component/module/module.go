package module

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/grafana/agent/component"
)

// ModuleComponent holds the common properties for module components.
type ModuleComponent struct {
	opts component.Options
	mod  component.Module

	mut    sync.RWMutex
	health component.Health
}

// Exports holds values which are exported from the run module.
type Exports struct {
	// Exports exported from the running module.
	Exports map[string]any `river:"exports,block"`
}

// NewModuleComponent initializes a new ModuleComponent.
func NewModuleComponent(o component.Options) (*ModuleComponent, error) {
	c := &ModuleComponent{
		opts: o,
	}
	var err error
	c.mod, err = o.ModuleController.NewModule(o.ID, func(exports map[string]any) {
		c.opts.OnStateChange(Exports{Exports: exports})
	})
	return c, err
}

// LoadFlowContent loads the flow controller with the current component content. It
// will set the component health in addition to return the error so that the consumer
// can rely on either or both.
func (c *ModuleComponent) LoadFlowContent(args map[string]any, contentValue string) error {
	err := c.mod.LoadConfig([]byte(contentValue), args)
	if err != nil {
		c.setHealth(component.Health{
			Health:     component.HealthTypeUnhealthy,
			Message:    fmt.Sprintf("failed to load module content: %s", err),
			UpdateTime: time.Now(),
		})

		return err
	}

	c.setHealth(component.Health{
		Health:     component.HealthTypeHealthy,
		Message:    "module content loaded",
		UpdateTime: time.Now(),
	})
	return nil
}

// RunFlowController runs the flow controller that all module components start.
func (c *ModuleComponent) RunFlowController(ctx context.Context) {
	c.mod.Run(ctx)
}

// CurrentHealth contains the implementation details for CurrentHealth in a module component.
func (c *ModuleComponent) CurrentHealth() component.Health {
	c.mut.RLock()
	defer c.mut.RUnlock()
	return c.health
}

// HTTPHandler returns the underlying module handle for the UI.
func (c *ModuleComponent) HTTPHandler() http.Handler {
	return c.mod.ComponentHandler()
}

// SetHealth contains the implementation details for setHealth in a module component.
func (c *ModuleComponent) setHealth(h component.Health) {
	c.mut.Lock()
	defer c.mut.Unlock()
	c.health = h
}
