// MOVE THIS SOMEWHERE
package controller

import (
	"fmt"
	"strings"

	"github.com/grafana/river/ast"
	"github.com/grafana/river/diag"
)

type CategorizedBlocks struct {
	Components    []*ast.BlockStmt
	Configs       []*ast.BlockStmt
	DeclareBlocks []*ast.BlockStmt
}

func CategorizeStatements(body ast.Body) (*CategorizedBlocks, *diag.Diagnostic) {
	// Look for predefined non-components blocks (i.e., logging), and store
	// everything else into a list of components.
	//
	// TODO(rfratto): should this code be brought into a helper somewhere? Maybe
	// in ast?

	var categorizedBlocks CategorizedBlocks

	for _, stmt := range body {
		switch stmt := stmt.(type) {
		case *ast.AttributeStmt:
			return nil, &diag.Diagnostic{
				Severity: diag.SeverityLevelError,
				StartPos: ast.StartPos(stmt.Name).Position(),
				EndPos:   ast.EndPos(stmt.Name).Position(),
				Message:  "unrecognized attribute " + stmt.Name.Name,
			}

		case *ast.BlockStmt:
			fullName := strings.Join(stmt.Name, ".")
			switch fullName {
			case "declare":
				categorizedBlocks.DeclareBlocks = append(categorizedBlocks.DeclareBlocks, stmt)
			case "logging", "tracing", "argument", "export":
				categorizedBlocks.Configs = append(categorizedBlocks.Configs, stmt)
			default:
				categorizedBlocks.Components = append(categorizedBlocks.Components, stmt)
			}

		default:
			return nil, &diag.Diagnostic{
				Severity: diag.SeverityLevelError,
				StartPos: ast.StartPos(stmt).Position(),
				EndPos:   ast.EndPos(stmt).Position(),
				Message:  fmt.Sprintf("unsupported statement type %T", stmt),
			}
		}
	}

	return &categorizedBlocks, nil
}
