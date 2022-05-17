// Package flow implements a component graph system.
package flow

import (
	"context"
	"sync"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/grafana/agent/pkg/flow/internal/controller"
	"github.com/grafana/agent/pkg/flow/internal/dag"
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
	cache       *controller.ValueCache
	sched       *controller.Scheduler

	cancel       context.CancelFunc
	exited       chan struct{}
	loadFinished chan struct{}

	graphMut   sync.RWMutex
	graph      *dag.Graph
	loadedOnce bool
	components []*controller.ComponentNode

	// Callbacks used for testing
	onComponentChanged func(uc *controller.ComponentNode)
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

	return &Controller{
		log:  o.Logger,
		opts: o,

		updateQueue: controller.NewQueue(),
		cache:       controller.NewValueCache(),
		sched:       controller.NewScheduler(),

		cancel:       cancel,
		exited:       make(chan struct{}, 1),
		loadFinished: make(chan struct{}, 1),

		graph: &dag.Graph{},
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
				c.handleUpdatedComponent(updated)
			}

		case <-c.loadFinished:
			level.Info(c.log).Log("msg", "scheduling loaded components")

			c.graphMut.RLock()
			components := c.components
			c.graphMut.RUnlock()

			runnables := make([]controller.RunnableNode, 0, len(components))
			for _, uc := range components {
				runnables = append(runnables, uc)
			}
			c.sched.Synchronize(runnables)
		}
	}
}

func (c *Controller) handleUpdatedComponent(uc *controller.ComponentNode) {
	// handleUpdatedComponent is called when uc's exports get updated.
	//
	// NOTE(rfratto): we call StoreComponent here as an optimization since the
	// OnStateChange callback may get invoked many times before we're ready for
	// processing it. Waiting to call StoreComponent allows us to minimize the
	// amount of times we need to convert its Exports to a cty.Value.
	c.cache.CacheExports(uc.ID(), uc.Exports())

	c.graphMut.RLock()
	defer c.graphMut.RUnlock()

	// Walk through all the dependants of uc and re-evaluate their inputs.
	_ = dag.WalkReverse(c.graph, []dag.Node{uc}, func(dep dag.Node) error {
		depComponent := dep.(*controller.ComponentNode)

		ectx := c.cache.BuildContext(rootEvalContext)
		if err := depComponent.Evaluate(ectx); err != nil {
			level.Warn(c.log).Log("msg", "failed to reevaluate component", "node_id", dep.NodeID(), "err", err)
			return nil
		}
		if c.onComponentChanged != nil {
			c.onComponentChanged(uc)
		}

		// Update the cache for our component since its config (probably) just
		// changed.
		c.cache.CacheArguments(uc.ID(), uc.Exports())
		return nil
	})
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
	c.graphMut.Lock()
	defer c.graphMut.Unlock()

	// NOTE(rfratto): to simplify loading logic, we create a new graph seeded
	// from the existing graph. Whomever implements config partials should be
	// cautious whether this approach prevents partials from working and refactor
	// this code as needed.

	componentOpts := controller.ComponentOptions{
		Logger:   c.log,
		DataPath: c.opts.DataPath,
		OnExportsChange: func(uc *controller.ComponentNode) {
			c.updateQueue.Enqueue(uc)
			if c.onComponentChanged != nil {
				c.onComponentChanged(uc)
			}
		},
	}
	newGraph, diags := buildGraph(componentOpts, c.graph, f.Components)

	// Perform a transitive reduction to clean up the graph.
	dag.Reduce(newGraph)

	// TODO(rfratto): detect cycles in the graph and add ignore any components
	// which are currently part of a cycle.

	// Our graph is now fully initialized and we can start doing a topological
	// walk to evaluate components.
	//
	// While walking the graph, we store components we come across in
	// allComponents, regardless of whether they evaluated properly or not. This
	// is then passed to our run() goroutine which will decide the subset of
	// components which can be run.
	var (
		allComponents []*controller.ComponentNode
		componentIDs  []controller.ComponentID
	)

	_ = dag.WalkTopological(newGraph, newGraph.Leaves(), func(n dag.Node) error {
		uc := n.(*controller.ComponentNode)
		allComponents = append(allComponents, uc)
		componentIDs = append(componentIDs, uc.ID())

		// If this node wasn't previously ignored we can try to evaluate it. It
		// should be added to the ignored list if the evaluation failed.
		ectx := c.cache.BuildContext(rootEvalContext)
		if err := uc.Evaluate(ectx); err != nil {
			return nil
		}
		if c.onComponentChanged != nil {
			c.onComponentChanged(uc)
		}

		// Update our cache with our evaluated component. We don't update the
		// Exports because the component isn't running yet.
		c.cache.CacheArguments(uc.ID(), uc.Arguments())
		return nil
	})

	// Store our new graph and synchronize our cache to remove any components
	// which have been removed.
	c.graph = newGraph
	c.cache.SyncIDs(componentIDs)

	if !c.loadedOnce && diags.HasErrors() {
		// The first call to Load should not run any components if there were
		// errors in the coniguration file.
		return diags
	}
	c.loadedOnce = true
	c.components = allComponents

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
