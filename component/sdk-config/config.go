package sdkconfig

import (
	"context"
	"io"
	"sync"
	"time"

	"github.com/grafana/agent/component"
	"github.com/prometheus/client_golang/prometheus"
)

// Config defines the configuration options for the host_info connector.
type Arguments struct {
	// Configuration of the actual service
	Config string `river:"config,attr"`
}

const example = `
	component.sdk-config "serviceA" {
		config = "asdadsasd"
	}

	component.sdk-config "serviceB" {
		config = "abcdef"
	}
`

var (
	_ component.Component       = (*Component)(nil)
	_ component.HealthComponent = (*Component)(nil)
)

func New(o component.Options, args Arguments) (*Component, error) {
	c := &Component{
		opts: o,

		reloadCh: make(chan struct{}, 1),
	}

	err := o.Registerer.Register(c.lastAccessed)
	if err != nil {
		return nil, err
	}
	// Perform an update which will immediately set our exports to the initial
	// contents of the file.
	if err = c.Update(args); err != nil {
		return nil, err
	}
	return c, nil
}

// Component implements the local.file component.
type Component struct {
	opts component.Options

	mut           sync.Mutex
	args          Arguments
	latestContent string
	detector      io.Closer

	healthMut sync.RWMutex
	health    component.Health

	// reloadCh is a buffered channel which is written to when the watched file
	// should be reloaded by the component.
	reloadCh     chan struct{}
	lastAccessed prometheus.Gauge
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
	return nil
}
