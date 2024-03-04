package sdkconfig

import (
	"context"
	"strings"
	"time"

	"github.com/grafana/agent/internal/component"
	"github.com/grafana/agent/internal/component/module"
	"github.com/grafana/agent/internal/featuregate"
	"github.com/grafana/agent/internal/service/debugdial"
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

	parts := strings.Split(o.ID, ".")
	name := "sdkconfig"
	if len(parts) > 0 {
		name = parts[len(parts)-1]
	}

	c := &Component{
		opts:             o,
		debugDialService: ddServiceData.(configStore),
		reloadCh:         make(chan struct{}, 1),
		name:             name,
	}

	if err = c.Update(args); err != nil {
		return nil, err
	}
	return c, nil
}

// Component implements the local.file component.
type Component struct {
	opts component.Options

	name string

	debugDialService configStore

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
	c.debugDialService.Delete(c.name)

	return nil
}

// Update implements component.Component.
func (c *Component) Update(args component.Arguments) error {
	c.debugDialService.Store(c.name, args.(Arguments).Config)
	return nil
}
