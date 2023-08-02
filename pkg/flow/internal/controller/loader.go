package controller

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/grafana/agent/component"
	"github.com/grafana/agent/pkg/flow/internal/dag"
	"github.com/grafana/agent/pkg/flow/tracing"
	"github.com/grafana/agent/pkg/river/ast"
	"github.com/grafana/agent/pkg/river/diag"
	"github.com/grafana/agent/service"
	"github.com/grafana/ckit"
	"github.com/grafana/ckit/peer"
	"github.com/hashicorp/go-multierror"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	_ "github.com/grafana/agent/pkg/flow/internal/testcomponents" // Include test components
)

// The Loader builds and evaluates ComponentNodes from River blocks.
type Loader struct {
	log      log.Logger
	tracer   trace.TracerProvider
	globals  ComponentGlobals
	services []service.Service
	host     service.Host

	mut               sync.RWMutex
	graph             *dag.Graph
	originalGraph     *dag.Graph
	componentNodes    []*ComponentNode
	serviceNodes      []*ServiceNode
	cache             *valueCache
	blocks            []*ast.BlockStmt // Most recently loaded blocks, used for writing
	cm                *controllerMetrics
	cc                *controllerCollector
	moduleExportIndex int
}

// LoaderOptions holds options for creating a Loader.
type LoaderOptions struct {
	// ComponentGlobals contains data to use when creating components.
	ComponentGlobals ComponentGlobals

	Services []service.Service // Services to load into the DAG.
	Host     service.Host      // Service host (when running services).
}

// NewLoader creates a new Loader. Components built by the Loader will be built
// with co for their options.
func NewLoader(opts LoaderOptions) *Loader {
	var (
		globals  = opts.ComponentGlobals
		services = opts.Services
		host     = opts.Host
	)

	l := &Loader{
		log:      log.With(globals.Logger, "controller_id", globals.ControllerID),
		tracer:   tracing.WrapTracerForLoader(globals.TraceProvider, globals.ControllerID),
		globals:  globals,
		services: services,
		host:     host,

		graph:         &dag.Graph{},
		originalGraph: &dag.Graph{},
		cache:         newValueCache(),
		cm:            newControllerMetrics(globals.ControllerID),
	}
	l.cc = newControllerCollector(l, globals.ControllerID)

	if globals.Registerer != nil {
		globals.Registerer.MustRegister(l.cc)
		globals.Registerer.MustRegister(l.cm)
	}

	globals.Clusterer.Node.Observe(ckit.FuncObserver(func(peers []peer.Peer) (reregister bool) {
		tracer := l.tracer.Tracer("")
		spanCtx, span := tracer.Start(context.Background(), "ClusterStateChange", trace.WithSpanKind(trace.SpanKindInternal))
		defer span.End()
		for _, cmp := range l.Components() {
			if cc, ok := cmp.managed.(component.ClusteredComponent); ok {
				if cc.ClusterUpdatesRegistration() {
					_, span := tracer.Start(spanCtx, "ClusteredComponentReevaluation", trace.WithSpanKind(trace.SpanKindInternal))
					span.SetAttributes(attribute.String("node_id", cmp.NodeID()))
					defer span.End()

					err := cmp.Reevaluate()
					if err != nil {
						level.Error(l.log).Log("msg", "failed to reevaluate component", "componentID", cmp.NodeID(), "err", err)
					}
				}
			}
		}
		return true
	}))

	return l
}

