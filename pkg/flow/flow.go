// Package flow implements the Flow component graph system. Flow configuration
// files are parsed from River, which contain a listing of components to run.
//
// # Components
//
// Each component has a set of arguments (River attributes and blocks) and
// optionally a set of exported fields. Components can reference the exports of
// other components using River expressions.
//
// See the top-level component package for more information on components, and
// subpackages for defined components.
//
// # Component Health
//
// A component will have various health states during its lifetime:
//
//  1. Unknown:   The initial health state for new components.
//  2. Healthy:   A healthy component
//  3. Unhealthy: An unhealthy component.
//  4. Exited:    A component which is no longer running.
//
// Health states are paired with a time for when the health state was generated
// and a message providing more detail for the health state.
//
// Components can report their own health states. The health state reported by
// a component is merged with the Flow-level health of that component: an error
// when evaluating the configuration for a component will always be reported as
// unhealthy until the next successful evaluation.
//
// # Component Evaluation
//
// The process of converting the River block associated with a component into
// the appropriate Go struct is called "component evaluation."
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
	"net"
	"sync"

	"github.com/go-kit/log/level"
	"github.com/grafana/agent/component"
	"github.com/grafana/agent/pkg/cluster"
	"github.com/grafana/agent/pkg/flow/internal/controller"
	"github.com/grafana/agent/pkg/flow/internal/dag"
	"github.com/grafana/agent/pkg/flow/logging"
	"github.com/grafana/agent/pkg/flow/tracing"
	"github.com/prometheus/client_golang/prometheus"
	oteltrace "go.opentelemetry.io/otel/trace"
	"go.uber.org/atomic"
)

// Options holds static options for a flow controller.
type Options struct {
	// ControllerID is an identifier used to represent the controller.
	// ControllerID is used to generate a globally unique display name for
	// components in a binary where multiple controllers are used.
	//
	// If running multiple Flow controllers, each controller must have a
	// different value for ControllerID to be able to differentiate between
	// components in telemetry data.
	ControllerID string

	// Logger to use for controller logs and components. A no-op logger will be
	// created if this is nil.
	Logger *logging.Logger

	// Tracer for components to use. A no-op tracer will be created if this is
	// nil.
	Tracer oteltrace.TracerProvider

	// Clusterer for implementing distributed behavior among components running
	// on different nodes.
	Clusterer *cluster.Clusterer

	// Directory where components can write data. Constructed components will be
	// given a subdirectory of DataPath using the local ID of the component.
	//
	// If running multiple Flow controllers, each controller must have a
	// different value for DataPath to prevent components from colliding.
	DataPath string

	// Reg is the prometheus register to use
	Reg prometheus.Registerer

	// HTTPPathPrefix is the path prefix given to managed components. May be
	// empty. When provided, it should be an absolute path.
	//
	// Components will be given a path relative to HTTPPathPrefix using their
	// local ID.
	//
	// If running multiple Flow controllers, each controller must have a
	// different value for HTTPPathPrefix to prevent components from colliding.
	HTTPPathPrefix string

	// HTTPListenAddr is the base address (host:port) where component APIs are
	// exposed to other components.
	HTTPListenAddr string

	// OnExportsChange is called when the exports of the controller change.
	// Exports are controlled by "export" configuration blocks. If
	// OnExportsChange is nil, export configuration blocks are not allowed in the
	// loaded config file.
	OnExportsChange func(exports map[string]any)

	// DialFunc is a function to use for components to properly connect to
	// HTTPListenAddr. If nil, DialFunc defaults to (&net.Dialer{}).DialContext.
	DialFunc func(ctx context.Context, network, address string) (net.Conn, error)
}

// Flow is the Flow system.
type Flow struct {
	log       *logging.Logger
	tracer    oteltrace.TracerProvider
	clusterer *cluster.Clusterer
	opts      Options

	updateQueue *controller.Queue
	sched       *controller.Scheduler
	loader      *controller.Loader

	loadFinished chan struct{}

	loadMut    sync.RWMutex
	loadedOnce atomic.Bool
}

