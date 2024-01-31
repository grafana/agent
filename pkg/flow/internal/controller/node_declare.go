package controller

import (
	"fmt"
	"strings"
	"sync"

	"github.com/grafana/river/ast"
	"github.com/grafana/river/vm"
)

// DeclareNode represents a declare block in the DAG.
type DeclareNode struct {
	nodeID        string // ID of the node.
	componentName string // Name of the component being declared.

	mut   sync.RWMutex
	block *ast.BlockStmt // Definition of DeclareNode.
}

var _ BlockNode = (*DeclareNode)(nil)

// NewDeclareNode creates a new DeclareNode with its definition.
func NewDeclareNode(block *ast.BlockStmt) *DeclareNode {
	if fullName := strings.Join(block.Name, "."); fullName != "declare" {
		panic("controller: NewDeclareNode called with non-declare block.")
	}
	if block.Label == "" {
		panic("controller: NewDeclareNode called with a block without a label")
	}

	return &DeclareNode{
		nodeID:        fmt.Sprintf("declare.%s", block.Label),
		componentName: block.Label,
		block:         block,
	}
}

// Evaluate does nothing for DeclareNode.
func (cn *DeclareNode) Evaluate(scope *vm.Scope) error {
	return nil
}

// Block implements BlockNode and returns the current block of the managed config node.
func (cn *DeclareNode) Block() *ast.BlockStmt {
	cn.mut.RLock()
	defer cn.mut.RUnlock()
	return cn.block
}

// NodeID implements dag.Node and returns the unique ID for the config node.
func (cn *DeclareNode) NodeID() string { return cn.nodeID }

// ComponentName returns the name of the component being declared.
func (cn *DeclareNode) ComponentName() string { return cn.componentName }

// Definition implements [CustomComponent] and returns the body of the declare
// node.
func (cn *DeclareNode) Definition() (ast.Body, error) {
	cn.mut.RLock()
	defer cn.mut.RUnlock()

	return cn.block.Body, nil
}