// Apply loads a new set of components into the Loader. Apply will drop any
// previously loaded component which is not described in the set of River
// blocks.
//
// Apply will reuse existing components if there is an existing component which
// matches the component ID specified by any of the provided River blocks.
// Reused components will be updated to point at the new River block.
//
// Apply will perform an evaluation of all loaded components before returning.
// The provided parentContext can be used to provide global variables and
// functions to components. A child context will be constructed from the parent
// to expose values of other components.
func (l *Loader) Apply(args map[string]any, componentBlocks []*ast.BlockStmt, configBlocks []*ast.BlockStmt) diag.Diagnostics {
	start := time.Now()
	l.mut.Lock()
	defer l.mut.Unlock()
	l.cm.controllerEvaluation.Set(1)
	defer l.cm.controllerEvaluation.Set(0)

	for key, value := range args {
		l.cache.CacheModuleArgument(key, value)
	}
	l.cache.SyncModuleArgs(args)

	newGraph, diags := l.loadNewGraph(args, componentBlocks, configBlocks)
	if diags.HasErrors() {
		return diags
	}

	var (
		components   = make([]*ComponentNode, 0, len(componentBlocks))
		componentIDs = make([]ComponentID, 0, len(componentBlocks))
		services     = make([]*ServiceNode, 0, len(l.services))
	)

	tracer := l.tracer.Tracer("")
	spanCtx, span := tracer.Start(context.Background(), "GraphEvaluate", trace.WithSpanKind(trace.SpanKindInternal))
	defer span.End()

	logger := log.With(l.log, "trace_id", span.SpanContext().TraceID())
	level.Info(logger).Log("msg", "starting complete graph evaluation")
	defer func() {
		span.SetStatus(codes.Ok, "")

		duration := time.Since(start)
		level.Info(logger).Log("msg", "finished complete graph evaluation", "duration", duration)
		l.cm.componentEvaluationTime.Observe(duration.Seconds())
	}()

	l.cache.ClearModuleExports()

	// Evaluate all the components.
	_ = dag.WalkTopological(&newGraph, newGraph.Leaves(), func(n dag.Node) error {
		_, span := tracer.Start(spanCtx, "EvaluateNode", trace.WithSpanKind(trace.SpanKindInternal))
		span.SetAttributes(attribute.String("node_id", n.NodeID()))
		defer span.End()

		start := time.Now()
		defer func() {
			level.Info(logger).Log("msg", "finished node evaluation", "node_id", n.NodeID(), "duration", time.Since(start))
		}()

		var err error

		switch n := n.(type) {
		case *ComponentNode:
			components = append(components, n)
			componentIDs = append(componentIDs, n.ID())

			if err = l.evaluate(logger, n); err != nil {
				var evalDiags diag.Diagnostics
				if errors.As(err, &evalDiags) {
					diags = append(diags, evalDiags...)
				} else {
					diags.Add(diag.Diagnostic{
						Severity: diag.SeverityLevelError,
						Message:  fmt.Sprintf("Failed to build component: %s", err),
						StartPos: ast.StartPos(n.Block()).Position(),
						EndPos:   ast.EndPos(n.Block()).Position(),
					})
				}
			}

		case *ServiceNode:
			services = append(services, n)

			if err = l.evaluate(logger, n); err != nil {
				var evalDiags diag.Diagnostics
				if errors.As(err, &evalDiags) {
					diags = append(diags, evalDiags...)
				} else {
					diags.Add(diag.Diagnostic{
						Severity: diag.SeverityLevelError,
						Message:  fmt.Sprintf("Failed to evaluate service: %s", err),
						StartPos: ast.StartPos(n.Block()).Position(),
						EndPos:   ast.EndPos(n.Block()).Position(),
					})
				}
			}

		case BlockNode:
			if err = l.evaluate(logger, n); err != nil {
				diags.Add(diag.Diagnostic{
					Severity: diag.SeverityLevelError,
					Message:  fmt.Sprintf("Failed to evaluate node for config block: %s", err),
					StartPos: ast.StartPos(n.Block()).Position(),
					EndPos:   ast.EndPos(n.Block()).Position(),
				})
			}
			if exp, ok := n.(*ExportConfigNode); ok {
				l.cache.CacheModuleExportValue(exp.Label(), exp.Value())
			}
		}

		// We only use the error for updating the span status; we don't return the
		// error because we want to evaluate as many nodes as we can.
		if err != nil {
			span.SetStatus(codes.Error, err.Error())
		} else {
			span.SetStatus(codes.Ok, "")
		}
		return nil
	})

	l.componentNodes = components
	l.serviceNodes = services
	l.graph = &newGraph
	l.cache.SyncIDs(componentIDs)
	l.blocks = componentBlocks
	l.cm.componentEvaluationTime.Observe(time.Since(start).Seconds())
	if l.globals.OnExportsChange != nil && l.cache.ExportChangeIndex() != l.moduleExportIndex {
		l.moduleExportIndex = l.cache.ExportChangeIndex()
		l.globals.OnExportsChange(l.cache.CreateModuleExports())
	}
	return diags
}

