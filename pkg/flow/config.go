package flow

import (
	"fmt"
	"strings"

	"github.com/grafana/agent/pkg/river/ast"
	"github.com/grafana/agent/pkg/river/diag"
	"github.com/grafana/agent/pkg/river/parser"
)

// An Argument is an input to a Flow module.
type Argument struct {
	// Name of the argument.
	Name string `river:",label"`

	// Whether the Argument must be provided when evaluating the file.
	Optional bool `river:"optional,attr,optional"`

	// Description for the Argument.
	Comment string `river:"comment,attr,optional"`

	// Default value for the argument.
	Default any `river:"default,attr,optional"`
}

// File holds the contents of a parsed Flow file.
type File struct {
	name string    // File name given to ReadFile.
	node *ast.File // Raw File node.

	// components holds the list of raw River AST blocks describing components.
	// The Flow controller can interpret them.
	components   []*ast.BlockStmt
	configBlocks []*ast.BlockStmt
}

// ReadFile parses the River file specified by bb into a File. name should be
// the name of the file used for reporting errors.
func ReadFile(name string, bb []byte) (*File, error) {
	node, err := parser.ParseFile(name, bb)
	if err != nil {
		return nil, err
	}

	// Look for predefined non-components blocks (i.e., logging), and store
	// everything else into a list of components.
	//
	// TODO(rfratto): should this code be brought into a helper somewhere? Maybe
	// in ast?
	var (
		components []*ast.BlockStmt
		configs    []*ast.BlockStmt
	)

	for _, stmt := range node.Body {
		switch stmt := stmt.(type) {
		case *ast.AttributeStmt:
			return nil, diag.Diagnostic{
				Severity: diag.SeverityLevelError,
				StartPos: ast.StartPos(stmt.Name).Position(),
				EndPos:   ast.EndPos(stmt.Name).Position(),
				Message:  "unrecognized attribute " + stmt.Name.Name,
			}

		case *ast.BlockStmt:
			fullName := strings.Join(stmt.Name, ".")
			switch fullName {
			case "logging":
				configs = append(configs, stmt)
			case "tracing":
				configs = append(configs, stmt)
			case "argument":
				configs = append(configs, stmt)
			case "export":
				configs = append(configs, stmt)
			default:
				components = append(components, stmt)
			}

		default:
			return nil, diag.Diagnostic{
				Severity: diag.SeverityLevelError,
				StartPos: ast.StartPos(stmt).Position(),
				EndPos:   ast.EndPos(stmt).Position(),
				Message:  fmt.Sprintf("unsupported statement type %T", stmt),
			}
		}
	}

	return &File{
		name:         name,
		node:         node,
		components:   components,
		configBlocks: configs,
	}, nil
}
