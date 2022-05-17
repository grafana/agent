// Package flow implements a component graph system.
package flow

import (
	"context"
	"sync"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/grafana/agent/pkg/flow/internal/controller"
	"github.com/hashicorp/hcl/v2"
)

// Options holds static options for a flow controller.
type Options struct {
	// TODO(rfratto): replace Logger below with an io.Writer and have the
	// Controller manage the logger instead.

	// Optional logger where components will log to.
	Logger log.Logger

	// Directory where components can write data. Components will create
	// subdirectories for component-specific data.
	DataPath string
}

// Controller is the Flow controller system.
type Controller struct {
	log  log.Logger
	opts Options

	updateQueue *controller.Queue
	sched       *controller.Scheduler
	loader      *controller.Loader

	cancel       context.CancelFunc
	exited       chan struct{}
	loadFinished chan struct{}

	loadMut    sync.RWMutex
	loadedOnce bool
}

// NewController creates and starts a new Flow controller. Call Close to stop
// the controller.
func NewController(o Options) *Controller {
	c, ctx := newController(o)
	go c.run(ctx)
	return c
}

func newController(o Options) (*Controller, context.Context) {
	ctx, cancel := context.WithCancel(context.Background())

	var (
		queue  = controller.NewQueue()
		sched  = controller.NewScheduler()
		loader = controller.NewLoader(controller.ComponentOptions{
			Logger:   o.Logger,
			DataPath: o.DataPath,
			OnExportsChange: func(cn *controller.ComponentNode) {
				// Changed components should be queued for reevaluation.
				queue.Enqueue(cn)
			},
		})
	)

	return &Controller{
		log:  o.Logger,
		opts: o,

		updateQueue: queue,
		sched:       sched,
		loader:      loader,

		cancel:       cancel,
		exited:       make(chan struct{}, 1),
		loadFinished: make(chan struct{}, 1),
	}, ctx
}

func (c *Controller) run(ctx context.Context) {
	defer close(c.exited)
	defer level.Debug(c.log).Log("msg", "flow controller exiting")

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	for {
		select {
		case <-ctx.Done():
			return

		case <-c.updateQueue.Chan():
			updated := c.updateQueue.TryDequeue()
			if updated != nil {
				level.Debug(c.log).Log("msg", "handling component with updated state", "node_id", updated.NodeID())
				c.loader.Reevaluate(rootEvalContext, updated)
			}

		case <-c.loadFinished:
			level.Info(c.log).Log("msg", "scheduling loaded components")

			components := c.loader.Components()
			runnables := make([]controller.RunnableNode, 0, len(components))
			for _, uc := range components {
				runnables = append(runnables, uc)
			}
			c.sched.Synchronize(runnables)
		}
	}
}

// LoadFile synchronizes the state of the controller with the current config
// file. Components in the graph will be marked as unhealthy if there was an
// error encountered during Load.
//
// The controller will only start running components after Load is called once
// without any configuration errors.
//
// LoadFile will return an error value of hcl.Diagnostics. hcl.Diagnostics is
// used to report both warnings and configuration errors.
func (c *Controller) LoadFile(f *File) error {
	c.loadMut.Lock()
	defer c.loadMut.Unlock()

	diags := c.loader.Apply(rootEvalContext, f.Components)
	if !c.loadedOnce && diags.HasErrors() {
		// The first call to Load should not run any components if there were
		// errors in the coniguration file.
		return diags
	}
	c.loadedOnce = true

	select {
	case c.loadFinished <- struct{}{}:
	default:
		// A refresh is already scheduled
	}
	return diagsOrNil(diags)
}

func diagsOrNil(d hcl.Diagnostics) error {
	if len(d) > 0 {
		return d
	}
	return nil
}

// Close closes the controller and all running components.
func (c *Controller) Close() error {
	c.cancel()
	<-c.exited
	return c.sched.Close()
}
