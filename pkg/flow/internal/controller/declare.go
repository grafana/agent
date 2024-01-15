package controller

import "github.com/grafana/river/ast"

// Declare represents the content of a declare block as AST and as plain string.
type Declare struct {
	Block   *ast.BlockStmt
	Content string
}
