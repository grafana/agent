package controller

import (
	"strings"
	"sync"

	"github.com/grafana/river/ast"
	"github.com/grafana/river/vm"
)

// DeclareNode represents a declare block in the DAG.
type DeclareNode struct {
	label         string
	nodeID        string
	componentName string
	mut           sync.RWMutex
	block         *ast.BlockStmt
}

var _ BlockNode = (*DeclareNode)(nil)

const declareType = "declare"

// NewDeclareNode creates a new declare node with a content which will be loaded by custom components.
func NewDeclareNode(block *ast.BlockStmt) *DeclareNode {
	return &DeclareNode{
		label:         block.Label,
		nodeID:        BlockComponentID(block).String(),
		componentName: block.GetBlockName(),
		block:         block,
	}
}

// Evaluate does nothing for this node.
func (cn *DeclareNode) Evaluate(scope *vm.Scope) error {
	return nil
}

// Label returns the label of the block.
func (cn *DeclareNode) Label() string { return cn.label }

// Block implements BlockNode and returns the current block of the managed config node.
func (cn *DeclareNode) Block() *ast.BlockStmt {
	return cn.block
}

// NodeID implements dag.Node and returns the unique ID for the config node.
func (cn *DeclareNode) NodeID() string { return cn.nodeID }

// UpdateBlock updates the managed River block.
//
// UpdateBlock will panic if the block does not match the component ID of the
// DeclareNode.
func (cn *DeclareNode) UpdateBlock(b *ast.BlockStmt) {
	if !BlockComponentID(b).Equals(strings.Split(cn.nodeID, ".")) {
		panic("UpdateBlock called with an River block with a different ID")
	}

	cn.mut.Lock()
	defer cn.mut.Unlock()
	cn.block = b
}
