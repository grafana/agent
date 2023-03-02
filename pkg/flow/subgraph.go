package flow

import (
	"context"
	"fmt"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"go.opentelemetry.io/otel/trace"

	"github.com/go-kit/log/level"
	"github.com/grafana/agent/component"
	"github.com/grafana/agent/pkg/flow/internal/controller"
	"github.com/grafana/agent/pkg/flow/logging"
	"github.com/grafana/agent/pkg/river/diag"
	"github.com/grafana/agent/pkg/river/vm"
	"github.com/hashicorp/go-multierror"
)

// subgraph represents a namespace in the flow subsystem. The namespace is defined by the calling
// owner. In the case of normal operation that is the flow component with a namespace of "".
// This is defined as the main module, with `module.*` components loading submodules.
// A parent owner has a unique id that will propagate to child components to create a unique
// namespace id that is parent.id + component.id.
type subgraph struct {
	mut          sync.RWMutex
	log          *logging.Logger
	parent       component.SubgraphOwner
	parentGraph  *subgraph
	scope        *vm.Scope
	loader       *controller.Loader
	updateQueue  *controller.Queue
	sched        *controller.Scheduler
	globals      controller.ComponentGlobals
	children     map[component.SubgraphOwner]child
	exited       chan struct{}
	loadFinished chan struct{}
	cancel       func()
	cm           *controller.ControllerMetrics
}

// child represents a child subgraph with the created context and cancel func.
// This is used when we need to stop the child for shutting down or reloading.
type child struct {
	graph  *subgraph
	ctx    context.Context
	cancel context.CancelFunc
}

// newSubgraph creates a subgraph for use with modules, either main or child.
func newSubgraph(
	parent component.SubgraphOwner,
	parentGraph *subgraph,
	log *logging.Logger,
	trace trace.TracerProvider,
	datapath string,
	register prometheus.Registerer,
	httplistener string,
	cm *controller.ControllerMetrics,
) *subgraph {

	sg := &subgraph{
		parent:       parent,
		sched:        controller.NewScheduler(),
		updateQueue:  controller.NewQueue(),
		children:     make(map[component.SubgraphOwner]child),
		parentGraph:  parentGraph,
		exited:       make(chan struct{}, 1),
		loadFinished: make(chan struct{}, 1),
		log:          log,
		cm:           cm,
	}
	globals := controller.ComponentGlobals{
		Logger:        log,
		TraceProvider: controller.WrapTracer(trace, parent.ID()),
		DataPath:      datapath,
		OnExportsChange: func(cn *controller.ComponentNode) {
			// Changed components should be queued for reevaluation.
			sg.updateQueue.Enqueue(cn)
		},
		Registerer:     register,
		HTTPListenAddr: httplistener,
	}
	sg.globals = globals
	sg.loader = controller.NewLoader(globals, cm)
	return sg
}

// loadInitialSubgraph is a special case used for the main subpgraph, instead of using the delegate we
// force passing in flow so only it can call this.
func (s *subgraph) loadInitialSubgraph(flow *Flow, config []byte, filename string) ([]component.Component, diag.Diagnostics, error) {
	file, err := ReadFile(filename, config)
	if err != nil {
		return nil, nil, err
	}
	diags, comps := s.loader.Apply(s, flow, s.scope, file.Components, file.ConfigBlocks)
	if diags.HasErrors() {
		return nil, diags, diags.ErrorOrNil()
	}
	ctx, cancel := context.WithCancel(context.Background())
	s.cancel = cancel
	go s.run(ctx)
	return comps, diags, nil
}