// Cleanup unregisters any existing metrics.
func (l *Loader) Cleanup() {
	if l.globals.Registerer == nil {
		return
	}
	l.globals.Registerer.Unregister(l.cm)
	l.globals.Registerer.Unregister(l.cc)
}

// loadNewGraph creates a new graph from the provided blocks and validates it.
func (l *Loader) loadNewGraph(args map[string]any, componentBlocks []*ast.BlockStmt, configBlocks []*ast.BlockStmt) (dag.Graph, diag.Diagnostics) {
	var g dag.Graph

	// Split component blocks into blocks for components and services.
	componentBlocks, serviceBlocks := l.splitComponentBlocks(componentBlocks)

	// Fill our graph with service blocks, which must be added before any other
	// block.
	diags := l.populateServiceNodes(&g, serviceBlocks)

	// Fill our graph with config blocks.
	configBlockDiags := l.populateConfigBlockNodes(args, &g, configBlocks)
	diags = append(diags, configBlockDiags...)

	// Fill our graph with components.
	componentNodeDiags := l.populateComponentNodes(&g, componentBlocks)
	diags = append(diags, componentNodeDiags...)

	// Write up the edges of the graph
	wireDiags := l.wireGraphEdges(&g)
	diags = append(diags, wireDiags...)

	// Validate graph to detect cycles
	err := dag.Validate(&g)
	if err != nil {
		diags = append(diags, multierrToDiags(err)...)
		return g, diags
	}

	// Copy the original graph, this is so we can have access to the original graph for things like displaying a UI or
	// debug information.
	l.originalGraph = g.Clone()
	// Perform a transitive reduction of the graph to clean it up.
	dag.Reduce(&g)

	return g, diags
}

func (l *Loader) splitComponentBlocks(blocks []*ast.BlockStmt) (componentBlocks, serviceBlocks []*ast.BlockStmt) {
	componentBlocks = make([]*ast.BlockStmt, 0, len(blocks))
	serviceBlocks = make([]*ast.BlockStmt, 0, len(l.services))

	serviceNames := make(map[string]struct{}, len(l.services))
	for _, svc := range l.services {
		serviceNames[svc.Definition().Name] = struct{}{}
	}

	for _, block := range blocks {
		if _, isService := serviceNames[BlockComponentID(block).String()]; isService {
			serviceBlocks = append(serviceBlocks, block)
		} else {
			componentBlocks = append(componentBlocks, block)
		}
	}

	return componentBlocks, serviceBlocks
}

// populateServiceNodes adds service nodes to the graph.
func (l *Loader) populateServiceNodes(g *dag.Graph, serviceBlocks []*ast.BlockStmt) diag.Diagnostics {
	var diags diag.Diagnostics

	// First, build the services.
	for _, svc := range l.services {
		id := svc.Definition().Name

		if g.GetByID(id) != nil {
			diags.Add(diag.Diagnostic{
				Severity: diag.SeverityLevelError,
				Message:  fmt.Sprintf("cannot add service %q; node with same ID already exists", id),
			})

			continue
		}

		var node *ServiceNode

		// Check the graph from the previous call to Load to see we can copy an
		// existing instance of ServiceNode.
		if exist := l.graph.GetByID(id); exist != nil {
			node = exist.(*ServiceNode)
		} else {
			node = NewServiceNode(l.host, svc)
		}

		node.UpdateBlock(nil) // Reset configuration to nil.
		g.Add(node)
	}

	// Now, assign blocks to services.
	for _, block := range serviceBlocks {
		blockID := BlockComponentID(block).String()
		node := g.GetByID(blockID).(*ServiceNode)

		// Blocks assigned to services are reset to nil in the previous loop.
		//
		// If the block is non-nil, it means that there was a duplicate block
		// configuring the same service found in a previous iteration of this loop.
		if node.Block() != nil {
			diags.Add(diag.Diagnostic{
				Severity: diag.SeverityLevelError,
				Message:  fmt.Sprintf("duplicate definition of %q", blockID),
				StartPos: ast.StartPos(block).Position(),
				EndPos:   ast.EndPos(block).Position(),
			})
			continue
		}

		node.UpdateBlock(block)
	}

	return diags
}

