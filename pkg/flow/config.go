package flow

import (
	"fmt"
	"strings"

	"github.com/grafana/agent/pkg/flow/logging"
	"github.com/grafana/agent/pkg/river/ast"
	"github.com/grafana/agent/pkg/river/diag"
	"github.com/grafana/agent/pkg/river/parser"
	"github.com/grafana/agent/pkg/river/vm"
)

// File holds the contents of a parsed Flow file.
type File struct {
	Name string    // File name given to ReadFile.
	Node *ast.File // Raw File node.

	Logging logging.Options

	// Components holds the list of raw River AST blocks describing components.
	// The Flow controller can interpret them.
	Components []*ast.BlockStmt
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
		loggerBlock *ast.BlockStmt
		components  []*ast.BlockStmt
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
				loggerBlock = stmt
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

	loggingOpts := logging.DefaultOptions
	if loggerBlock != nil {
		if err := vm.New(loggerBlock.Body).Evaluate(nil, &loggingOpts); err != nil {
			return nil, err
		}
	}

	return &File{
		Name:       name,
		Node:       node,
		Logging:    loggingOpts,
		Components: components,
	}, nil
}
