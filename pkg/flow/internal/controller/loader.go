package controller

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/grafana/agent/component"
	"github.com/grafana/agent/pkg/flow/internal/dag"
	"github.com/grafana/agent/pkg/river/ast"
	"github.com/grafana/agent/pkg/river/diag"
	"github.com/grafana/agent/pkg/river/vm"
	"github.com/hashicorp/go-multierror"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	_ "github.com/grafana/agent/pkg/flow/internal/testcomponents" // Include test components
)

// The Loader builds and evaluates ComponentNodes from River blocks.
type Loader struct {
	log             log.Logger
	tracer          trace.TracerProvider
	globals         ComponentGlobals
	onExportsChange func(map[string]any)

	mut           sync.RWMutex
	graph         *dag.Graph
	originalGraph *dag.Graph
	components    []*ComponentNode
	cache         *valueCache
	blocks        []*ast.BlockStmt // Most recently loaded blocks, used for writing
	cm            *controllerMetrics
}

// NewLoader creates a new Loader. Components built by the Loader will be built
// with co for their options.
func NewLoader(globals ComponentGlobals) *Loader {
	l := &Loader{
		log:             globals.Logger,
		tracer:          globals.TraceProvider,
		globals:         globals,
		onExportsChange: globals.OnExportsChange,

		graph:         &dag.Graph{},
		originalGraph: &dag.Graph{},
		cache:         newValueCache(),
		cm:            newControllerMetrics(globals.Registerer),
	}
	cc := newControllerCollector(l)
	if globals.Registerer != nil {
		globals.Registerer.MustRegister(cc)
	}
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
func (l *Loader) Apply(parentScope *vm.Scope, blocks []*ast.BlockStmt, configBlocks []*ast.BlockStmt) diag.Diagnostics {
	start := time.Now()
	l.mut.Lock()
	defer l.mut.Unlock()
	l.cm.controllerEvaluation.Set(1)
	defer l.cm.controllerEvaluation.Set(0)

	var (
		diags    diag.Diagnostics
		newGraph dag.Graph
	)

	// Pre-populate graph with a ConfigNode.
	c, configBlockDiags := NewConfigNode(configBlocks, l.log, l.tracer, l.onExportsChange, l.isModule())
	diags = append(diags, configBlockDiags...)
	newGraph.Add(c)

	// Handle the rest of the graph as ComponentNodes.
	populateDiags := l.populateGraph(&newGraph, blocks)
	diags = append(diags, populateDiags...)

	wireDiags := l.wireGraphEdges(parentScope, &newGraph)
	diags = append(diags, wireDiags...)

	// Validate graph to detect cycles
	err := dag.Validate(&newGraph)
	if err != nil {
		diags = append(diags, multierrToDiags(err)...)
		return diags
	}
	// Copy the original graph, this is so we can have access to the original graph for things like displaying a UI or
	// debug information.
	l.originalGraph = newGraph.Clone()
	// Perform a transitive reduction of the graph to clean it up.
	dag.Reduce(&newGraph)

	var (
		components   = make([]*ComponentNode, 0, len(blocks))
		componentIDs = make([]ComponentID, 0, len(blocks))
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

	// Evaluate all of the components.
	_ = dag.WalkTopological(&newGraph, newGraph.Leaves(), func(n dag.Node) error {
		_, span := tracer.Start(spanCtx, "EvaluateNode", trace.WithSpanKind(trace.SpanKindInternal))
		span.SetAttributes(attribute.String("node_id", n.NodeID()))
		defer span.End()

		start := time.Now()
		defer func() {
			level.Info(logger).Log("msg", "finished node evaluation", "node_id", n.NodeID(), "duration", time.Since(start))
		}()

		var err error

		switch c := n.(type) {
		case *ComponentNode:
			components = append(components, c)
			componentIDs = append(componentIDs, c.ID())

			if err = l.evaluate(logger, parentScope, c); err != nil {
				var evalDiags diag.Diagnostics
				if errors.As(err, &evalDiags) {
					diags = append(diags, evalDiags...)
				} else {
					diags.Add(diag.Diagnostic{
						Severity: diag.SeverityLevelError,
						Message:  fmt.Sprintf("Failed to build component: %s", err),
						StartPos: ast.StartPos(n.(*ComponentNode).block).Position(),
						EndPos:   ast.EndPos(n.(*ComponentNode).block).Position(),
					})
				}
			}
		case *ConfigNode:
			var errBlock *ast.BlockStmt
			if errBlock, err = l.evaluateConfig(logger, parentScope, c); err != nil {
				diags.Add(diag.Diagnostic{
					Severity: diag.SeverityLevelError,
					Message:  fmt.Sprintf("Failed to evaluate node for config blocks: %s", err),
					StartPos: ast.StartPos(errBlock).Position(),
					EndPos:   ast.EndPos(errBlock).Position(),
				})
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

	l.components = components
	l.graph = &newGraph
	l.cache.SyncIDs(componentIDs)
	l.blocks = blocks
	l.cm.componentEvaluationTime.Observe(time.Since(start).Seconds())
	return diags
}

func (l *Loader) populateGraph(g *dag.Graph, blocks []*ast.BlockStmt) diag.Diagnostics {
	// Fill our graph with components.
	var (
		diags    diag.Diagnostics
		blockMap = make(map[string]*ast.BlockStmt, len(blocks))
	)
	for _, block := range blocks {
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

		if exist := l.graph.GetByID(id); exist != nil {
			// Re-use the existing component and update its block
			c = exist.(*ComponentNode)
			c.UpdateBlock(block)
		} else {
			componentName := strings.Join(block.Name, ".")
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

func (l *Loader) wireGraphEdges(parent *vm.Scope, g *dag.Graph) diag.Diagnostics {
	var diags diag.Diagnostics

	for _, n := range g.Nodes() {
		refs, nodeDiags := ComponentReferences(parent, n, g)
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
	return l.cache.BuildContext(nil).Variables
}

// Components returns the current set of loaded components.
func (l *Loader) Components() []*ComponentNode {
	l.mut.RLock()
	defer l.mut.RUnlock()
	return l.components
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
func (l *Loader) EvaluateDependencies(parentScope *vm.Scope, c *ComponentNode) {
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

		var err error

		switch n := n.(type) {
		case *ComponentNode:
			err = l.evaluate(logger, parentScope, n)
		case *ConfigNode:
			_, err = l.evaluateConfig(logger, parentScope, n)
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
}

// evaluate constructs the final context for c and evaluates it. mut must be
// held when calling evaluate.
func (l *Loader) evaluate(logger log.Logger, parent *vm.Scope, c *ComponentNode) error {
	ectx := l.cache.BuildContext(parent)
	err := c.Evaluate(ectx)
	// Always update the cache both the arguments and exports, since both might
	// change when a component gets re-evaluated. We also want to cache the arguments and exports in case of an error
	l.cache.CacheArguments(c.ID(), c.Arguments())
	l.cache.CacheExports(c.ID(), c.Exports())
	if err != nil {
		level.Error(logger).Log("msg", "failed to evaluate component", "component", c.NodeID(), "err", err)
		return err
	}
	return nil
}

// evaluateConfig constructs the final context for the special config Node and
// evaluates it. mut must be held when calling evaluateConfig.
func (l *Loader) evaluateConfig(logger log.Logger, parent *vm.Scope, c *ConfigNode) (*ast.BlockStmt, error) {
	ectx := l.cache.BuildContext(parent)
	errBlock, err := c.Evaluate(ectx)
	if err != nil {
		level.Error(logger).Log("msg", "failed to evaluate config", "node", c.NodeID(), "err", err)
		return errBlock, err
	}
	return nil, nil
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
	return l.onExportsChange != nil && l.globals.ControllerID != ""
}
