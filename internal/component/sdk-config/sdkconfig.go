package sdkconfig

import (
	"context"
	"io"
	"strings"
	"sync"
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
}

func New(o component.Options, args Arguments) (*Component, error) {
	ddServiceData, err := o.GetServiceData(debugdial.ServiceName)
	if err != nil {
		return nil, err
	}

	c := &Component{
		opts:             o,
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

	mut      sync.Mutex
	args     Arguments
	detector io.Closer

	healthMut sync.RWMutex
	health    component.Health

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
	<-ctx.Done()
	return nil
}

// Update implements component.Component.
func (c *Component) Update(args component.Arguments) error {
	parts := strings.Split(c.opts.ID, ".")
	serviceName := parts[len(parts)-1]
	c.debugDialService.Store(serviceName, args.(Arguments).Config)
	return nil
}