// populateConfigBlockNodes adds any config blocks to the graph.
func (l *Loader) populateConfigBlockNodes(args map[string]any, g *dag.Graph, configBlocks []*ast.BlockStmt) diag.Diagnostics {
	var (
		diags   diag.Diagnostics
		nodeMap = NewConfigNodeMap()
	)

	for _, block := range configBlocks {
		node, newConfigNodeDiags := NewConfigNode(block, l.globals)
		diags = append(diags, newConfigNodeDiags...)

		if g.GetByID(node.NodeID()) != nil {
			diags.Add(diag.Diagnostic{
				Severity: diag.SeverityLevelError,
				Message:  fmt.Sprintf("%q block already declared", node.NodeID()),
				StartPos: ast.StartPos(block).Position(),
				EndPos:   ast.EndPos(block).Position(),
			})

			continue
		}

		nodeMapDiags := nodeMap.Append(node)
		diags = append(diags, nodeMapDiags...)
		if diags.HasErrors() {
			continue
		}

		g.Add(node)
	}

	validateDiags := nodeMap.Validate(l.isModule(), args)
	diags = append(diags, validateDiags...)

	// If a logging config block is not provided, we create an empty node which uses defaults.
	if nodeMap.logging == nil && !l.isModule() {
		c := NewDefaultLoggingConfigNode(l.globals)
		g.Add(c)
	}

	// If a tracing config block is not provided, we create an empty node which uses defaults.
	if nodeMap.tracing == nil && !l.isModule() {
		c := NewDefaulTracingConfigNode(l.globals)
		g.Add(c)
	}

	return diags
}

// populateComponentNodes adds any components to the graph.
func (l *Loader) populateComponentNodes(g *dag.Graph, componentBlocks []*ast.BlockStmt) diag.Diagnostics {
	var (
		diags    diag.Diagnostics
		blockMap = make(map[string]*ast.BlockStmt, len(componentBlocks))
	)
	for _, block := range componentBlocks {
		var c *ComponentNode
		id := BlockComponentID(block).String()

		if orig, redefined := blockMap[id]; redefined {
			diags.Add(diag.Diagnostic{
				Severity: diag.SeverityLevelError,
				Message:  fmt.Sprintf("Component %s already declared at %s", id, ast.StartPos(orig).Position()),
				StartPos: block.NamePos.Position(),
				EndPos:   block.NamePos.Add(len(id) - 1).Position(),
			})
			continue
		}
		blockMap[id] = block

		// Check the graph from the previous call to Load to see we can copy an
		// existing instance of ComponentNode.
		if exist := l.graph.GetByID(id); exist != nil {
			c = exist.(*ComponentNode)
			c.UpdateBlock(block)
		} else {
			componentName := block.GetBlockName()
			registration, exists := component.Get(componentName)
			if !exists {
				diags.Add(diag.Diagnostic{
					Severity: diag.SeverityLevelError,
					Message:  fmt.Sprintf("Unrecognized component name %q", componentName),
					StartPos: block.NamePos.Position(),
					EndPos:   block.NamePos.Add(len(componentName) - 1).Position(),
				})
				continue
			}

			if registration.Singleton && block.Label != "" {
				diags.Add(diag.Diagnostic{
					Severity: diag.SeverityLevelError,
					Message:  fmt.Sprintf("Component %q does not support labels", componentName),
					StartPos: block.LabelPos.Position(),
					EndPos:   block.LabelPos.Add(len(block.Label) + 1).Position(),
				})
				continue
			}

			if !registration.Singleton && block.Label == "" {
				diags.Add(diag.Diagnostic{
					Severity: diag.SeverityLevelError,
					Message:  fmt.Sprintf("Component %q must have a label", componentName),
					StartPos: block.NamePos.Position(),
					EndPos:   block.NamePos.Add(len(componentName) - 1).Position(),
				})
				continue
			}

			if registration.Singleton && l.isModule() {
				diags.Add(diag.Diagnostic{
					Severity: diag.SeverityLevelError,
					Message:  fmt.Sprintf("Component %q is a singleton and unsupported inside a module", componentName),
					StartPos: block.NamePos.Position(),
					EndPos:   block.NamePos.Add(len(componentName) - 1).Position(),
				})
				continue
			}

			// Create a new component
			c = NewComponentNode(l.globals, block)
		}

		g.Add(c)
	}

	return diags
}

