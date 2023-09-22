package module

import (
	"context"
	"fmt"
	"reflect"
	"sync"
	"time"

	"github.com/grafana/agent/component"
)

// ModuleComponent holds the common properties for module components.
type ModuleComponent struct {
	opts component.Options
	mod  component.Module

	mut           sync.RWMutex
	health        component.Health
	latestContent string
	latestArgs    map[string]any
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
	c.mod, err = o.ModuleController.NewModule("", func(exports map[string]any) {
		c.opts.OnStateChange(Exports{Exports: exports})
	})
	return c, err
}

// LoadFlowSource loads the flow controller with the current component source.
// It will set the component health in addition to return the error so that the consumer can rely on either or both.
// If the content is the same as the last time it was successfully loaded, it will not be reloaded.
func (c *ModuleComponent) LoadFlowSource(args map[string]any, contentValue string) error {
	if reflect.DeepEqual(args, c.getLatestArgs()) && contentValue == c.getLatestContent() {
		return nil
	}

	err := c.mod.LoadConfig([]byte(contentValue), args)
	if err != nil {
		c.setHealth(component.Health{
			Health:     component.HealthTypeUnhealthy,
			Message:    fmt.Sprintf("failed to load module content: %s", err),
			UpdateTime: time.Now(),
		})

		return err
	}

	c.setLatestArgs(args)
	c.setLatestContent(contentValue)
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

// SetHealth contains the implementation details for setHealth in a module component.
func (c *ModuleComponent) setHealth(h component.Health) {
	c.mut.Lock()
	defer c.mut.Unlock()
	c.health = h
}

func (c *ModuleComponent) setLatestContent(content string) {
	c.mut.Lock()
	defer c.mut.Unlock()
	c.latestContent = content
}

func (c *ModuleComponent) getLatestContent() string {
	c.mut.RLock()
	defer c.mut.RUnlock()
	return c.latestContent
}

func (c *ModuleComponent) setLatestArgs(args map[string]any) {
	c.mut.Lock()
	defer c.mut.Unlock()

	c.latestArgs = make(map[string]any)
	for key, value := range args {
		c.latestArgs[key] = value
	}
}

func (c *ModuleComponent) getLatestArgs() map[string]any {
	c.mut.RLock()
	defer c.mut.RUnlock()
	return c.latestArgs
}
