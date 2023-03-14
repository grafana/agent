package controller

import (
	"fmt"
	"strings"
	"sync"

	"github.com/grafana/agent/pkg/flow/internal/dag"
	"github.com/grafana/agent/pkg/flow/logging"
	"github.com/grafana/agent/pkg/flow/tracing"
	"github.com/grafana/agent/pkg/river/ast"
	"github.com/grafana/agent/pkg/river/diag"
	"github.com/grafana/agent/pkg/river/vm"
)

const (
	configNodeID = "configNode"

	loggingBlockID = "logging"
	tracingBlockID = "tracing"
	exportBlockID  = "export"
)

// ConfigNode is a controller node which manages agent configuration.
// The graph will always have _exactly one_ instance of ConfigNode, which will
// be used to contain the state of all config blocks.
type ConfigNode struct {
	globals ComponentGlobals

	mut          sync.RWMutex
	blocks       []*ast.BlockStmt    // Current River blocks to derive config from
	loggingArgs  logging.SinkOptions // Evaluated logging arguments for the config
	loggingBlock *ast.BlockStmt
	loggingEval  *vm.Evaluator
	tracingArgs  tracing.Options
	tracingBlock *ast.BlockStmt
	tracingEval  *vm.Evaluator
	exportBlocks map[string]*ast.BlockStmt
	exportEvals  map[string]*vm.Evaluator

	// Set to true for configuration blocks that were provided.
	foundLogging, foundTracing bool

	onExportsChanged func(map[string]any)
}

// ConfigBlockID returns the string name for a config block.
func ConfigBlockID(block *ast.BlockStmt) string {
	return strings.Join(BlockComponentID(block), ".")
}

var _ dag.Node = (*ConfigNode)(nil)

// NewConfigNode creates a new ConfigNode from an initial ast.BlockStmt.
// The underlying config isn't applied until Evaluate is called.
func NewConfigNode(blocks []*ast.BlockStmt, globals ComponentGlobals, onExportsChanged func(map[string]any)) (*ConfigNode, diag.Diagnostics) {
	var (
		blockMap = make(map[string]*ast.BlockStmt, len(blocks))
		diags    diag.Diagnostics

		loggingBlock ast.BlockStmt
		tracingBlock ast.BlockStmt

		exportBlocks = make(map[string]*ast.BlockStmt)
		exportEvals  = make(map[string]*vm.Evaluator)

		foundLogging = false
		foundTracing = false
	)

	for _, b := range blocks {
		var (
			name = strings.Join(b.Name, ".")
			id   = strings.Join(BlockComponentID(b), ".")
		)

		if orig, redefined := blockMap[id]; redefined {
			diags.Add(diag.Diagnostic{
				Severity: diag.SeverityLevelError,
				Message:  fmt.Sprintf("Config block %s already declared at %s", id, ast.StartPos(orig).Position()),
				StartPos: b.NamePos.Position(),
				EndPos:   b.NamePos.Add(len(id) - 1).Position(),
			})
			continue
		}

		switch name {
		case loggingBlockID:
			loggingBlock = *b
			foundLogging = true
		case tracingBlockID:
			tracingBlock = *b
			foundTracing = true
		case exportBlockID:
			if onExportsChanged == nil {
				diags.Add(diag.Diagnostic{
					Severity: diag.SeverityLevelError,
					Message:  "export blocks not allowed when not using modules",
					StartPos: ast.StartPos(b).Position(),
					EndPos:   ast.EndPos(b).Position(),
				})
			}

			exportBlocks[b.Label] = b
			exportEvals[b.Label] = vm.New(b.Body)
		}

		blockMap[id] = b
	}

	// Pre-populate arguments with their default values.
	var (
		loggerOptions = logging.DefaultSinkOptions
		tracerOptions = tracing.DefaultOptions
	)
	return &ConfigNode{
		globals: globals,

		blocks: blocks,

		loggingArgs:  loggerOptions,
		loggingBlock: &loggingBlock,
		loggingEval:  vm.New(loggingBlock.Body),

		tracingArgs:  tracerOptions,
		tracingBlock: &tracingBlock,
		tracingEval:  vm.New(tracingBlock.Body),

		exportBlocks: exportBlocks,
		exportEvals:  exportEvals,

		foundLogging: foundLogging,
		foundTracing: foundTracing,

		onExportsChanged: onExportsChanged,
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

	evals := []func(*vm.Scope) (*ast.BlockStmt, error){
		cn.evaluateLogging,
		cn.evaluateTracing,
		cn.evaluateExports,
	}
	for _, eval := range evals {
		if stmt, err := eval(scope); err != nil {
			return stmt, err
		}
	}
	return nil, nil
}

func (cn *ConfigNode) evaluateLogging(scope *vm.Scope) (*ast.BlockStmt, error) {
	if !cn.foundLogging {
		// Skip evaluating logging if the logging block wasn't provided.
		return nil, nil
	}

	// Evaluate logging block fields and store a copy.
	args := logging.DefaultSinkOptions
	if err := cn.loggingEval.Evaluate(scope, &args); err != nil {
		return cn.loggingBlock, fmt.Errorf("decoding River: %w", err)
	}
	cn.loggingArgs = args

	if err := cn.globals.LogSink.Update(cn.loggingArgs); err != nil {
		return cn.loggingBlock, fmt.Errorf("could not update logger: %w", err)
	}
	return nil, nil
}

func (cn *ConfigNode) evaluateTracing(scope *vm.Scope) (*ast.BlockStmt, error) {
	if !cn.foundTracing {
		// Skip evaluating tracing if the tracing block wasn't provided.
		return nil, nil
	}

	// Evaluate logging block fields and store a copy.
	args := tracing.DefaultOptions
	if err := cn.tracingEval.Evaluate(scope, &args); err != nil {
		return cn.tracingBlock, fmt.Errorf("decoding River: %w", err)
	}
	cn.tracingArgs = args

	t, ok := cn.globals.TraceProvider.(*tracing.Tracer)
	if ok {
		err := t.Update(cn.tracingArgs)
		if err != nil {
			return cn.tracingBlock, fmt.Errorf("could not update logger: %v", err)
		}
	}
	return nil, nil
}

type exportBlock struct {
	Value any `river:"value,attr"`
}

func (cn *ConfigNode) evaluateExports(scope *vm.Scope) (*ast.BlockStmt, error) {
	exports := make(map[string]any, len(cn.exportBlocks))

	for name, block := range cn.exportBlocks {
		eval := cn.exportEvals[name]

		var export exportBlock
		if err := eval.Evaluate(scope, &export); err != nil {
			return block, fmt.Errorf("decoding River: %w", err)
		}

		exports[name] = export.Value
	}

	if cn.onExportsChanged != nil {
		cn.onExportsChanged(exports)
	}
	return nil, nil
}

// LoggingArgs returns the arguments used to configure the logger.
func (cn *ConfigNode) LoggingArgs() logging.SinkOptions {
	cn.mut.RLock()
	defer cn.mut.RUnlock()
	return cn.loggingArgs
}

// TracingArgs returns the arguments used to configure the tracer.
func (cn *ConfigNode) TracingArgs() tracing.Options {
	cn.mut.RLock()
	defer cn.mut.RUnlock()
	return cn.tracingArgs
}
