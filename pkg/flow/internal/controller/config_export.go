package controller

import (
	"fmt"

	"github.com/grafana/agent/pkg/river/ast"
	"github.com/grafana/agent/pkg/river/diag"
	"github.com/grafana/agent/pkg/river/vm"
)

type ExportConfigNode struct {
	configNode      sharedBlockNode
	onExportsChange func(exports map[string]any) // Invoked when the managed component updated its exports
}

var _ BlockNode = (*ExportConfigNode)(nil)

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
		configNode: sharedBlockNode{
			label:         block.Label,
			nodeID:        BlockComponentID(block).String(),
			componentName: block.GetBlockName(),

			block: block,
			eval:  vm.New(block.Body),
		},

		onExportsChange: globals.OnExportsChange,
	}, diags
}

type exportBlock struct {
	Value any `river:"value,attr"`
}

// Evaluate implements BlockNode and updates the arguments for the managed config block
// by re-evaluating its River block with the provided scope. The managed config block
// will be built the first time Evaluate is called.
//
// Evaluate will return an error if the River block cannot be evaluated or if
// decoding to arguments fails.
func (cn *ExportConfigNode) Evaluate(scope *vm.Scope) error {
	exports := make(map[string]any)

	var export exportBlock
	if err := cn.configNode.Evaluate(scope, &export); err != nil {
		return fmt.Errorf("decoding River: %w", err)
	}

	exports[cn.Label()] = export.Value

	if cn.onExportsChange != nil {
		cn.onExportsChange(exports)
	}
	return nil
}

// ComponentName implements BlockNode and returns the component's type, i.e. `local.file.test` returns `local.file`.
func (cn *ExportConfigNode) ComponentName() string { return cn.configNode.componentName }

// Label implements BlockNode and returns the label for the block or "" if none was specified.
func (cn *ExportConfigNode) Label() string { return cn.configNode.label }

// Block implements BlockNode and returns the current block of the managed config node.
func (cn *ExportConfigNode) Block() *ast.BlockStmt { return cn.configNode.Block() }

// NodeID implements dag.Node and returns the unique ID for the config node.
func (cn *ExportConfigNode) NodeID() string { return cn.configNode.nodeID }
