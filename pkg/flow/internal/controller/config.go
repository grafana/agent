package controller

import (
	"fmt"

	"github.com/grafana/agent/pkg/flow/internal/dag"
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
func NewConfigNode(block *ast.BlockStmt, globals ComponentGlobals) (BlockNode, diag.Diagnostics) {
	switch block.GetBlockName() {
	case argumentBlockID:
		return NewArgumentConfigNode(block, globals), nil
	case exportBlockID:
		return NewExportConfigNode(block, globals), nil
	case loggingBlockID:
		return NewLoggingConfigNode(block, globals), nil
	case tracingBlockID:
		return NewTracingConfigNode(block, globals), nil
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

// getLoggingNode returns the logging node on the graph or nil if it doesn't exist.
func getLoggingNode(g *dag.Graph) *LoggingConfigNode {
	node := g.GetByID(loggingBlockID)
	if node == nil {
		return nil
	}

	switch n := node.(type) {
	case *LoggingConfigNode:
		return n
	default:
		panic("Invalid logging node in the graph")
	}
}

// getTracingNode returns the tracing node on the graph or nil if it doesn't exist.
func getTracingNode(g *dag.Graph) *TracingConfigNode {
	node := g.GetByID(tracingBlockID)
	if node == nil {
		return nil
	}

	switch n := node.(type) {
	case *TracingConfigNode:
		return n
	default:
		panic("Invalid tracing node in the graph")
	}
}

// getArgumentNodes returns a slice of all argument config nodes on the graph.
func getArgumentNodes(g *dag.Graph) []*ArgumentConfigNode {
	nodes := g.GetByIDPrefix(argumentBlockID + ".")
	argumentNodes := make([]*ArgumentConfigNode, 0)

	for _, node := range nodes {
		switch n := node.(type) {
		case *ArgumentConfigNode:
			argumentNodes = append(argumentNodes, n)
		default:
			panic("Invalid argument node in the graph")
		}
	}

	return argumentNodes
}

// getExportNodes returns a slice of all export config nodes on the graph.
func getExportNodes(g *dag.Graph) []*ExportConfigNode {
	nodes := g.GetByIDPrefix(exportBlockID + ".")
	exportNodes := make([]*ExportConfigNode, 0)

	for _, node := range nodes {
		switch n := node.(type) {
		case *ExportConfigNode:
			exportNodes = append(exportNodes, n)
		default:
			panic("Invalid export node in the graph")
		}
	}

	return exportNodes
}

// Validate wraps all validators for ConfigNodeMap.
func validateConfigNodes(g *dag.Graph, isInModule bool, args map[string]any) diag.Diagnostics {
	var diags diag.Diagnostics

	newDiags := validateModuleConstraints(g, isInModule)
	diags = append(diags, newDiags...)

	newDiags = validateUnsupportedArguments(g, args)
	diags = append(diags, newDiags...)

	return diags
}

// validateModuleConstraints will make sure config blocks with module
// constraints get followed.
func validateModuleConstraints(g *dag.Graph, isInModule bool) diag.Diagnostics {
	var diags diag.Diagnostics

	if isInModule {
		loggingNode := getLoggingNode(g)
		if loggingNode != nil {
			diags.Add(diag.Diagnostic{
				Severity: diag.SeverityLevelError,
				Message:  "logging block not allowed inside a module",
				StartPos: ast.StartPos(loggingNode.Block()).Position(),
				EndPos:   ast.EndPos(loggingNode.Block()).Position(),
			})
		}

		tracingNode := getTracingNode(g)
		if tracingNode != nil {
			diags.Add(diag.Diagnostic{
				Severity: diag.SeverityLevelError,
				Message:  "tracing block not allowed inside a module",
				StartPos: ast.StartPos(tracingNode.Block()).Position(),
				EndPos:   ast.EndPos(tracingNode.Block()).Position(),
			})
		}
		return diags
	}

	for _, node := range getArgumentNodes(g) {
		diags.Add(diag.Diagnostic{
			Severity: diag.SeverityLevelError,
			Message:  "argument blocks only allowed inside a module",
			StartPos: ast.StartPos(node.Block()).Position(),
			EndPos:   ast.EndPos(node.Block()).Position(),
		})
	}

	for _, node := range getExportNodes(g) {
		diags.Add(diag.Diagnostic{
			Severity: diag.SeverityLevelError,
			Message:  "export blocks only allowed inside a module",
			StartPos: ast.StartPos(node.Block()).Position(),
			EndPos:   ast.EndPos(node.Block()).Position(),
		})
	}

	return diags
}

// validateUnsupportedArguments will validate each provided argument is
// supported in the config.
func validateUnsupportedArguments(g *dag.Graph, args map[string]any) diag.Diagnostics {
	var diags diag.Diagnostics

	for argName := range args {
		if g.GetByID("argument."+argName) != nil {
			continue
		}
		diags.Add(diag.Diagnostic{
			Severity: diag.SeverityLevelError,
			Message:  fmt.Sprintf("Provided argument %q is not defined in the module", argName),
		})
	}

	return diags
}
