// Package parser implements utilities for parsing River configuration files.
package parser

import (
	"fmt"

	"github.com/grafana/agent/pkg/river/ast"
	"github.com/grafana/agent/pkg/river/token"
)

// Error is an error encountered during parsing.
type Error struct {
	Position token.Position
	Message  string
}

// Error implements error.
func (e Error) Error() string {
	return fmt.Sprintf("%s: %s", e.Position, e.Message)
}

// ErrorList is a list of Error.
type ErrorList []*Error

// Add appends a new error into the ErrorList.
func (l *ErrorList) Add(e *Error) { *l = append(*l, e) }

// Error implements error.
func (l ErrorList) Error() string {
	switch len(l) {
	case 0:
		return "no errors"
	case 1:
		return l[0].Error()
	}
	return fmt.Sprintf("%s (and %d more errors)", l[0], len(l)-1)
}

// ParseFile parses an entire River configuration file. The data parameter
// should hold the file contents to parse, while the filename parameter is used
// for reporting errors.
//
// If an error was encountered during parsing, the returned AST will be nil and
// err will be an ErrorList with all the errors encountered during parsing.
func ParseFile(filename string, data []byte) (*ast.File, error) {
	p := newParser(filename, data)

	f := p.ParseFile()
	if len(p.errors) > 0 {
		return nil, p.errors
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
	if len(p.errors) > 0 {
		return nil, p.errors
	}
	return e, nil
}
