package sdkconfig

import (
	"context"
	"time"

	"github.com/grafana/agent/component"
	"github.com/grafana/agent/component/module"
	"github.com/grafana/agent/internal/featuregate"
	"github.com/grafana/agent/service/debugdial"
)

func init() {
	component.Register(component.Registration{
		Name:      "sdk.config",
		Stability: featuregate.StabilityExperimental,
		Args:      Arguments{},
		Exports:   module.Exports{},

		Build: func(opts component.Options, args component.Arguments) (component.Component, error) {
			return New(opts, args.(Arguments))
		},
	})
}

var (
	_ component.Component       = (*Component)(nil)
	_ component.HealthComponent = (*Component)(nil)
)

type configStore interface {
	Store(key, value any)
	Delete(key any)
}

func New(o component.Options, args Arguments) (*Component, error) {
	ddServiceData, err := o.GetServiceData(debugdial.ServiceName)
	if err != nil {
		return nil, err
	}

	c := &Component{
		opts:             o,
		prevArgs:         args,
		debugDialService: ddServiceData.(configStore),
		reloadCh:         make(chan struct{}, 1),
	}

	if err = c.Update(args); err != nil {
		return nil, err
	}
	return c, nil
}

// Component implements the local.file component.
type Component struct {
	opts component.Options

	debugDialService configStore
	prevArgs         Arguments
	// reloadCh is a buffered channel which is written to when the watched file
	// should be reloaded by the component.
	reloadCh chan struct{}
}

// CurrentHealth implements component.HealthComponent.
func (c *Component) CurrentHealth() component.Health {
	return component.Health{
		Health:     component.HealthTypeHealthy,
		Message:    "no op",
		UpdateTime: time.Now(),
	}
}

// Run implements component.Component.
func (c *Component) Run(ctx context.Context) error {
	// wait until context gets canceled and we should teardown this
	<-ctx.Done()

	// remove the config from the config store
	for _, sc := range c.prevArgs.Service {
		c.debugDialService.Delete(sc.Name)
	}
	return nil
}

// Update implements component.Component.
func (c *Component) Update(args component.Arguments) error {
	c.syncServices(c.prevArgs, args.(Arguments))
	c.prevArgs = args.(Arguments)
	return nil
}

func (c *Component) syncServices(prevArgs Arguments, args Arguments) {
	// delete removed services
	for _, sc := range prevArgs.Service {
		found := false
		for _, newSc := range args.Service {
			if sc.Name == newSc.Name {
				found = true
			}
		}
		if !found {
			c.debugDialService.Delete(sc.Name)
		}
	}
	// update or add new services
	for _, newSc := range args.Service {
		c.debugDialService.Store(newSc.Name, newSc.Config)
	}
}
