package controller

import (
	"fmt"
	"sync"

	"github.com/grafana/agent/pkg/flow/tracing"
	"github.com/grafana/agent/pkg/river/ast"
	"github.com/grafana/agent/pkg/river/diag"
	"github.com/grafana/agent/pkg/river/vm"
	"go.opentelemetry.io/otel/trace"
)

type TracingConfigNode struct {
	label         string
	nodeID        string
	componentName string
	traceProvider trace.TracerProvider // Tracer shared between all managed components.

	mut   sync.RWMutex
	block *ast.BlockStmt // Current River blocks to derive config from
	eval  *vm.Evaluator
}

// NewTracingConfigNode creates a new TracingConfigNode from an initial ast.BlockStmt.
// The underlying config isn't applied until Evaluate is called.
func NewTracingConfigNode(block *ast.BlockStmt, globals ComponentGlobals, isInModule bool) (*TracingConfigNode, diag.Diagnostics) {
	var diags diag.Diagnostics

	if isInModule {
		diags.Add(diag.Diagnostic{
			Severity: diag.SeverityLevelError,
			Message:  "tracing block not allowed inside a module",
			StartPos: ast.StartPos(block).Position(),
			EndPos:   ast.EndPos(block).Position(),
		})

		return nil, diags
	}

	return &TracingConfigNode{
		label:         block.Label,
		nodeID:        BlockComponentID(block).String(),
		componentName: block.GetBlockName(),
		traceProvider: globals.TraceProvider,

		block: block,
		eval:  vm.New(block.Body),
	}, diags
}

// Evaluate implements BlockNode and updates the arguments for the managed config block
// by re-evaluating its River block with the provided scope. The managed config block
// will be built the first time Evaluate is called.
//
// Evaluate will return an error if the River block cannot be evaluated or if
// decoding to arguments fails.
func (cn *TracingConfigNode) Evaluate(scope *vm.Scope) error {
	cn.mut.RLock()
	defer cn.mut.RUnlock()
	args := tracing.DefaultOptions
	if err := cn.eval.Evaluate(scope, &args); err != nil {
		return fmt.Errorf("decoding River: %w", err)
	}

	t, ok := cn.traceProvider.(*tracing.Tracer)
	if ok {
		err := t.Update(args)
		if err != nil {
			return fmt.Errorf("could not update tracer: %v", err)
		}
	}
	return nil
}

// Block implements BlockNode and returns the current block of the managed config node.
func (cn *TracingConfigNode) Block() *ast.BlockStmt {
	cn.mut.RLock()
	defer cn.mut.RUnlock()
	return cn.block
}

// NodeID implements dag.Node and returns the unique ID for the config node.
func (cn *TracingConfigNode) NodeID() string { return cn.nodeID }