// LoadSubgraph allows a component to load a subgraph with the calling component as the parent.
// This returns all the components both new and old. EXTREME care should be taken by the caller
// since the component is shared.
func (s *subgraph) LoadSubgraph(parent component.SubgraphOwner, config []byte) ([]component.Component, diag.Diagnostics, error) {

	// Check to see if there is already a graph loaded
	foundsg, found := s.getChild(parent)

	if found {
		// TODO figure out if we want to apply the old one back, we probably do
		err := foundsg.graph.close()
		if err != nil {

			return nil, nil, err
		}
		delete(s.children, parent)
	}

	file, err := ReadFile(parent.ID(), config)
	if err != nil {
		return nil, nil, err
	}
	sg := newSubgraph(parent, s, s.log, s.globals.TraceProvider, s.globals.DataPath, s.globals.Registerer, s.globals.HTTPListenAddr, s.cm)
	diags, comps := sg.loader.Apply(s, parent, s.scope, file.Components, file.ConfigBlocks)
	if diags.HasErrors() {
		return nil, diags, diags.ErrorOrNil()
	}
	ctx, cancel := context.WithCancel(context.Background())

	s.addChild(parent, child{
		graph:  sg,
		ctx:    ctx,
		cancel: cancel,
	})

	return comps, diags, nil
}

// UnloadSubgraph is used when you no longer need to load the graph
func (s *subgraph) UnloadSubgraph(parent component.SubgraphOwner) error {
	foundsg, found := s.getChild(parent)
	if !found {
		return fmt.Errorf("unable to find subgraph with parent id %s", parent.ID())
	}

	// TODO figure out if we want to apply the old one back, we probably do
	err := foundsg.graph.close()
	if err != nil {
		return err
	}
	s.deleteChild(parent)
	return nil
}

func (s *subgraph) StartSubgraph(parent component.SubgraphOwner) error {
	foundsg, found := s.getChild(parent)
	if !found {
		return fmt.Errorf("unable to find subgraph with parent id %s", parent.ID())
	}
	go foundsg.graph.run(foundsg.ctx)
	return nil
}

func (s *subgraph) Components() []*controller.ComponentNode {
	s.mut.RLock()
	defer s.mut.RUnlock()

	comps := make([]*controller.ComponentNode, 0)
	comps = append(comps, s.loader.Components()...)
	for _, x := range s.children {
		comps = append(comps, x.graph.Components()...)
	}
	return comps
}

func (s *subgraph) getChild(parent component.SubgraphOwner) (child, bool) {
	s.mut.RLock()
	defer s.mut.RUnlock()

	ch, found := s.children[parent]
	return ch, found
}

func (s *subgraph) addChild(parent component.SubgraphOwner, ch child) {
	s.mut.Lock()
	defer s.mut.Unlock()

	s.children[parent] = ch
}

func (s *subgraph) deleteChild(parent component.SubgraphOwner) {
	s.mut.Lock()
	defer s.mut.Unlock()

	delete(s.children, parent)
}

// close recursively closes all children
func (s *subgraph) close() error {
	s.mut.Lock()
	defer s.mut.Unlock()

	var result error
	for _, x := range s.children {
		err := x.graph.close()
		if err != nil {
			result = multierror.Append(result, err)
		}
	}
	s.cancel()
	<-s.exited
	err := s.sched.Close()
	if err != nil {
		result = multierror.Append(result, err)
	}
	return result
}

func (s *subgraph) run(ctx context.Context) {
	defer level.Debug(s.log).Log("msg", "subgraph exiting")
	defer close(s.exited)

	ctx, cancel := context.WithCancel(ctx)
	s.cancel = cancel
	s.loadFinished <- struct{}{}

	for {
		select {
		case <-ctx.Done():
			return

		case <-s.updateQueue.Chan():
			// We need to pop _everything_ from the queue and evaluate each of them.
			// If we only pop a single element, other components may sit waiting for
			// evaluation forever.
			for {
				updated := s.updateQueue.TryDequeue()
				if updated == nil {
					break
				}

				level.Debug(s.log).Log("msg", "handling component with updated state", "node_id", updated.NodeID())
				s.loader.EvaluateDependencies(nil, updated)
			}

		case <-s.loadFinished:
			level.Info(s.log).Log("msg", "scheduling loaded components")

			components := s.loader.Components()
			runnables := make([]controller.RunnableNode, 0, len(components))
			for _, uc := range components {
				runnables = append(runnables, uc)
			}
			err := s.sched.Synchronize(runnables)
			if err != nil {
				level.Error(s.log).Log("msg", "failed to load components", "err", err)
			}
		}
	}
}
