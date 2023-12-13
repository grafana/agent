package controller

import (
	"sync"

	"github.com/grafana/river/ast"
	"github.com/grafana/river/vm"
)

type DeclareNode struct {
	label         string
	nodeID        string
	componentName string
	content       string

	mut   sync.RWMutex
	block *ast.BlockStmt // Current River blocks to derive config from
}

var _ BlockNode = (*DeclareNode)(nil)

// NewDeclareNode creates a new DeclareNode from an initial ast.BlockStmt.
// The underlying config isn't applied until Evaluate is called.
func NewDeclareNode(block *ast.BlockStmt, content string) *DeclareNode {
	return &DeclareNode{
		label:         block.Label,
		nodeID:        BlockComponentID(block).String(),
		componentName: block.GetBlockName(),
		content:       content,

		block: block,
	}
}

func (cn *DeclareNode) ModuleContent() string {
	return cn.content
}

// Evaluate implements BlockNode and updates the arguments for the managed config block
// by re-evaluating its River block with the provided scope. The managed config block
// will be built the first time Evaluate is called.
//
// Evaluate will return an error if the River block cannot be evaluated or if
// decoding to arguments fails.
func (cn *DeclareNode) Evaluate(scope *vm.Scope) error {
	return nil
}

func (cn *DeclareNode) Label() string { return cn.label }

// Block implements BlockNode and returns the current block of the managed config node.
func (cn *DeclareNode) Block() *ast.BlockStmt {
	cn.mut.RLock()
	defer cn.mut.RUnlock()
	return cn.block
}

// NodeID implements dag.Node and returns the unique ID for the config node.
func (cn *DeclareNode) NodeID() string { return cn.nodeID }
