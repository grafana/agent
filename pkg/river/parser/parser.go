// Package parser implements utilities for parsing River configuration files.
package parser

import (
	"github.com/grafana/agent/pkg/river/ast"
)

// ParseFile parses an entire River configuration file. The data parameter
// should hold the file contents to parse, while the filename parameter is used
// for reporting errors.
//
// If an error was encountered during parsing, the returned AST will be nil and
// err will be an diag.Diagnostics all the errors encountered during parsing.
func ParseFile(filename string, data []byte) (*ast.File, error) {
	p := newParser(filename, data)

	f := p.ParseFile()
	if len(p.diags) > 0 {
		return nil, p.diags
	}
	return f, nil
}

// ParseExpression parses a single River expression from expr.
//
// If an error was encountered during parsing, the returned expression will be
// nil and err will be an ErrorList with all the errors encountered during
// parsing.
func ParseExpression(expr string) (ast.Expr, error) {
	p := newParser("", []byte(expr))

	e := p.ParseExpression()
	if len(p.diags) > 0 {
		return nil, p.diags
	}
	return e, nil
}
