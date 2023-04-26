package write

import (
	"context"
	"fmt"
	"sync"

	"github.com/grafana/agent/component/common/loki/client"

	"github.com/grafana/agent/component"
	"github.com/grafana/agent/component/common/loki"
	"github.com/grafana/agent/pkg/build"
)

var streamLagLabels = []string{"filename"}

func init() {
	component.Register(component.Registration{
		Name:    "loki.write",
		Args:    Arguments{},
		Exports: Exports{},

		Build: func(opts component.Options, args component.Arguments) (component.Component, error) {
			return New(opts, args.(Arguments))
		},
	})

	client.UserAgent = fmt.Sprintf("GrafanaAgent/%s", build.Version)
}

// Arguments holds values which are used to configure the loki.write component.
type Arguments struct {
	Endpoints      []EndpointOptions `river:"endpoint,block,optional"`
	ExternalLabels map[string]string `river:"external_labels,attr,optional"`
	MaxStreams     int               `river:"max_streams,attr,optional"`
}

// Exports holds the receiver that is used to send log entries to the
// loki.write component.
type Exports struct {
	Receiver loki.LogsReceiver `river:"receiver,attr"`
}

var (
	_ component.Component = (*Component)(nil)
)

// Component implements the loki.write component.
type Component struct {
	opts    component.Options
	metrics *client.Metrics

	mut      sync.RWMutex
	args     Arguments
	receiver loki.LogsReceiver
	clients  []client.Client
}

// New creates a new loki.write component.
func New(o component.Options, args Arguments) (*Component, error) {
	c := &Component{
		opts:    o,
		metrics: client.NewMetrics(o.Registerer, streamLagLabels),
	}

	// Create and immediately export the receiver which remains the same for
	// the component's lifetime.
	c.receiver = make(loki.LogsReceiver)
	o.OnStateChange(Exports{Receiver: c.receiver})

	// Call to Update() to start readers and set receivers once at the start.
	if err := c.Update(args); err != nil {
		return nil, err
	}

	return c, nil
}

// Run implements component.Component.
func (c *Component) Run(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		case entry := <-c.receiver:
			for _, client := range c.clients {
				if client != nil {
					select {
					case <-ctx.Done():
						return nil
					case client.Chan() <- entry:
						// no-op
					}
				}
			}
		}
	}
}

// Update implements component.Component.
func (c *Component) Update(args component.Arguments) error {
	newArgs := args.(Arguments)

	c.mut.Lock()
	defer c.mut.Unlock()
	c.args = newArgs

	for _, client := range c.clients {
		if client != nil {
			client.Stop()
		}
	}
	c.clients = make([]client.Client, len(newArgs.Endpoints))

	cfgs := newArgs.convertClientConfigs()
	// TODO (@tpaschalis) We could use a client.NewMulti here to push the
	// fanout logic back to the client layer, but I opted to keep it explicit
	// here a) for easier debugging and b) possible improvements in the future.
	for _, cfg := range cfgs {
		client, err := client.New(c.metrics, cfg, streamLagLabels, newArgs.MaxStreams, c.opts.Logger)
		if err != nil {
			return err
		}
		c.clients = append(c.clients, client)
	}

	return nil
}