// New creates and starts a new Flow controller. Call Close to stop
// the controller.
func New(o Options) *Flow {
	var (
		log       = o.Logger
		tracer    = o.Tracer
		clusterer = o.Clusterer
	)

	if tracer == nil {
		var err error
		tracer, err = tracing.New(tracing.DefaultOptions)
		if err != nil {
			// This shouldn't happen unless there's a bug
			panic(err)
		}
	}

	dialFunc := o.DialFunc
	if dialFunc == nil {
		dialFunc = (&net.Dialer{}).DialContext
	}

	var (
		modules = newModuleSystem(&ModuleOptions{
			Logger:    log,
			Tracer:    tracer,
			Clusterer: clusterer,
			Reg:       o.Reg,
		})
		queue  = controller.NewQueue()
		sched  = controller.NewScheduler()
		loader = controller.NewLoader(controller.ComponentGlobals{
			Logger:        log,
			TraceProvider: tracer,
			Clusterer:     clusterer,
			DataPath:      o.DataPath,
			OnComponentUpdate: func(cn *controller.ComponentNode) {
				// Changed components should be queued for reevaluation.
				queue.Enqueue(cn)
			},
			OnExportsChange: o.OnExportsChange,
			Registerer:      o.Reg,
			HTTPPathPrefix:  o.HTTPPathPrefix,
			HTTPListenAddr:  o.HTTPListenAddr,
			DialFunc:        dialFunc,
			ControllerID:    o.ControllerID,
			ModuleSystem:    modules,
		}, o.ControllerID)
	)
	return &Flow{
		log:    log,
		tracer: tracer,
		opts:   o,

		clusterer:   clusterer,
		updateQueue: queue,
		sched:       sched,
		loader:      loader,

		loadFinished: make(chan struct{}, 1),
	}
}

// Run starts the Flow controller, blocking until the provided context is
// canceled. Run must only be called once.
func (c *Flow) Run(ctx context.Context) {
	defer c.sched.Close()
	defer level.Debug(c.log).Log("msg", "flow controller exiting")

	for {
		select {
		case <-ctx.Done():
			return

		case <-c.updateQueue.Chan():
			// We need to pop _everything_ from the queue and evaluate each of them.
			// If we only pop a single element, other components may sit waiting for
			// evaluation forever.
			for {
				updated := c.updateQueue.TryDequeue()
				if updated == nil {
					break
				}

				level.Debug(c.log).Log("msg", "handling component with updated state", "node_id", updated.NodeID())
				c.loader.EvaluateDependencies(updated)
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
func (c *Flow) LoadFile(file *File, args map[string]any) error {
	c.loadMut.Lock()
	defer c.loadMut.Unlock()

	diags := c.loader.Apply(args, file.Components, file.ConfigBlocks)
	if !c.loadedOnce.Load() && diags.HasErrors() {
		// The first call to Load should not run any components if there were
		// errors in the configuration file.
		return diags
	}
	c.loadedOnce.Store(true)

	select {
	case c.loadFinished <- struct{}{}:
	default:
		// A refresh is already scheduled
	}
	return diags.ErrorOrNil()
}

// Ready returns whether the Flow controller has finished its initial load.
func (c *Flow) Ready() bool {
	return c.loadedOnce.Load()
}

// ComponentInfos returns the component infos.
func (c *Flow) ComponentInfos() []*component.ComponentInfo {
	c.loadMut.RLock()
	defer c.loadMut.RUnlock()

	cns := c.loader.Components()
	infos := make([]*component.ComponentInfo, len(cns))
	edges := c.loader.OriginalGraph().Edges()
	for i, com := range cns {
		nn := newFromNode(com, edges)
		infos[i] = nn
	}
	return infos
}

func newFromNode(cn *controller.ComponentNode, edges []dag.Edge) *component.ComponentInfo {
	references := make([]string, 0)
	referencedBy := make([]string, 0)
	for _, e := range edges {
		// Skip over any edge which isn't between two component nodes. This is a
		// temporary workaround needed until there's the concept of configuration
		// blocks from the API.
		//
		// Without this change, the graph fails to render when a configuration
		// block is referenced in the graph.
		//
		// TODO(rfratto): add support for config block nodes in the API and UI.
		if !isComponentNode(e.From) || !isComponentNode(e.To) {
			continue
		}

		if e.From.NodeID() == cn.NodeID() {
			references = append(references, e.To.NodeID())
		} else if e.To.NodeID() == cn.NodeID() {
			referencedBy = append(referencedBy, e.From.NodeID())
		}
	}
	h := cn.CurrentHealth()
	ci := &component.ComponentInfo{
		Label:        cn.Label(),
		ID:           cn.NodeID(),
		Name:         cn.ComponentName(),
		Type:         "block",
		References:   references,
		ReferencedBy: referencedBy,
		Health: &component.ComponentHealth{
			State:       h.Health.String(),
			Message:     h.Message,
			UpdatedTime: h.UpdateTime,
		},
	}
	return ci
}

func isComponentNode(n dag.Node) bool {
	_, ok := n.(*controller.ComponentNode)
	return ok
}
