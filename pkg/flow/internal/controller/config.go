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

type ConfigNodeMap struct {
	logging     *LoggingConfigNode
	tracing     *TracingConfigNode
	argumentMap map[string]*ArgumentConfigNode
	exportMap   map[string]*ExportConfigNode
}

func NewConfigNodeMap() *ConfigNodeMap {
	return &ConfigNodeMap{
		logging:     nil,
		tracing:     nil,
		argumentMap: map[string]*ArgumentConfigNode{},
		exportMap:   map[string]*ExportConfigNode{},
	}
}

// Append will add a config node to the ConfigNodeMap unless it already exists
// in which case it will return a diag specifying the problem.
func (nodeMap *ConfigNodeMap) Append(configNode BlockNode) diag.Diagnostics {
	var diags diag.Diagnostics

	switch n := configNode.(type) {
	case *ArgumentConfigNode:
		if _, exists := nodeMap.argumentMap[n.Label()]; exists {
			diags.Add(diag.Diagnostic{
				Severity: diag.SeverityLevelError,
				Message:  fmt.Sprintf("argument config block \"%s\" already declared at %s", n.Label(), ast.StartPos(nodeMap.argumentMap[n.Label()].Block()).Position()),
				StartPos: ast.StartPos(n.Block()).Position(),
				EndPos:   ast.EndPos(n.Block()).Position(),
			})

			return diags
		}

		nodeMap.argumentMap[n.Label()] = n
	case *ExportConfigNode:
		if _, exists := nodeMap.exportMap[n.Label()]; !exists {
			nodeMap.exportMap[n.Label()] = n
		} else {
			diags.Add(diag.Diagnostic{
				Severity: diag.SeverityLevelError,
				Message:  fmt.Sprintf("export config block \"%s\" already declared at %s", n.Label(), ast.StartPos(nodeMap.exportMap[n.Label()].Block()).Position()),
				StartPos: ast.StartPos(n.Block()).Position(),
				EndPos:   ast.EndPos(n.Block()).Position(),
			})
		}
	case *LoggingConfigNode:
		if nodeMap.logging == nil {
			nodeMap.logging = n
		} else {
			diags.Add(diag.Diagnostic{
				Severity: diag.SeverityLevelError,
				Message:  fmt.Sprintf("logging config block \"%s\" already declared at %s", loggingBlockID, ast.StartPos(nodeMap.logging.Block()).Position()),
				StartPos: ast.StartPos(n.Block()).Position(),
				EndPos:   ast.EndPos(n.Block()).Position(),
			})
		}
	case *TracingConfigNode:
		if nodeMap.tracing == nil {
			nodeMap.tracing = n
		} else {
			diags.Add(diag.Diagnostic{
				Severity: diag.SeverityLevelError,
				Message:  fmt.Sprintf("tracing config block \"%s\" already declared at %s", tracingBlockID, ast.StartPos(nodeMap.tracing.Block()).Position()),
				StartPos: ast.StartPos(n.Block()).Position(),
				EndPos:   ast.EndPos(n.Block()).Position(),
			})
		}
	default:
		diags.Add(diag.Diagnostic{
			Severity: diag.SeverityLevelError,
			Message:  fmt.Sprintf("unsupported config node type found \"%s\"", n.Block().Name),
			StartPos: ast.StartPos(n.Block()).Position(),
			EndPos:   ast.EndPos(n.Block()).Position(),
		})
	}

	return diags
}
