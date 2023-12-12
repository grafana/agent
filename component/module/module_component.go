package module

import (
	"context"
	"sync"
	"sync/atomic"

	"github.com/grafana/agent/component"
)

func init() {
	component.Register(component.Registration{
		Name:    "module_component",
		Args:    Arguments{},
		Exports: Exports{},

		Build: func(opts component.Options, args component.Arguments) (component.Component, error) {
			return New(opts, args.(Arguments))
		},
	})
}

// Arguments holds values which are used to configure the module.
type Arguments = map[string]any

// Component implements the module.file component.
type Component struct {
	opts component.Options
	mod  *ModuleComponent

	mut         sync.RWMutex
	args        Arguments
	content     string
	updatedOnce atomic.Bool
}

var (
	_ component.Component       = (*Component)(nil)
	_ component.HealthComponent = (*Component)(nil)
)

// New creates a new module.file component.
func New(o component.Options, args Arguments) (*Component, error) {
	m, err := NewModuleComponent(o)
	if err != nil {
		return nil, err
	}

	c := &Component{
		opts: o,
		mod:  m,
		args: args,
	}
	// we don't update on create because we don't have the content yet
	return c, nil
}

// Run implements component.Component.
func (c *Component) Run(ctx context.Context) error {
	c.mod.RunFlowController(ctx)
	return nil
}

// Update implements component.Component.
func (c *Component) Update(args component.Arguments) error {
	newArgs := args.(Arguments)
	c.setArgs(newArgs)
	c.updatedOnce.Store(true)
	return c.reload()
}

// UpdateContent reloads the module with a new config
func (c *Component) UpdateContent(content string) error {
	if content != c.getContent() {
		c.setContent(content)
		return c.reload()
	}
	return nil
}

func (c *Component) reload() error {
	if c.getContent() == "" || !c.updatedOnce.Load() {
		return nil // the module is not yet ready
	}
	return c.mod.LoadFlowSource(c.getArgs(), c.getContent())
}

// CurrentHealth implements component.HealthComponent.
func (c *Component) CurrentHealth() component.Health {
	return c.mod.CurrentHealth()
}

// getArgs is a goroutine safe way to get args
func (c *Component) getArgs() Arguments {
	c.mut.RLock()
	defer c.mut.RUnlock()
	return c.args
}

// setArgs is a goroutine safe way to set args
func (c *Component) setArgs(args Arguments) {
	c.mut.Lock()
	c.args = args
	c.mut.Unlock()
}

// getContent is a goroutine safe way to get content
func (c *Component) getContent() string {
	c.mut.RLock()
	defer c.mut.RUnlock()
	return c.content
}

// setContent is a goroutine safe way to set content
func (c *Component) setContent(content string) {
	c.mut.Lock()
	c.content = content
	c.mut.Unlock()
}
