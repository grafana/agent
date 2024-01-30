package controller

import (
	"fmt"
	"strings"

	"github.com/grafana/agent/pkg/flow/internal/importsource"
	"github.com/grafana/river/ast"
	"github.com/grafana/river/diag"
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
	case importsource.BlockImportFile, importsource.BlockImportGit, importsource.BlockImportHTTP:
		return NewImportConfigNode(block, globals, importsource.GetSourceType(block.GetBlockName())), nil
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
	importMap   map[string]*ImportConfigNode
}

// NewConfigNodeMap will create an initial ConfigNodeMap. Append must be called
// to populate NewConfigNodeMap.
func NewConfigNodeMap() *ConfigNodeMap {
	return &ConfigNodeMap{
		logging:     nil,
		tracing:     nil,
		argumentMap: map[string]*ArgumentConfigNode{},
		exportMap:   map[string]*ExportConfigNode{},
		importMap:   map[string]*ImportConfigNode{},
	}
}

// Append will add a config node to the ConfigNodeMap. This will overwrite
// values on the ConfigNodeMap that are matched and previously set.
func (nodeMap *ConfigNodeMap) Append(configNode BlockNode) diag.Diagnostics {
	var diags diag.Diagnostics

	switch n := configNode.(type) {
	case *ArgumentConfigNode:
		nodeMap.argumentMap[n.Label()] = n
	case *ExportConfigNode:
		nodeMap.exportMap[n.Label()] = n
	case *LoggingConfigNode:
		nodeMap.logging = n
	case *TracingConfigNode:
		nodeMap.tracing = n
	case *ImportConfigNode:
		nodeMap.importMap[n.Label()] = n
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
func (nodeMap *ConfigNodeMap) Validate(isInModule bool, args map[string]any) diag.Diagnostics {
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
		return diags
	}

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

	return diags
}

// ValidateUnsupportedArguments will validate each provided argument is
// supported in the config.
func (nodeMap *ConfigNodeMap) ValidateUnsupportedArguments(args map[string]any) diag.Diagnostics {
	var diags diag.Diagnostics

	for argName := range args {
		if _, found := nodeMap.argumentMap[argName]; found {
			continue
		}
		diags.Add(diag.Diagnostic{
			Severity: diag.SeverityLevelError,
			Message:  fmt.Sprintf("Provided argument %q is not defined in the module", argName),
		})
	}

	return diags
}

func (nodeMap *ConfigNodeMap) findImportNodeReferences(declare *Declare) map[*ImportConfigNode]struct{} {
	uniqueReferences := make(map[*ImportConfigNode]struct{})
	nodeMap.collectImportNodeReferences(declare.block.Body, uniqueReferences)
	return uniqueReferences
}

// collectCustomComponentDependencies collects recursively references to import nodes through an AST body.
func (nodeMap *ConfigNodeMap) collectImportNodeReferences(stmts ast.Body, uniqueReferences map[*ImportConfigNode]struct{}) {
	for _, stmt := range stmts {
		switch stmt := stmt.(type) {
		case *ast.BlockStmt:
			componentName := strings.Join(stmt.Name, ".")
			switch componentName {
			case "declare":
				nodeMap.collectImportNodeReferences(stmt.Body, uniqueReferences)
			default:
				potentialImportLabel := stmt.Name[0]
				if node, exists := nodeMap.importMap[potentialImportLabel]; exists {
					uniqueReferences[node] = struct{}{}
				}
			}
		}
	}
}
