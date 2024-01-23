package schema

import (
	"fmt"
	"strings"

	"github.com/grafana/river/ast"
	"github.com/grafana/river/diag"
)

type Verifier struct {
	schema Json
}

func (v *Verifier) Verify(dag *ast.File) *diag.Diagnostic {
	var (
		components []*ast.BlockStmt
		configs    []*ast.BlockStmt
		services   []*ast.BlockStmt
	)

	for _, stmt := range dag.Body {
		switch stmt := stmt.(type) {
		case *ast.AttributeStmt:
			return &diag.Diagnostic{
				Severity: diag.SeverityLevelError,
				StartPos: ast.StartPos(stmt.Name).Position(),
				EndPos:   ast.EndPos(stmt.Name).Position(),
				Message:  "unrecognized attribute " + stmt.Name.Name,
			}

		case *ast.BlockStmt:
			fullName := strings.Join(stmt.Name, ".")

			switch {
			case v.isConfigBlock(fullName):
				configs = append(configs, stmt)
			case v.isService(fullName):
				services = append(services, stmt)
			case v.isComponent(fullName):
				components = append(components, stmt)
			default:
				return &diag.Diagnostic{
					Severity: diag.SeverityLevelError,
					StartPos: ast.StartPos(stmt).Position(),
					EndPos:   ast.EndPos(stmt).Position(),
					Message:  fmt.Sprintf("unknown statement %T", stmt),
				}
			}

		default:
			return &diag.Diagnostic{
				Severity: diag.SeverityLevelError,
				StartPos: ast.StartPos(stmt).Position(),
				EndPos:   ast.EndPos(stmt).Position(),
				Message:  fmt.Sprintf("unsupported statement type %T", stmt),
			}
		}
	}

	// Verify each block.

	return nil
}

func (v *Verifier) isConfigBlock(name string) bool {
	for _, cb := range v.schema.ConfigBlocks {
		if cb.Name == name {
			return true
		}
	}
	return false
}

func (v *Verifier) isService(name string) bool {
	for _, sb := range v.schema.Services {
		if sb.Name == name {
			return true
		}
	}
	return false
}

func (v *Verifier) isComponent(name string) bool {
	for _, cb := range v.schema.Components {
		if cb.Name == name {
			return true
		}
	}
	return false
}
