package controller

import (
	"fmt"

	"github.com/grafana/agent/pkg/flow/internal/dag"
	"github.com/grafana/agent/pkg/river/ast"
	"github.com/grafana/agent/pkg/river/diag"
	"github.com/grafana/agent/pkg/river/vm"
)

type ExportConfigNode struct {
	configNode ConfigNode
}

var _ dag.Node = (*ExportConfigNode)(nil)

// NewExportConfigNode creates a new ExportConfigNode from an initial ast.BlockStmt.
// The underlying config isn't applied until Evaluate is called.
func NewExportConfigNode(block *ast.BlockStmt, globals ComponentGlobals, isInModule bool) (*ExportConfigNode, diag.Diagnostics) {
	var diags diag.Diagnostics

	if !isInModule {
		diags.Add(diag.Diagnostic{
			Severity: diag.SeverityLevelError,
			Message:  "export blocks only allowed inside a module",
			StartPos: ast.StartPos(block).Position(),
			EndPos:   ast.EndPos(block).Position(),
		})
	}

	return &ExportConfigNode{
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

type exportBlock struct {
	Value any `river:"value,attr"`
}

// Evaluate implements dag.Node and updates the arguments for the managed config block
// by re-evaluating its River block with the provided scope. The managed config block
// will be built the first time Evaluate is called.
//
// Evaluate will return an error if the River block cannot be evaluated or if
// decoding to arguments fails.
func (cn *ExportConfigNode) Evaluate(scope *vm.Scope) error {
	cn.configNode.mut.Lock()
	defer cn.configNode.mut.Unlock()
	exports := make(map[string]any, 1)

	var export exportBlock
	if err := cn.configNode.eval.Evaluate(scope, &export); err != nil {
		return fmt.Errorf("decoding River: %w", err)
	}

	exports[cn.configNode.label] = export.Value

	if cn.configNode.globals.OnExportsChange != nil {
		cn.configNode.globals.OnExportsChange(exports)
	}
	return nil
}

// Block implements dag.Node and returns the current block of the managed config node.
func (cn *ExportConfigNode) Block() *ast.BlockStmt {
	cn.configNode.mut.RLock()
	defer cn.configNode.mut.RUnlock()
	return cn.configNode.block
}

// NodeID implements dag.Node and returns the unique ID for the config node.
func (cn *ExportConfigNode) NodeID() string { return cn.configNode.nodeID }
