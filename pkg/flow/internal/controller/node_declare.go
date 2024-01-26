package controller

import (
	"github.com/grafana/river/ast"
	"github.com/grafana/river/vm"
)

type DeclareNode struct {
	label         string
	nodeID        string
	componentName string
	// A declare content is static, it does not change during the lifetime of the node.
	declare *Declare
}

var _ BlockNode = (*DeclareNode)(nil)

// NewDeclareNode creates a new declare node with a content which will be loaded by custom components.
func NewDeclareNode(declare *Declare) *DeclareNode {
	return &DeclareNode{
		label:         declare.block.Label,
		nodeID:        BlockComponentID(declare.block).String(),
		componentName: declare.block.GetBlockName(),
		declare:       declare,
	}
}

func (cn *DeclareNode) Declare() *Declare {
	return cn.declare
}

// Evaluate does nothing for this node.
func (cn *DeclareNode) Evaluate(scope *vm.Scope) error {
	return nil
}

// Label returns the label of the block.
func (cn *DeclareNode) Label() string { return cn.label }

// Block implements BlockNode and returns the current block of the managed config node.
func (cn *DeclareNode) Block() *ast.BlockStmt {
	return cn.declare.block
}

// NodeID implements dag.Node and returns the unique ID for the config node.
func (cn *DeclareNode) NodeID() string { return cn.nodeID }
