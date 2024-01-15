package controller

import (
	"strings"

	"github.com/grafana/river/ast"
)

type ModuleReference struct {
	componentName string
	importLabel   string
	declareLabel  string
	importNode    *ImportConfigNode
	declareNode   *DeclareNode
}

// GetModuleReferences traverses the AST of the provided declare and collects references to known modules.
// Panics if declare is nil.
func GetModuleReferences(
	declare *Declare,
	importNodes map[string]*ImportConfigNode,
	declareNodes map[string]*DeclareNode,
	parentModuleDefinitions map[string]string,
) ([]ModuleReference, error) {

	uniqueReferences := make(map[string]ModuleReference)
	getModuleReferences(declare.Block.Body, importNodes, declareNodes, uniqueReferences, parentModuleDefinitions)

	references := make([]ModuleReference, 0, len(uniqueReferences))
	for _, ref := range uniqueReferences {
		references = append(references, ref)
	}

	return references, nil
}

func getModuleReferences(
	stmts ast.Body,
	importNodes map[string]*ImportConfigNode,
	declareNodes map[string]*DeclareNode,
	uniqueReferences map[string]ModuleReference,
	parentModuleDefinitions map[string]string,
) {
	for _, stmt := range stmts {
		switch stmt := stmt.(type) {
		case *ast.BlockStmt:
			componentName := strings.Join(stmt.Name, ".")
			switch componentName {
			case "declare":
				getModuleReferences(stmt.Body, importNodes, declareNodes, uniqueReferences, parentModuleDefinitions)
			default:
				potentialImportLabel, potentialDeclareLabel := ExtractImportAndDeclareLabels(componentName)
				if declareNode, ok := declareNodes[potentialDeclareLabel]; ok {
					uniqueReferences[componentName] = ModuleReference{componentName: componentName, importLabel: "", declareLabel: potentialDeclareLabel, declareNode: declareNode}
				} else if importNode, ok := importNodes[potentialImportLabel]; ok {
					uniqueReferences[componentName] = ModuleReference{componentName: componentName, importLabel: potentialImportLabel, declareLabel: potentialDeclareLabel, importNode: importNode}
				} else if _, ok := parentModuleDefinitions[componentName]; ok {
					uniqueReferences[componentName] = ModuleReference{componentName: componentName, importLabel: potentialImportLabel, declareLabel: potentialDeclareLabel}
				}
			}
		}
	}
}
