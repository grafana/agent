package controller

import (
	"fmt"
	"strings"
	"sync"

	"github.com/go-kit/log"
	"github.com/grafana/agent/pkg/flow/logging"
	"github.com/grafana/agent/pkg/river/ast"
	"github.com/grafana/agent/pkg/river/diag"
	"github.com/grafana/agent/pkg/river/vm"
)

const (
	configNodeID = "configNode"

	loggingBlockID = "logging"
)

// ConfigNode is a controller node which manages agent configuration.
type ConfigNode struct {
	mut          sync.RWMutex
	blocks       []*ast.BlockStmt // Current River blocks to derive config from
	logger       log.Logger
	loggingArgs  logging.Options // Evaluated logging arguments for the config
	loggingBlock *ast.BlockStmt
	loggingEval  *vm.Evaluator
}

// ConfigBlockID returns the string name for a config block.
func ConfigBlockID(block *ast.BlockStmt) string {
	return strings.Join(block.Name, ".")
}

// NewConfigNode creates a new ConfigNode from an initial ast.BlockStmt.
// The underlying config isn't applied until Evaluate is called.
func NewConfigNode(blocks []*ast.BlockStmt, l log.Logger) (*ConfigNode, diag.Diagnostics) {
	var (
		blockMap = make(map[string]*ast.BlockStmt, len(blocks))
		diags    diag.Diagnostics

		loggingBlock ast.BlockStmt
	)

	for _, b := range blocks {
		id := ConfigBlockID(b)
		if orig, redefined := blockMap[id]; redefined {
			diags.Add(diag.Diagnostic{
				Severity: diag.SeverityLevelError,
				Message:  fmt.Sprintf("Config block %s already declared at %s", id, ast.StartPos(orig).Position()),
				StartPos: b.NamePos.Position(),
				EndPos:   b.NamePos.Add(len(id) - 1).Position(),
			})
			continue
		}

		switch id {
		case loggingBlockID:
			loggingBlock = *b
		}

		blockMap[id] = b
	}

	loggerOptions := logging.DefaultOptions // Pre-populate arguments with their zero values.
	return &ConfigNode{
		blocks:       blocks,
		logger:       l,
		loggingArgs:  loggerOptions,
		loggingBlock: &loggingBlock,
		loggingEval:  vm.New(loggingBlock.Body),
	}, diags
}

// NodeID implements dag.Node and returns the unique ID for the config node.
func (cn *ConfigNode) NodeID() string { return configNodeID }

// Evaluate updates the config block by re-evaluating its River block with the
// provided scope. The config will be built the first time Evaluate is called.
//
// Evaluate will return an error if the River block cannot be evaluated or if
// decoding to arguments fails.
func (cn *ConfigNode) Evaluate(scope *vm.Scope) (*ast.BlockStmt, error) {
	cn.mut.Lock()
	defer cn.mut.Unlock()

	// Evaluate logging block fields and store a copy.
	args := logging.DefaultOptions
	if err := cn.loggingEval.Evaluate(scope, &args); err != nil {
		return cn.loggingBlock, fmt.Errorf("decoding River: %w", err)
	}
	cn.loggingArgs = args

	l, ok := cn.logger.(*logging.Logger)
	if ok {
		err := l.Update(cn.loggingArgs)
		if err != nil {
			return cn.loggingBlock, fmt.Errorf("could not update logger: %v", err)
		}
	}
	return nil, nil
}

// LoggingArgs returns the arguments used to configure the logger.
func (cn *ConfigNode) LoggingArgs() logging.Options {
	cn.mut.RLock()
	defer cn.mut.RUnlock()
	return cn.loggingArgs
}
