package scrape

import (
	"context"
	"fmt"

	"github.com/grafana/agent/component"
)

func init() {
	component.Register(component.Registration{
		Name: "metrics.scrape",
		Args: Arguments{},
		// Exports: Exports{},

		Build: func(opts component.Options, args component.Arguments) (component.Component, error) {
			return New(opts, args.(Arguments))
		},
	})
}

type Target map[string]string

// Arguments holds values which are used to configure the metrics.scrape component.
type Arguments struct {
	Targets []Target `hcl:"targets"`
}

// Exports holds values which are exported by the metrics.scrape component.
type Exports struct{}

// Component implements the metrics.Scrape component.
type Component struct {
	opts component.Options
}

var (
	_ component.Component = (*Component)(nil)
)

// New creates a new metrics.scrape component.
func New(o component.Options, args Arguments) (*Component, error) {
	c := &Component{opts: o}

	// Call to Update() to set the output once at the start
	if err := c.Update(args); err != nil {
		return nil, err
	}

	return c, nil
}

// Run implements component.Component.
func (c *Component) Run(ctx context.Context) error {
	fmt.Println("Running", c.opts.ID)
	<-ctx.Done()
	return nil
}

// Update implements component.Component.
func (c *Component) Update(args component.Arguments) error {
	fmt.Println("Updated", c.opts.ID)
	// c.opts.OnStateChange(Exports{
	// 	Output: targets,
	// })
	return nil
}
