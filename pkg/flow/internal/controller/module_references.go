package controller

import (
	"strings"

	"github.com/grafana/river/ast"
	"github.com/grafana/river/parser"
)

type ModuleReference struct {
	componentName         string
	importLabel           string
	declareLabel          string
	moduleContentProvider ModuleContentProvider
}

// This function will parse the provided river content and collect references to known modules.
func GetModuleReferences(
	content string,
	importNodes map[string]*ImportConfigNode,
	declareNodes map[string]*DeclareNode,
	parentModuleDependencies map[string]string,
) ([]ModuleReference, error) {
	uniqueReferences := make(map[string]ModuleReference)
	err := getModuleReferences(content, importNodes, declareNodes, uniqueReferences, parentModuleDependencies)
	if err != nil {
		return nil, err
	}

	references := make([]ModuleReference, 0, len(uniqueReferences))
	for _, ref := range uniqueReferences {
		references = append(references, ref)
	}

	return references, nil
}

func getModuleReferences(
	content string,
	importNodes map[string]*ImportConfigNode,
	declareNodes map[string]*DeclareNode,
	uniqueReferences map[string]ModuleReference,
	parentModuleDependencies map[string]string,
) error {

	node, err := parser.ParseFile("", []byte(content))
	if err != nil {
		return err
	}

	for _, stmt := range node.Body {
		switch stmt := stmt.(type) {
		case *ast.BlockStmt:
			componentName := strings.Join(stmt.Name, ".")
			switch componentName {
			case "declare":
				declareContent := string(content[stmt.LCurlyPos.Position().Offset+1 : stmt.RCurlyPos.Position().Offset-1])
				err = getModuleReferences(declareContent, importNodes, declareNodes, uniqueReferences, parentModuleDependencies)
				if err != nil {
					return err
				}
			default:
				potentialImportLabel, potentialDeclareLabel := ExtractImportAndDeclareLabels(componentName)
				if declareNode, ok := declareNodes[potentialDeclareLabel]; ok {
					uniqueReferences[componentName] = ModuleReference{componentName: componentName, importLabel: "", declareLabel: potentialDeclareLabel, moduleContentProvider: declareNode}
				} else if importNode, ok := importNodes[potentialImportLabel]; ok {
					uniqueReferences[componentName] = ModuleReference{componentName: componentName, importLabel: potentialImportLabel, declareLabel: potentialDeclareLabel, moduleContentProvider: importNode}
				} else if _, ok := parentModuleDependencies[componentName]; ok {
					uniqueReferences[componentName] = ModuleReference{componentName: componentName, importLabel: potentialImportLabel, declareLabel: potentialDeclareLabel, moduleContentProvider: nil}
				}
			}
		}
	}
	return nil
}
