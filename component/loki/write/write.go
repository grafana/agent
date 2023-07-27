package write

import (
	"context"
	"fmt"
	"sync"

	"github.com/grafana/agent/component"
	"github.com/grafana/agent/component/common/loki"
	"github.com/grafana/agent/component/common/loki/client"
	"github.com/grafana/agent/component/common/loki/limit"
	"github.com/grafana/agent/component/common/loki/wal"
	"github.com/grafana/agent/pkg/build"
)

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
	WAL            wal.Config        `river:"wal,block,optional"`
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

	mut       sync.RWMutex
	args      Arguments
	receiver  loki.LogsReceiver
	manager   client.Client
	walWriter *wal.Writer
	to        loki.EntryHandler
}

// New creates a new loki.write component.
func New(o component.Options, args Arguments) (*Component, error) {
	c := &Component{
		opts:    o,
		metrics: client.NewMetrics(o.Registerer),
	}

	// Create and immediately export the receiver which remains the same for
	// the component's lifetime.
	c.receiver = loki.NewLogsReceiver()
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
		case entry := <-c.receiver.Chan():
			select {
			case <-ctx.Done():
				return nil
			case c.to.Chan() <- entry:
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

	if c.walWriter != nil {
		c.walWriter.Stop()
	}
	if c.manager != nil {
		c.manager.Stop()
	}

	cfgs := newArgs.convertClientConfigs()

	var err error
	var notifier client.WriterEventsNotifier = client.NilNotifier
	c.walWriter = nil
	if newArgs.WAL.Enabled {
		c.walWriter, err = wal.NewWriter(newArgs.WAL, c.opts.Logger, c.opts.Registerer)
		if err != nil {
			return fmt.Errorf("error creating wal writer")
		}
		notifier = c.walWriter
		c.to = c.walWriter
	}

	c.manager, err = client.NewManager(c.metrics, c.opts.Logger, limit.Config{
		MaxStreams: newArgs.MaxStreams,
	}, c.opts.Registerer, newArgs.WAL, notifier, cfgs...)

	if !newArgs.WAL.Enabled {
		c.to = c.manager
	}

	return err
}
