package controller

import (
	"github.com/grafana/river/ast"
	"github.com/grafana/river/vm"
)

// TODO: This struct is an empty shell for now, its implementation will come via another PR.

type ImportConfigNode struct {
	label  string
	nodeID string
	block  *ast.BlockStmt
}

var _ BlockNode = (*ImportConfigNode)(nil)

func (in *ImportConfigNode) Evaluate(scope *vm.Scope) error {
	return nil
}

// NodeID implements dag.Node and returns the unique ID for this node. The
// NodeID is the string representation of the component's ID from its River
// block.
func (in *ImportConfigNode) NodeID() string { return in.nodeID }

// ImportedDeclares returns all declare blocks that it imported.
func (in *ImportConfigNode) ImportedDeclares() map[string]ast.Body {
	return nil
}

// ImportConfigNodesChildren returns the ImportConfigNodesChildren of this ImportConfigNode.
func (in *ImportConfigNode) ImportConfigNodesChildren() map[string]*ImportConfigNode {
	return nil
}

// Block implements BlockNode and returns the current block of the managed config node.
func (in *ImportConfigNode) Block() *ast.BlockStmt {
	return in.block
}