// Wire up all the related nodes
func (l *Loader) wireGraphEdges(g *dag.Graph) diag.Diagnostics {
	var diags diag.Diagnostics

	for _, n := range g.Nodes() {
		// First, wire up dependencies on services.
		switch n := n.(type) {
		case *ServiceNode: // Service depending on other services.
			for _, depName := range n.Definition().DependsOn {
				dep := g.GetByID(depName)
				if dep == nil {
					diags.Add(diag.Diagnostic{
						Severity: diag.SeverityLevelError,
						Message:  fmt.Sprintf("service %q has invalid reference to service %q", n.NodeID(), depName),
					})
					continue
				}

				g.AddEdge(dag.Edge{From: n, To: dep})
			}

		case *ComponentNode: // Component depending on service.
			for _, depName := range n.Registration().NeedsServices {
				dep := g.GetByID(depName)
				if dep == nil {
					diags.Add(diag.Diagnostic{
						Severity: diag.SeverityLevelError,
						Message:  fmt.Sprintf("component depends on undefined service %q; please report this issue to project maintainers", depName),
						StartPos: ast.StartPos(n.Block()).Position(),
						EndPos:   ast.EndPos(n.Block()).Position(),
					})
					continue
				}

				g.AddEdge(dag.Edge{From: n, To: dep})
			}
		}

		// Finally, wire component references.
		refs, nodeDiags := ComponentReferences(n, g)
		for _, ref := range refs {
			g.AddEdge(dag.Edge{From: n, To: ref.Target})
		}
		diags = append(diags, nodeDiags...)
	}

	return diags
}

// Variables returns the Variables the Loader exposes for other Flow components
// to reference.
func (l *Loader) Variables() map[string]interface{} {
	return l.cache.BuildContext().Variables
}

// Components returns the current set of loaded components.
func (l *Loader) Components() []*ComponentNode {
	l.mut.RLock()
	defer l.mut.RUnlock()
	return l.componentNodes
}

// Services returns the current set of service nodes.
func (l *Loader) Services() []*ServiceNode {
	l.mut.RLock()
	defer l.mut.RUnlock()
	return l.serviceNodes
}

// Graph returns a copy of the DAG managed by the Loader.
func (l *Loader) Graph() *dag.Graph {
	l.mut.RLock()
	defer l.mut.RUnlock()
	return l.graph.Clone()
}

// OriginalGraph returns a copy of the graph before Reduce was called. This can be used if you want to show a UI of the
// original graph before the reduce function was called.
func (l *Loader) OriginalGraph() *dag.Graph {
	l.mut.RLock()
	defer l.mut.RUnlock()
	return l.originalGraph.Clone()
}

