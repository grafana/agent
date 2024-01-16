package controller

import (
	"strings"

	"github.com/grafana/river/ast"
)

type CustomComponentDependency struct {
	componentName string
	importLabel   string
	declareLabel  string
	importNode    *ImportConfigNode
	declareNode   *DeclareNode
}

// GetCustomComponentDependencies traverses the AST of the provided declare and collects references to known custom components.
// Panics if declare is nil.
func GetCustomComponentDependencies(
	declare *Declare,
	importNodes map[string]*ImportConfigNode,
	declareNodes map[string]*DeclareNode,
	parentDeclareContents map[string]string,
) ([]CustomComponentDependency, error) {

	uniqueReferences := make(map[string]CustomComponentDependency)
	getCustomComponentDependencies(declare.Block.Body, importNodes, declareNodes, uniqueReferences, parentDeclareContents)

	references := make([]CustomComponentDependency, 0, len(uniqueReferences))
	for _, ref := range uniqueReferences {
		references = append(references, ref)
	}

	return references, nil
}

func getCustomComponentDependencies(
	stmts ast.Body,
	importNodes map[string]*ImportConfigNode,
	declareNodes map[string]*DeclareNode,
	uniqueReferences map[string]CustomComponentDependency,
	parentDeclareContents map[string]string,
) {
	for _, stmt := range stmts {
		switch stmt := stmt.(type) {
		case *ast.BlockStmt:
			componentName := strings.Join(stmt.Name, ".")
			switch componentName {
			case "declare":
				getCustomComponentDependencies(stmt.Body, importNodes, declareNodes, uniqueReferences, parentDeclareContents)
			default:
				potentialImportLabel, potentialDeclareLabel := ExtractImportAndDeclareLabels(componentName)
				if declareNode, ok := declareNodes[potentialDeclareLabel]; ok {
					uniqueReferences[componentName] = CustomComponentDependency{componentName: componentName, importLabel: "", declareLabel: potentialDeclareLabel, declareNode: declareNode}
				} else if importNode, ok := importNodes[potentialImportLabel]; ok {
					uniqueReferences[componentName] = CustomComponentDependency{componentName: componentName, importLabel: potentialImportLabel, declareLabel: potentialDeclareLabel, importNode: importNode}
				} else if _, ok := parentDeclareContents[componentName]; ok {
					uniqueReferences[componentName] = CustomComponentDependency{componentName: componentName, importLabel: potentialImportLabel, declareLabel: potentialDeclareLabel}
				}
			}
		}
	}
}
