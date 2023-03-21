package controller

import (
	"fmt"

	"github.com/grafana/agent/pkg/flow/tracing"
	"github.com/grafana/agent/pkg/river/ast"
	"github.com/grafana/agent/pkg/river/diag"
	"github.com/grafana/agent/pkg/river/vm"
	"go.opentelemetry.io/otel/trace"
)

type TracingConfigNode struct {
	configNode    sharedBlockNode
	traceProvider trace.TracerProvider // Tracer shared between all managed components.
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
	}

	return &TracingConfigNode{
		configNode: sharedBlockNode{
			label:         block.Label,
			nodeID:        BlockComponentID(block).String(),
			componentName: block.GetBlockName(),

			block: block,
			eval:  vm.New(block.Body),
		},
		traceProvider: globals.TraceProvider,
	}, diags
}

// Evaluate implements BlockNode and updates the arguments for the managed config block
// by re-evaluating its River block with the provided scope. The managed config block
// will be built the first time Evaluate is called.
//
// Evaluate will return an error if the River block cannot be evaluated or if
// decoding to arguments fails.
func (cn *TracingConfigNode) Evaluate(scope *vm.Scope) error {
	args := tracing.DefaultOptions
	if err := cn.configNode.Evaluate(scope, &args); err != nil {
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

// ComponentName implements BlockNode and returns the component's type, i.e. `local.file.test` returns `local.file`.
func (cn *TracingConfigNode) ComponentName() string { return cn.configNode.componentName }

// Label implements BlockNode and returns the label for the block or "" if none was specified.
func (cn *TracingConfigNode) Label() string { return cn.configNode.label }

// Block implements BlockNode and returns the current block of the managed config node.
func (cn *TracingConfigNode) Block() *ast.BlockStmt { return cn.configNode.Block() }

// NodeID implements dag.Node and returns the unique ID for the config node.
func (cn *TracingConfigNode) NodeID() string { return cn.configNode.nodeID }