// EvaluateDependencies re-evaluates components which depend directly or
// indirectly on c. EvaluateDependencies should be called whenever a component
// updates its exports.
//
// The provided parentContext can be used to provide global variables and
// functions to components. A child context will be constructed from the parent
// to expose values of other components.
func (l *Loader) EvaluateDependencies(c *ComponentNode) {
	tracer := l.tracer.Tracer("")

	l.mut.RLock()
	defer l.mut.RUnlock()

	l.cm.controllerEvaluation.Set(1)
	defer l.cm.controllerEvaluation.Set(0)
	start := time.Now()

	spanCtx, span := tracer.Start(context.Background(), "GraphEvaluatePartial", trace.WithSpanKind(trace.SpanKindInternal))
	span.SetAttributes(attribute.String("initiator", c.NodeID()))
	defer span.End()

	logger := log.With(l.log, "trace_id", span.SpanContext().TraceID())
	level.Info(logger).Log("msg", "starting partial graph evaluation")
	defer func() {
		span.SetStatus(codes.Ok, "")

		duration := time.Since(start)
		level.Info(logger).Log("msg", "finished partial graph evaluation", "duration", duration)
		l.cm.componentEvaluationTime.Observe(duration.Seconds())
	}()

	// Make sure we're in-sync with the current exports of c.
	l.cache.CacheExports(c.ID(), c.Exports())

	_ = dag.WalkReverse(l.graph, []dag.Node{c}, func(n dag.Node) error {
		if n == c {
			// Skip over the starting component; the starting component passed to
			// EvaluateDependencies had its exports changed and none of its input
			// arguments will need re-evaluation.
			return nil
		}

		_, span := tracer.Start(spanCtx, "EvaluateNode", trace.WithSpanKind(trace.SpanKindInternal))
		span.SetAttributes(attribute.String("node_id", n.NodeID()))
		defer span.End()

		start := time.Now()
		defer func() {
			level.Info(logger).Log("msg", "finished node evaluation", "node_id", n.NodeID(), "duration", time.Since(start))
		}()

		var err error

		switch n := n.(type) {
		case BlockNode:
			err = l.evaluate(logger, n)
			if exp, ok := n.(*ExportConfigNode); ok {
				l.cache.CacheModuleExportValue(exp.Label(), exp.Value())
			}
		}

		// We only use the error for updating the span status; we don't return the
		// error because we want to evaluate as many nodes as we can.
		if err != nil {
			span.SetStatus(codes.Error, err.Error())
		} else {
			span.SetStatus(codes.Ok, "")
		}
		return nil
	})

	if l.globals.OnExportsChange != nil && l.cache.ExportChangeIndex() != l.moduleExportIndex {
		l.globals.OnExportsChange(l.cache.CreateModuleExports())
		l.moduleExportIndex = l.cache.ExportChangeIndex()
	}
}

// evaluate constructs the final context for the BlockNode and
// evaluates it. mut must be held when calling evaluate.
func (l *Loader) evaluate(logger log.Logger, bn BlockNode) error {
	ectx := l.cache.BuildContext()
	err := bn.Evaluate(ectx)

	switch c := bn.(type) {
	case *ComponentNode:
		// Always update the cache both the arguments and exports, since both might
		// change when a component gets re-evaluated. We also want to cache the arguments and exports in case of an error
		l.cache.CacheArguments(c.ID(), c.Arguments())
		l.cache.CacheExports(c.ID(), c.Exports())
	case *ArgumentConfigNode:
		if _, found := l.cache.moduleArguments[c.Label()]; !found {
			if c.Optional() {
				l.cache.CacheModuleArgument(c.Label(), c.Default())
			} else {
				err = fmt.Errorf("missing required argument %q to module", c.Label())
			}
		}
	}

	if err != nil {
		level.Error(logger).Log("msg", "failed to evaluate config", "node", bn.NodeID(), "err", err)
		return err
	}
	return nil
}

func multierrToDiags(errors error) diag.Diagnostics {
	var diags diag.Diagnostics
	for _, err := range errors.(*multierror.Error).Errors {
		// TODO(rfratto): should this include position information?
		diags.Add(diag.Diagnostic{
			Severity: diag.SeverityLevelError,
			Message:  err.Error(),
		})
	}
	return diags
}

// If the definition of a module ever changes, update this.
func (l *Loader) isModule() bool {
	// Either 1 of these checks is technically sufficient but let's be extra careful.
	return l.globals.OnExportsChange != nil && l.globals.ControllerID != ""
}
