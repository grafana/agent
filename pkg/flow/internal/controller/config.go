package controller

import (
	"fmt"
	"strings"
	"sync"

	"github.com/grafana/agent/pkg/flow/internal/dag"
	"github.com/grafana/agent/pkg/river/ast"
	"github.com/grafana/agent/pkg/river/diag"
	"github.com/grafana/agent/pkg/river/vm"
)

const (
	exportBlockID  = "export"
	loggingBlockID = "logging"
	tracingBlockID = "tracing"
)

// Shared config properties for all config block nodes
type ConfigNode struct {
	label         string
	nodeID        string
	componentName string
	globals       ComponentGlobals

	mut   sync.RWMutex
	block *ast.BlockStmt // Current River blocks to derive config from
	eval  *vm.Evaluator
}

// NewConfigNode creates a new ConfigNode from an initial ast.BlockStmt.
// The underlying config isn't applied until Evaluate is called.
func NewConfigNode(block *ast.BlockStmt, globals ComponentGlobals, isInModule bool) (dag.Node, diag.Diagnostics) {
	switch GetBlockName(block) {
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
			Message:  fmt.Sprintf("invalid config block type %s while creating new config node", GetBlockName(block)),
			StartPos: ast.StartPos(block).Position(),
			EndPos:   ast.EndPos(block).Position(),
		})
		return nil, diags
	}
}

// Helper method for how to get the "." delimited block name.
func GetBlockName(block *ast.BlockStmt) string {
	return strings.Join(block.Name, ".")
}
