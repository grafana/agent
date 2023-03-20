package controller

import (
	"fmt"

	"github.com/grafana/agent/pkg/flow/internal/dag"
	"github.com/grafana/agent/pkg/flow/tracing"
	"github.com/grafana/agent/pkg/river/ast"
	"github.com/grafana/agent/pkg/river/diag"
	"github.com/grafana/agent/pkg/river/vm"
)

type TracingConfigNode struct {
	configNode ConfigNode
}

var _ dag.Node = (*TracingConfigNode)(nil)

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
		configNode: ConfigNode{
			label:         block.Label,
			nodeID:        BlockComponentID(block).String(),
			componentName: GetBlockName(block),
			globals:       globals,

			block: block,
			eval:  vm.New(block.Body),
		},
	}, diags
}

// Evaluate implements dag.Node and updates the arguments for the managed config block
// by re-evaluating its River block with the provided scope. The managed config block
// will be built the first time Evaluate is called.
//
// Evaluate will return an error if the River block cannot be evaluated or if
// decoding to arguments fails.
func (cn *TracingConfigNode) Evaluate(scope *vm.Scope) error {
	cn.configNode.mut.Lock()
	defer cn.configNode.mut.Unlock()
	args := tracing.DefaultOptions
	if err := cn.configNode.eval.Evaluate(scope, &args); err != nil {
		return fmt.Errorf("decoding River: %w", err)
	}

	t, ok := cn.configNode.globals.TraceProvider.(*tracing.Tracer)
	if ok {
		err := t.Update(args)
		if err != nil {
			return fmt.Errorf("could not update tracer: %v", err)
		}
	}
	return nil
}

// Block implements dag.Node and returns the current block of the managed config node.
func (cn *TracingConfigNode) Block() *ast.BlockStmt {
	cn.configNode.mut.RLock()
	defer cn.configNode.mut.RUnlock()
	return cn.configNode.block
}

// NodeID implements dag.Node and returns the unique ID for the config node.
func (cn *TracingConfigNode) NodeID() string { return cn.configNode.nodeID }
