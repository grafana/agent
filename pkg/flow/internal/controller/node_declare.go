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
	block *ast.BlockStmt
}

var _ BlockNode = (*DeclareNode)(nil)

// NewDeclareNode creates a new declare node with a content which will be loaded by declare component node.
func NewDeclareNode(block *ast.BlockStmt, content string) *DeclareNode {
	return &DeclareNode{
		label:         block.Label,
		nodeID:        BlockComponentID(block).String(),
		componentName: block.GetBlockName(),
		content:       content,

		block: block,
	}
}

func (cn *DeclareNode) ModuleContent() (string, error) {
	cn.mut.Lock()
	defer cn.mut.Unlock()
	return cn.content, nil
}

// Evaluate does nothing for this node.
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
