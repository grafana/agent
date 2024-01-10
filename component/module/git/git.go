// Package git implements the module.git component.
package git

import (
	"context"
	"sync"

	"go.uber.org/atomic"

	"github.com/grafana/agent/component"
	"github.com/grafana/agent/component/module"
	remote_git "github.com/grafana/agent/component/remote/git"
)

func init() {
	component.Register(component.Registration{
		Name:    "module.git",
		Args:    Arguments{},
		Exports: module.Exports{},

		Build: func(opts component.Options, args component.Arguments) (component.Component, error) {
			return New(opts, args.(Arguments))
		},
	})
}

// Arguments holds values which are used to configure the module.git component.
type Arguments struct {
	RemoteGitArguments remote_git.Arguments `river:",squash"`

	Arguments map[string]any `river:"arguments,block,optional"`
}

// SetToDefault implements river.Defaulter.
func (args *Arguments) SetToDefault() {
	args.RemoteGitArguments.SetToDefault()
}

// Component implements the module.git component.
type Component struct {
	opts component.Options
	mod  *module.ModuleComponent

	mut     sync.RWMutex
	args    Arguments
	content string

	managedRemoteGit *remote_git.Component
	inUpdate         atomic.Bool
	isCreated        atomic.Bool
}

var (
	_ component.Component       = (*Component)(nil)
	_ component.HealthComponent = (*Component)(nil)
)

// New creates a new module.git component.
func New(o component.Options, args Arguments) (*Component, error) {
	m, err := module.NewModuleComponent(o)
	if err != nil {
		return nil, err
	}
	c := &Component{
		opts: o,
		args: args,
		mod:  m,
	}
	defer c.isCreated.Store(true)

	c.managedRemoteGit, err = c.newManagedLocalComponent(o)
	if err != nil {
		return nil, err
	}
	if err := c.Update(args); err != nil {
		return nil, err
	}

	return c, nil
}

// NewManagedLocalComponent creates the new remote.git managed component.
func (c *Component) newManagedLocalComponent(o component.Options) (*remote_git.Component, error) {
	remoteGitOpts := o
	remoteGitOpts.OnStateChange = func(e component.Exports) {
		c.setContent(e.(remote_git.Exports).Content.Value)

		if !c.inUpdate.Load() && c.isCreated.Load() {
			// Any errors found here are reported via component health
			_ = c.mod.LoadFlowSource(c.getArgs().Arguments, c.getContent())
		}
	}

	return remote_git.New(remoteGitOpts, c.getArgs().RemoteGitArguments)
}

// Run implements component.Component.
func (c *Component) Run(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	ch := make(chan error, 1)
	go func() {
		err := c.managedRemoteGit.Run(ctx)
		if err != nil {
			ch <- err
		}
	}()

	go c.mod.RunFlowController(ctx)

	for {
		select {
		case <-ctx.Done():
			return nil
		case err := <-ch:
			return err
		}
	}
}

// Update implements component.Component.
func (c *Component) Update(args component.Arguments) error {
	c.inUpdate.Store(true)
	defer c.inUpdate.Store(false)

	newArgs := args.(Arguments)
	c.setArgs(newArgs)

	err := c.managedRemoteGit.Update(newArgs.RemoteGitArguments)
	if err != nil {
		return err
	}

	// Force a content load here and bubble up any error. This will catch problems
	// on initial load.
	return c.mod.LoadFlowSource(newArgs.Arguments, c.getContent())
}

// CurrentHealth implements component.HealthComponent.
func (c *Component) CurrentHealth() component.Health {
	leastHealthy := component.LeastHealthy(
		c.managedRemoteGit.CurrentHealth(),
		c.mod.CurrentHealth(),
	)

	if leastHealthy.Health == component.HealthTypeHealthy {
		return c.mod.CurrentHealth()
	}
	return leastHealthy
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
