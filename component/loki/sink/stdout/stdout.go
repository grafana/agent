package stdout

import (
	"context"
	"fmt"
	"sync"

	"github.com/grafana/agent/component"
	"github.com/grafana/loki/clients/pkg/promtail/api"
)

func init() {
	component.Register(component.Registration{
		Name:    "loki.sink.stdout",
		Args:    Arguments{},
		Exports: Exports{},

		Build: func(opts component.Options, args component.Arguments) (component.Component, error) {
			return New(opts, args.(Arguments))
		},
	})
}

const (
	pathLabel = "__path__"
)

// Arguments holds values which are used to configure the loki.sink.stdout
// component.
type Arguments struct{}

// Exports holds the values exported by the loki.sink.stdout component.
type Exports struct {
	Receiver chan api.Entry `river:"receiver,attr"`
}

// DefaultArguments defines the default settings for log scraping.
var DefaultArguments = Arguments{}

// UnmarshalRiver implements river.Unmarshaler.
func (arg *Arguments) UnmarshalRiver(f func(interface{}) error) error {
	*arg = DefaultArguments

	type args Arguments
	return f((*args)(arg))
}

var (
	_ component.Component = (*Component)(nil)
)

// Component implements the loki.source.file component.
type Component struct {
	opts component.Options

	mut      sync.RWMutex
	args     Arguments
	receiver chan api.Entry
}

// New creates a new loki.sink.stdout component.
func New(o component.Options, args Arguments) (*Component, error) {
	ch := make(chan api.Entry)
	c := &Component{
		opts:     o,
		receiver: ch,
	}

	// Call to Update() once at the start.
	if err := c.Update(args); err != nil {
		return nil, err
	}

	// Immediately export the receiver which remains the same for the component
	// lifetime.
	o.OnStateChange(Exports{Receiver: c.receiver})

	return c, nil
}

// Run implements component.Component.
func (c *Component) Run(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		case entry := <-c.receiver:
			fmt.Printf("Receiver %s got entry %s\n", c.opts.ID, entry)
		}
	}
}

// Update implements component.Component.
func (c *Component) Update(args component.Arguments) error {
	newArgs := args.(Arguments)

	c.mut.Lock()
	defer c.mut.Unlock()
	c.args = newArgs

	return nil
}
