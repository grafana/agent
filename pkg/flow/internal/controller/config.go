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
	loggingBlockID = "logging"
	tracingBlockID = "tracing"
	exportBlockID  = "export"
)

// ConfigNode is a controller node which manages agent configuration.
// The graph will always have _exactly one_ instance of ConfigNode, which will
// be used to contain the state of all config blocks.
type ConfigNode struct {
	label         string
	nodeID        string
	componentName string
	globals       ComponentGlobals

	mut   sync.RWMutex
	block *ast.BlockStmt // Current River blocks to derive config from
	eval  *vm.Evaluator
}

// ConfigBlockID returns the string name for a config block.
func ConfigBlockID(block *ast.BlockStmt) string {
	return strings.Join(BlockComponentID(block), ".")
}

var _ dag.Node = (*ConfigNode)(nil)

// NewConfigNode creates a new ConfigNode from an initial ast.BlockStmt.
// The underlying config isn't applied until Evaluate is called.
func NewConfigNode(block *ast.BlockStmt, globals ComponentGlobals, isInModule bool) (*ConfigNode, diag.Diagnostics) {
	var (
		diags diag.Diagnostics

		name   = strings.Join(block.Name, ".")
		nodeID = BlockComponentID(block).String()
	)

	switch name {
	case loggingBlockID:
		if isInModule {
			diags.Add(diag.Diagnostic{
				Severity: diag.SeverityLevelError,
				Message:  "logging block not allowed inside a module",
				StartPos: ast.StartPos(block).Position(),
				EndPos:   ast.EndPos(block).Position(),
			})
		}
	case tracingBlockID:
		if isInModule {
			diags.Add(diag.Diagnostic{
				Severity: diag.SeverityLevelError,
				Message:  "tracing block not allowed inside a module",
				StartPos: ast.StartPos(block).Position(),
				EndPos:   ast.EndPos(block).Position(),
			})
		}
	case exportBlockID:
		if !isInModule {
			diags.Add(diag.Diagnostic{
				Severity: diag.SeverityLevelError,
				Message:  "export blocks only allowed inside a module",
				StartPos: ast.StartPos(block).Position(),
				EndPos:   ast.EndPos(block).Position(),
			})
		}
	}

	return &ConfigNode{
		label:         block.Label,
		nodeID:        nodeID,
		componentName: name,
		globals:       globals,

		block: block,
		eval:  vm.New(block.Body),
	}, diags
}

// NodeID implements dag.Node and returns the unique ID for the config node.
func (cn *ConfigNode) NodeID() string { return cn.nodeID }

// Evaluate updates the config block by re-evaluating its River block with the
// provided scope. The config will be built the first time Evaluate is called.
//
// Evaluate will return an error if the River block cannot be evaluated or if
// decoding to arguments fails.
func (cn *ConfigNode) Evaluate(scope *vm.Scope) error {
	cn.mut.Lock()
	defer cn.mut.Unlock()

	switch cn.componentName {
	case loggingBlockID:
		return cn.evaluateLogging(scope)
	case tracingBlockID:
		return cn.evaluateTracing(scope)
	case exportBlockID:
		return cn.evaluateExports(scope)
	}

	return nil
}

func (cn *ConfigNode) evaluateLogging(scope *vm.Scope) error {
	args := logging.DefaultSinkOptions
	if err := cn.eval.Evaluate(scope, &args); err != nil {
		return fmt.Errorf("decoding River: %w", err)
	}

	if err := cn.globals.LogSink.Update(args); err != nil {
		return fmt.Errorf("could not update logger: %w", err)
	}

	return nil
}

func (cn *ConfigNode) evaluateTracing(scope *vm.Scope) error {
	args := tracing.DefaultOptions
	if err := cn.eval.Evaluate(scope, &args); err != nil {
		return fmt.Errorf("decoding River: %w", err)
	}

	t, ok := cn.globals.TraceProvider.(*tracing.Tracer)
	if ok {
		err := t.Update(args)
		if err != nil {
			return fmt.Errorf("could not update logger: %v", err)
		}
	}
	return nil
}

type exportBlock struct {
	Value any `river:"value,attr"`
}

func (cn *ConfigNode) evaluateExports(scope *vm.Scope) error {
	exports := make(map[string]any, 1)

	var export exportBlock
	if err := cn.eval.Evaluate(scope, &export); err != nil {
		return fmt.Errorf("decoding River: %w", err)
	}

	exports[cn.label] = export.Value

	if cn.globals.OnExportsChange != nil {
		cn.globals.OnExportsChange(exports)
	}
	return nil
}
