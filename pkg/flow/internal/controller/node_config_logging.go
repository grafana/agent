package controller

import (
	"fmt"
	"sync"
	"time"

	"github.com/go-kit/log"
	"github.com/grafana/agent/pkg/flow/logging"
	"github.com/grafana/river/ast"
	"github.com/grafana/river/vm"
	"go.uber.org/atomic"
)

var _ BlockNode = (*LoggingConfigNode)(nil)

type LoggingConfigNode struct {
	nodeID         string
	componentName  string
	l              log.Logger
	lastUpdateTime atomic.Time

	mut   sync.RWMutex
	block *ast.BlockStmt // Current River blocks to derive config from
	eval  *vm.Evaluator
}

// NewLoggingConfigNode creates a new LoggingConfigNode from an initial ast.BlockStmt.
// The underlying config isn't applied until Evaluate is called.
func NewLoggingConfigNode(block *ast.BlockStmt, globals ComponentGlobals) *LoggingConfigNode {
	return &LoggingConfigNode{
		nodeID:        BlockComponentID(block).String(),
		componentName: block.GetBlockName(),
		l:             globals.Logger,

		block: block,
		eval:  vm.New(block.Body),
	}
}

// NewDefaultLoggingConfigNode creates a new LoggingConfigNode with nil block and eval.
// This will force evaluate to use the default logging options for this node.
func NewDefaultLoggingConfigNode(globals ComponentGlobals) *LoggingConfigNode {
	return &LoggingConfigNode{
		nodeID:        loggingBlockID,
		componentName: loggingBlockID,
		l:             globals.Logger,

		block: nil,
		eval:  nil,
	}
}

// Evaluate implements BlockNode and updates the arguments for the managed config block
// by re-evaluating its River block with the provided scope. The managed config block
// will be built the first time Evaluate is called.
//
// Evaluate will return an error if the River block cannot be evaluated or if
// decoding to arguments fails.
func (cn *LoggingConfigNode) Evaluate(scope *vm.Scope) error {
	cn.mut.RLock()
	defer cn.mut.RUnlock()
	args := logging.DefaultOptions
	if cn.eval != nil {
		if err := cn.eval.Evaluate(scope, &args); err != nil {
			return fmt.Errorf("decoding River: %w", err)
		}
	}

	if err := cn.l.(*logging.Logger).Update(args); err != nil {
		return fmt.Errorf("could not update logger: %w", err)
	}
	cn.lastUpdateTime.Store(time.Now())
	return nil
}

// Block implements BlockNode and returns the current block of the managed config node.
func (cn *LoggingConfigNode) Block() *ast.BlockStmt {
	cn.mut.RLock()
	defer cn.mut.RUnlock()
	return cn.block
}

// NodeID implements dag.Node and returns the unique ID for the config node.
func (cn *LoggingConfigNode) NodeID() string { return cn.nodeID }

// LastUpdateTime returns the time corresponding to the last time where the node was updated.
func (cn *LoggingConfigNode) LastUpdateTime() time.Time {
	return cn.lastUpdateTime.Load()
}
