package controller

import (
	"fmt"
	"strings"
	"sync"

	"github.com/grafana/river/ast"
	"github.com/grafana/river/vm"
)

type ArgumentConfigNode struct {
	label         string
	nodeID        string
	componentName string

	mut          sync.RWMutex
	block        *ast.BlockStmt // Current River blocks to derive config from
	eval         *vm.Evaluator
	defaultValue any
	optional     bool
}

var _ BlockNode = (*ArgumentConfigNode)(nil)

// NewArgumentConfigNode creates a new ArgumentConfigNode from an initial ast.BlockStmt.
// The underlying config isn't applied until Evaluate is called.
func NewArgumentConfigNode(block *ast.BlockStmt, globals ComponentGlobals) *ArgumentConfigNode {
	return &ArgumentConfigNode{
		label:         block.Label,
		nodeID:        BlockComponentID(block).String(),
		componentName: block.GetBlockName(),

		block: block,
		eval:  vm.New(block.Body),
	}
}

type argumentBlock struct {
	Optional bool   `river:"optional,attr,optional"`
	Default  any    `river:"default,attr,optional"`
	Comment  string `river:"comment,attr,optional"`
}

// Evaluate implements BlockNode and updates the arguments for the managed config block
// by re-evaluating its River block with the provided scope. The managed config block
// will be built the first time Evaluate is called.
//
// Evaluate will return an error if the River block cannot be evaluated or if
// decoding to arguments fails.
func (cn *ArgumentConfigNode) Evaluate(scope *vm.Scope) error {
	cn.mut.Lock()
	defer cn.mut.Unlock()

	var argument argumentBlock
	if err := cn.eval.Evaluate(scope, &argument); err != nil {
		return fmt.Errorf("decoding River: %w", err)
	}

	cn.defaultValue = argument.Default
	cn.optional = argument.Optional

	return nil
}

func (cn *ArgumentConfigNode) Optional() bool {
	cn.mut.RLock()
	defer cn.mut.RUnlock()
	return cn.optional
}

func (cn *ArgumentConfigNode) Default() any {
	cn.mut.RLock()
	defer cn.mut.RUnlock()
	return cn.defaultValue
}

func (cn *ArgumentConfigNode) Label() string { return cn.label }

// Block implements BlockNode and returns the current block of the managed config node.
func (cn *ArgumentConfigNode) Block() *ast.BlockStmt {
	cn.mut.RLock()
	defer cn.mut.RUnlock()
	return cn.block
}

// NodeID implements dag.Node and returns the unique ID for the config node.
func (cn *ArgumentConfigNode) NodeID() string { return cn.nodeID }

// UpdateBlock updates the River block used to construct arguments.
// The new block isn't used until the next time Evaluate is invoked.
//
// UpdateBlock will panic if the block does not match the component ID of the
// ArgumentConfigNode.
func (cn *ArgumentConfigNode) UpdateBlock(b *ast.BlockStmt) {
	if !BlockComponentID(b).Equals(strings.Split(cn.nodeID, ".")) {
		panic("UpdateBlock called with an River block with a different ID")
	}

	cn.mut.Lock()
	defer cn.mut.Unlock()
	cn.block = b
	cn.eval = vm.New(b.Body)
}
