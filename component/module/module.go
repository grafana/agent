package module

import (
	"context"
	"fmt"
	"reflect"
	"sync"
	"time"

	"github.com/go-kit/log"
	"github.com/grafana/agent/component"
	"github.com/grafana/agent/pkg/flow/logging/level"
)

// ModuleComponent holds the common properties for module components.
type ModuleComponent struct {
	opts ModuleComponentOptions
	mod  component.Module

	mut           sync.RWMutex
	health        component.Health
	latestContent string
	latestArgs    map[string]any
}

// ModuleComponentOptions holds the required options to create a ModuleComponent.
type ModuleComponentOptions struct {
	// ModuleController allows for the creation of modules.
	ModuleController component.ModuleController

	// ID of the custom component. Guaranteed to be globally unique across all running components.
	ID string

	// Logger the custom component may use for logging. Logs emitted with the logger
	// always include the component ID as a field.
	Logger log.Logger

	// OnStateChange may be invoked at any time by the custom component when its exports values change.
	// The Flow controller then will queue re-processing components which depend on the changed component.
	OnStateChange func(e component.Exports)
}

// Deprecated: Exports holds values which are exported from the run module. New modules use map[string]any directly.
type Exports struct {
	// Exports exported from the running module.
	Exports map[string]any `river:"exports,block"`
}

// Deprecated: Use NewModuleComponentV2 instead.
func NewModuleComponent(o component.Options) (*ModuleComponent, error) {
	c := &ModuleComponent{
		opts: ModuleComponentOptions{
			ModuleController: o.ModuleController,
			ID:               o.ID,
			Logger:           o.Logger,
			OnStateChange:    o.OnStateChange,
		},
	}
	var err error
	c.mod, err = o.ModuleController.NewModule("", func(exports map[string]any) {
		c.opts.OnStateChange(Exports{Exports: exports})
	})
	return c, err
}

// NewModuleComponentV2 initializes a new ModuleComponent.
func NewModuleComponentV2(o ModuleComponentOptions) (*ModuleComponent, error) {
	c := &ModuleComponent{
		opts: o,
	}
	var err error
	c.mod, err = o.ModuleController.NewModule("", func(exports map[string]any) {
		c.opts.OnStateChange(exports)
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
	err := c.mod.Run(ctx)
	if err != nil {
		level.Error(c.opts.Logger).Log("msg", "error running module", "id", c.opts.ID, "err", err)
	}
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
