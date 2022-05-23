// Package flow implements the Flow component graph system. Flow configuration
// files are parsed from HCL, which contain a listing of components to run.
//
// Components
//
// Each component has a set of arguments (HCL attributes and blocks) and
// optionally a set of exported fields. Components can reference the arguments
// or exports of other components using HCL expressions.
//
// See the top-level component package for more information on components, and
// subpackages for defined components.
//
// Component Health
//
// A component will have various health states during its lifetime:
//
//     1. Unknown:   The initial health state for new components.
//     2. Healthy:   A healthy component
//     3. Unhealthy: An unhealthy component.
//     4. Exited:    A component which is no longer running.
//
// Health states are paired with a time for when the health state was generated
// and a message providing more detail for the health state.
//
// Components can report their own health states. The health state reported by
// a component is merged with the Flow-level health of that component: an error
// when evaluating the configuration for a component will always be reported as
// unhealthy until the next successful evaluation.
//
// Component Evaluation
//
// The process of converting the HCL block associated with a component into the
// appropriate Go struct is called "component evaluation."
//
// Components are only evaluated after all components they reference have been
// evaluated; cyclic dependencies are invalid.
//
// If a component updates its Exports at runtime, other components which directly
// or indirectly reference the updated component will have their Arguments
// re-evaluated.
//
// The arguments and exports for a component will be left in their last valid
// state if a component shuts down or is given an invalid config. This prevents
// a domino effect of a single failed component taking down other components
// which are otherwise healthy.
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

// Flow is the Flow system.
type Flow struct {
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

// New creates and starts a new Flow controller. Call Close to stop
// the controller.
func New(o Options) *Flow {
	c, ctx := NewFlow(o)
	go c.run(ctx)
	return c
}

func NewFlow(o Options) (*Flow, context.Context) {
	ctx, cancel := context.WithCancel(context.Background())

	var (
		queue  = controller.NewQueue()
		sched  = controller.NewScheduler()
		loader = controller.NewLoader(controller.ComponentGlobals{
			Logger:   o.Logger,
			DataPath: o.DataPath,
			OnExportsChange: func(cn *controller.ComponentNode) {
				// Changed components should be queued for reevaluation.
				queue.Enqueue(cn)
			},
		})
	)

	return &Flow{
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

func (c *Flow) run(ctx context.Context) {
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
			err := c.sched.Synchronize(runnables)
			if err != nil {
				level.Error(c.log).Log("msg", "failed to load components", "err", err)
			}
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
func (c *Flow) LoadFile(f *File) error {
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
func (c *Flow) Close() error {
	c.cancel()
	<-c.exited
	return c.sched.Close()
}
