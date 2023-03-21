package controller

import (
	"fmt"

	"github.com/grafana/agent/pkg/flow/logging"
	"github.com/grafana/agent/pkg/river/ast"
	"github.com/grafana/agent/pkg/river/diag"
	"github.com/grafana/agent/pkg/river/vm"
)

type LoggingConfigNode struct {
	configNode sharedBlockNode
	logSink    *logging.Sink // Sink used for Logging.
}

// NewLoggingConfigNode creates a new LoggingConfigNode from an initial ast.BlockStmt.
// The underlying config isn't applied until Evaluate is called.
func NewLoggingConfigNode(block *ast.BlockStmt, globals ComponentGlobals, isInModule bool) (*LoggingConfigNode, diag.Diagnostics) {
	var diags diag.Diagnostics

	if isInModule {
		diags.Add(diag.Diagnostic{
			Severity: diag.SeverityLevelError,
			Message:  "logging block not allowed inside a module",
			StartPos: ast.StartPos(block).Position(),
			EndPos:   ast.EndPos(block).Position(),
		})
	}

	return &LoggingConfigNode{
		configNode: sharedBlockNode{
			label:         block.Label,
			nodeID:        BlockComponentID(block).String(),
			componentName: block.GetBlockName(),

			block: block,
			eval:  vm.New(block.Body),
		},
		logSink: globals.LogSink,
	}, diags
}

// Evaluate implements BlockNode and updates the arguments for the managed config block
// by re-evaluating its River block with the provided scope. The managed config block
// will be built the first time Evaluate is called.
//
// Evaluate will return an error if the River block cannot be evaluated or if
// decoding to arguments fails.
func (cn *LoggingConfigNode) Evaluate(scope *vm.Scope) error {
	args := logging.DefaultSinkOptions
	if err := cn.configNode.Evaluate(scope, &args); err != nil {
		return fmt.Errorf("decoding River: %w", err)
	}

	if err := cn.logSink.Update(args); err != nil {
		return fmt.Errorf("could not update logger: %w", err)
	}

	return nil
}

// ComponentName implements BlockNode and returns the component's type, i.e. `local.file.test` returns `local.file`.
func (cn *LoggingConfigNode) ComponentName() string { return cn.configNode.componentName }

// Label implements BlockNode and returns the label for the block or "" if none was specified.
func (cn *LoggingConfigNode) Label() string { return cn.configNode.label }

// Block implements BlockNode and returns the current block of the managed config node.
func (cn *LoggingConfigNode) Block() *ast.BlockStmt { return cn.configNode.Block() }

// NodeID implements dag.Node and returns the unique ID for the config node.
func (cn *LoggingConfigNode) NodeID() string { return cn.configNode.nodeID }
