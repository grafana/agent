package controller

import "github.com/grafana/river/ast"

// Declare represents the content of a declare block as AST and as plain string.
type Declare struct {
	block *ast.BlockStmt
	// TODO: we would not need this content field if the content of the block was saved in ast.BlockStmt when parsing.
	// Not only it looks redundant but it allows discrepancies between the block and the content.
	content string
}

// NewDeclare creates a new Declare from its AST and its plain string content.
func NewDeclare(block *ast.BlockStmt, content string) *Declare {
	return &Declare{block: block, content: content}
}
