package controller

import "github.com/grafana/river/ast"

// Should this be defined somewhere else?
type Declare struct {
	Block   *ast.BlockStmt
	Content string
}
