package flow

import (
	"context"
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
// delegate. In the case of normal operation that is the flow component with a namespace of "".
// A parent delegate has a unique id that will propagate to child components to create a unique
// namespace id that is parent.id + component.id.
type subgraph struct {
	mut          sync.Mutex
	log          *logging.Logger
	parent       component.DelegateComponent
	parentGraph  *subgraph
	scope        *vm.Scope
	loader       *controller.Loader
	updateQueue  *controller.Queue
	sched        *controller.Scheduler
	globals      controller.ComponentGlobals
	children     map[component.DelegateComponent]child
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

// newSubgraph takes in a lot of arguments, we avoid using the
func newSubgraph(
	parent component.DelegateComponent,
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
		children:     make(map[component.DelegateComponent]child),
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

// loadInitialSubgraph is a special case used for the root subpgraph, instead of using the delegate we
// force passing in flow so only it can call this.
func (s *subgraph) loadInitialSubgraph(flow *Flow, config []byte) ([]component.Component, diag.Diagnostics, error) {
	s.mut.Lock()
	defer s.mut.Unlock()

	file, err := ReadFile("", config)
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
func (s *subgraph) LoadSubgraph(parent component.DelegateComponent, config []byte) ([]component.Component, diag.Diagnostics, error) {
	s.mut.Lock()
	defer s.mut.Unlock()

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

	s.children[parent] = child{
		graph:  sg,
		ctx:    ctx,
		cancel: cancel,
	}
	go sg.run(ctx)
	return comps, diags, nil
}

func (s *subgraph) Components() []*controller.ComponentNode {
	// TODO should probably streamline this
	s.mut.Lock()
	defer s.mut.Unlock()
	comps := make([]*controller.ComponentNode, 0)
	comps = append(comps, s.loader.Components()...)
	for _, x := range s.children {
		comps = append(comps, x.graph.Components()...)
	}
	return comps
}

// close recursively closes all children
func (s *subgraph) close() error {
	var result error
	for _, x := range s.children {
		result = multierror.Append(result, x.graph.close())
	}
	s.cancel()
	<-s.exited
	result = multierror.Append(result, s.sched.Close())
	return result
}

func (s *subgraph) run(ctx context.Context) {
	defer close(s.exited)
	defer level.Debug(s.log).Log("msg", "subgraph exiting")

	ctx, cancel := context.WithCancel(ctx)
	s.cancel = cancel
	defer s.cancel()

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
