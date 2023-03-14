package autodiscovery

import (
	"context"
	"time"

	"github.com/grafana/agent/component"
)

func init() {
	component.Register(component.Registration{
		Name:    "autodiscovery",
		Args:    Arguments{},
		Exports: Exports{},
		Build: func(opts component.Options, args component.Arguments) (component.Component, error) {
			return New(opts, args.(Arguments))
		},
	})
}

type Exports struct {
	Config string `river:"config,attr"`
}

type Arguments struct {
	RefreshPeriod time.Duration `river:"refresh_period,attr,optional"`
	IgnoreList    []string      `river:"ignore_list,attr,optional"`
}

var _ component.Component = (*Component)(nil)

type Component struct {
}

// Run implements component.Compnoent.
func (c *Component) Run(ctx context.Context) error {
}

// Update implements component.Compnoent.
func (c *Component) Update(args component.Arguments) error {
}

func New(o component.Options, args component.Arguments) (*Component, error) {
	c := &Component{
		opts:    o,
		creator: creator,
		// buffered to avoid deadlock from the first immediate update
		newDiscoverer: make(chan struct{}, 1),
	}
	return c, c.Update(args)
}
