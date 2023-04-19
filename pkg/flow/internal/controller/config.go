package controller

import (
	"fmt"

	"github.com/grafana/agent/pkg/river/ast"
	"github.com/grafana/agent/pkg/river/diag"
)

const (
	argumentBlockID = "argument"
	exportBlockID   = "export"
	loggingBlockID  = "logging"
	tracingBlockID  = "tracing"
)

// NewConfigNode creates a new ConfigNode from an initial ast.BlockStmt.
// The underlying config isn't applied until Evaluate is called.
func NewConfigNode(block *ast.BlockStmt, globals ComponentGlobals, isInModule bool) (BlockNode, diag.Diagnostics) {
	switch block.GetBlockName() {
	case argumentBlockID:
		return NewArgumentConfigNode(block, globals, isInModule)
	case exportBlockID:
		return NewExportConfigNode(block, globals, isInModule)
	case loggingBlockID:
		return NewLoggingConfigNode(block, globals, isInModule)
	case tracingBlockID:
		return NewTracingConfigNode(block, globals, isInModule)
	default:
		var diags diag.Diagnostics
		diags.Add(diag.Diagnostic{
			Severity: diag.SeverityLevelError,
			Message:  fmt.Sprintf("invalid config block type %s while creating new config node", block.GetBlockName()),
			StartPos: ast.StartPos(block).Position(),
			EndPos:   ast.EndPos(block).Position(),
		})
		return nil, diags
	}
}
