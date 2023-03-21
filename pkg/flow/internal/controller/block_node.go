package controller

import (
	"sync"

	"github.com/grafana/agent/pkg/flow/internal/dag"
	"github.com/grafana/agent/pkg/river/ast"
	"github.com/grafana/agent/pkg/river/vm"
)

// BlockNode is a node in the DAG which manages a River block
// and can be evaluated.
type BlockNode interface {
	dag.Node

	// Label returns the label for the block or "" if none was specified.
	Label() string

	// ComponentName returns the component's type, i.e. `local.file.test` returns `local.file`.
	ComponentName() string

	// Block returns the current block of the managed config node.
	Block() *ast.BlockStmt

	// Evaluate updates the arguments for the managed component
	// by re-evaluating its River block with the provided scope. The managed component
	// will be built the first time Evaluate is called.
	//
	// Evaluate will return an error if the River block cannot be evaluated or if
	// decoding to arguments fails.
	Evaluate(scope *vm.Scope) error
}

// sharedBlockNode contains all shared functionality for a block node
type sharedBlockNode struct {
	label         string
	nodeID        string
	componentName string

	mut   sync.RWMutex
	block *ast.BlockStmt // Current River blocks to derive config from
	eval  *vm.Evaluator
}

// Block returns the current block of the managed config node.
func (cn *sharedBlockNode) Block() *ast.BlockStmt {
	cn.mut.RLock()
	defer cn.mut.RUnlock()
	return cn.block
}

// Evaluate executes eval's Evaluate function.
func (cn *sharedBlockNode) Evaluate(scope *vm.Scope, args interface{}) (err error) {
	cn.mut.RLock()
	defer cn.mut.RUnlock()
	return cn.eval.Evaluate(scope, &args)
}
