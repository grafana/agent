package flow

import (
	"fmt"
	"strings"

	"github.com/grafana/agent/pkg/river/ast"
	"github.com/grafana/agent/pkg/river/diag"
	"github.com/grafana/agent/pkg/river/parser"
)

// Source holds the contents of a parsed Flow source.
type Source struct {
	// components holds the list of raw River AST blocks describing components.
	// The Flow controller can interpret them.
	components   []*ast.BlockStmt
	configBlocks []*ast.BlockStmt
}

// ParseSource parses the River contents specified by bb into a Source. name
// should be the name of the source used for reporting errors.
func ParseSource(name string, bb []byte) (*Source, error) {
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

	return &Source{
		components:   components,
		configBlocks: configs,
	}, nil
}
