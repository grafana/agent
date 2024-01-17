package controller

import "github.com/grafana/river/ast"

// Declare represents the content of a declare block as AST and as plain string.
type Declare struct {
	block   *ast.BlockStmt
	content string
}

// NewDeclare creates a new Declare from its AST and its plain string content.
func NewDeclare(block *ast.BlockStmt, content string) *Declare {
	return &Declare{block: block, content: content}
}
