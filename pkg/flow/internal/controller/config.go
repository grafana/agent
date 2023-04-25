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

// ConfigNodeMap represents the config BlockNodes in their explicit types.
// This is helpful when validating node conditions specific to config node
// types.
type ConfigNodeMap struct {
	logging     *LoggingConfigNode
	tracing     *TracingConfigNode
	argumentMap map[string]*ArgumentConfigNode
	exportMap   map[string]*ExportConfigNode
}

// NewConfigNodeMap will create an initial ConfigNodeMap. Append must be called
// to populate NewConfigNodeMap.
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
				Message:  fmt.Sprintf("argument config block %q already declared at %s", n.Label(), ast.StartPos(nodeMap.argumentMap[n.Label()].Block()).Position()),
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
				Message:  fmt.Sprintf("export config block %q already declared at %s", n.Label(), ast.StartPos(nodeMap.exportMap[n.Label()].Block()).Position()),
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
				Message:  fmt.Sprintf("logging config block %q already declared at %s", loggingBlockID, ast.StartPos(nodeMap.logging.Block()).Position()),
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
				Message:  fmt.Sprintf("tracing config block %q already declared at %s", tracingBlockID, ast.StartPos(nodeMap.tracing.Block()).Position()),
				StartPos: ast.StartPos(n.Block()).Position(),
				EndPos:   ast.EndPos(n.Block()).Position(),
			})
		}
	default:
		diags.Add(diag.Diagnostic{
			Severity: diag.SeverityLevelError,
			Message:  fmt.Sprintf("unsupported config node type found %q", n.Block().Name),
			StartPos: ast.StartPos(n.Block()).Position(),
			EndPos:   ast.EndPos(n.Block()).Position(),
		})
	}

	return diags
}

// Validate wraps all validators for ConfigNodeMap.
func (nodeMap *ConfigNodeMap) Validate(isInModule bool, args *map[string]any) diag.Diagnostics {
	var diags diag.Diagnostics

	newDiags := nodeMap.ValidateModuleConstraints(isInModule)
	diags = append(diags, newDiags...)

	newDiags = nodeMap.ValidateUnsupportedArguments(args)
	diags = append(diags, newDiags...)

	return diags
}

// ValidateModuleConstraints will make sure config blocks with module
// constraints get followed.
func (nodeMap *ConfigNodeMap) ValidateModuleConstraints(isInModule bool) diag.Diagnostics {
	var diags diag.Diagnostics

	if isInModule {
		if nodeMap.logging != nil {
			diags.Add(diag.Diagnostic{
				Severity: diag.SeverityLevelError,
				Message:  "logging block not allowed inside a module",
				StartPos: ast.StartPos(nodeMap.logging.Block()).Position(),
				EndPos:   ast.EndPos(nodeMap.logging.Block()).Position(),
			})
		}

		if nodeMap.tracing != nil {
			diags.Add(diag.Diagnostic{
				Severity: diag.SeverityLevelError,
				Message:  "tracing block not allowed inside a module",
				StartPos: ast.StartPos(nodeMap.tracing.Block()).Position(),
				EndPos:   ast.EndPos(nodeMap.tracing.Block()).Position(),
			})
		}
	} else {
		for key := range nodeMap.argumentMap {
			diags.Add(diag.Diagnostic{
				Severity: diag.SeverityLevelError,
				Message:  "argument blocks only allowed inside a module",
				StartPos: ast.StartPos(nodeMap.argumentMap[key].Block()).Position(),
				EndPos:   ast.EndPos(nodeMap.argumentMap[key].Block()).Position(),
			})
		}

		for key := range nodeMap.exportMap {
			diags.Add(diag.Diagnostic{
				Severity: diag.SeverityLevelError,
				Message:  "export blocks only allowed inside a module",
				StartPos: ast.StartPos(nodeMap.exportMap[key].Block()).Position(),
				EndPos:   ast.EndPos(nodeMap.exportMap[key].Block()).Position(),
			})
		}
	}

	return diags
}

// ValidateUnsupportedArguments will validate each provided argument is
// supported in the config.
func (nodeMap *ConfigNodeMap) ValidateUnsupportedArguments(args *map[string]any) diag.Diagnostics {
	var diags diag.Diagnostics

	if args != nil {
		for argName := range *args {
			if _, ok := nodeMap.argumentMap[argName]; !ok {
				diags.Add(diag.Diagnostic{
					Severity: diag.SeverityLevelError,
					Message:  fmt.Sprintf("Unsupported argument %q was provided to a module.", argName),
				})
			}
		}
	}

	return diags
}
