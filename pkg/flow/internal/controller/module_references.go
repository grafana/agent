package controller

import (
	"strings"

	"github.com/grafana/river/ast"
	"github.com/grafana/river/parser"
)

type ModuleReference struct {
	fullName              string // declareLabel / import1.declareLabel / import1.import2.declareLabel
	namespace             string // ""           / import1             / import1
	scopedName            string // declareLabel / declareLabel       / import2.declareLabel
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
			fullName := strings.Join(stmt.Name, ".")
			switch fullName {
			case "declare":
				declareContent := string(content[stmt.LCurlyPos.Position().Offset+1 : stmt.RCurlyPos.Position().Offset-1])
				err = getModuleReferences(declareContent, importNodes, declareNodes, uniqueReferences, parentModuleDependencies)
				if err != nil {
					return err
				}
			default:
				parts := strings.Split(fullName, ".")
				firstPart := parts[0]
				var scopedName string
				if len(parts) > 1 {
					scopedName = strings.Join(parts[1:], ".")
				}

				if declareNode, ok := declareNodes[firstPart]; ok {
					uniqueReferences[fullName] = ModuleReference{fullName: fullName, namespace: "", scopedName: firstPart, moduleContentProvider: declareNode}
				} else if importNode, ok := importNodes[firstPart]; ok {
					uniqueReferences[fullName] = ModuleReference{fullName: fullName, namespace: firstPart, scopedName: scopedName, moduleContentProvider: importNode}
				} else if _, ok := parentModuleDependencies[fullName]; ok {
					uniqueReferences[fullName] = ModuleReference{fullName: fullName, namespace: firstPart, scopedName: scopedName, moduleContentProvider: nil}
				}
			}
		}
	}
	return nil
}
