package process

import (
	"context"
	"reflect"
	"sync"

	"github.com/grafana/agent/component"
	"github.com/grafana/agent/component/common/loki"
	"github.com/grafana/agent/component/loki/process/internal/stages"
)

func init() {
	component.Register(component.Registration{
		Name:    "loki.process",
		Args:    Arguments{},
		Exports: Exports{},
		Build: func(opts component.Options, args component.Arguments) (component.Component, error) {
			return New(opts, args.(Arguments))
		},
	})
}

// Arguments holds values which are used to configure the loki.process
// component.
type Arguments struct {
	ForwardTo []loki.LogsReceiver  `river:"forward_to,attr"`
	Stages    []stages.StageConfig `river:"stage,enum,optional"`
}

// Exports exposes the receiver that can be used to send log entries to
// loki.process.
type Exports struct {
	Receiver loki.LogsReceiver `river:"receiver,attr"`
}

var (
	_ component.Component = (*Component)(nil)
)

// Component implements the loki.process component.
type Component struct {
	opts component.Options

	mut          sync.RWMutex
	receiver     loki.LogsReceiver
	fanout       []loki.LogsReceiver
	processIn    chan<- loki.Entry
	processOut   chan loki.Entry
	entryHandler loki.EntryHandler
	stages       []stages.StageConfig
}

// New creates a new loki.process component.
func New(o component.Options, args Arguments) (*Component, error) {
	c := &Component{
		opts: o,
	}

	// Create and immediately export the receiver which remains the same for
	// the component's lifetime.
	c.receiver = make(loki.LogsReceiver)
	c.processOut = make(loki.LogsReceiver)
	o.OnStateChange(Exports{Receiver: c.receiver})

	// Call to Update() to start readers and set receivers once at the start.
	if err := c.Update(args); err != nil {
		return nil, err
	}

	return c, nil
}

// Run implements component.Component.
func (c *Component) Run(ctx context.Context) error {
	defer func() {
		c.mut.RLock()
		if c.entryHandler != nil {
			c.entryHandler.Stop()
		}
		close(c.processOut)
		close(c.processIn)
		c.mut.RUnlock()
	}()
	wg := &sync.WaitGroup{}
	wg.Add(2)
	go c.handleIn(ctx, wg)
	go c.handleOut(ctx, wg)

	wg.Wait()
	return nil
}

// Update implements component.Component.
func (c *Component) Update(args component.Arguments) error {
	newArgs := args.(Arguments)

	c.mut.Lock()
	defer c.mut.Unlock()

	// We want to create a new pipeline if the config changed or if this is the
	// first load. This will allow a component with no stages to function
	// properly.
	if stagesChanged(c.stages, newArgs.Stages) || c.stages == nil {
		if c.entryHandler != nil {
			c.entryHandler.Stop()
		}

		pipeline, err := stages.NewPipeline(c.opts.Logger, newArgs.Stages, &c.opts.ID, c.opts.Registerer)
		if err != nil {
			return err
		}
		c.entryHandler = loki.NewEntryHandler(c.processOut, func() {})
		c.processIn = pipeline.Wrap(c.entryHandler).Chan()
		c.stages = newArgs.Stages
	}

	c.fanout = newArgs.ForwardTo

	return nil
}

func (c *Component) handleIn(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		select {
		case <-ctx.Done():
			return
		case entry := <-c.receiver:
			c.mut.RLock()
			select {
			case <-ctx.Done():
				return
			case c.processIn <- entry:
				// no-op
			}
			c.mut.RUnlock()
		}
	}
}

func (c *Component) handleOut(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		select {
		case <-ctx.Done():
			return
		case entry := <-c.processOut:
			c.mut.RLock()
			for _, f := range c.fanout {
				select {
				case <-ctx.Done():
					return
				case f <- entry:
					// no-op
				}
			}
			c.mut.RUnlock()
		}
	}
}

func stagesChanged(prev, next []stages.StageConfig) bool {
	if len(prev) != len(next) {
		return true
	}
	for i := range prev {
		if !reflect.DeepEqual(prev[i], next[i]) {
			return true
		}
	}
	return false
}
