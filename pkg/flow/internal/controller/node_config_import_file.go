package controller

import (
	"context"
	"fmt"
	"path"
	"path/filepath"
	"reflect"
	"sync"
	"time"

	"github.com/go-kit/log"
	"github.com/grafana/agent/component"
	"github.com/grafana/agent/component/local/file"
	"github.com/grafana/agent/pkg/flow/logging/level"
	"github.com/grafana/agent/pkg/flow/tracing"
	"github.com/grafana/river/ast"
	"github.com/grafana/river/vm"
	"github.com/prometheus/client_golang/prometheus"
)

type ImportFileConfigNode struct {
	label                 string
	nodeID                string
	componentName         string
	globalID              string
	fileComponent         *file.Component
	managedOpts           component.Options
	registry              *prometheus.Registry
	onImportContentChange func(importID string, newContent string)

	mut      sync.RWMutex
	block    *ast.BlockStmt // Current River blocks to derive config from
	eval     *vm.Evaluator
	argument component.Arguments

	healthMut  sync.RWMutex
	evalHealth component.Health // Health of the last evaluate
	runHealth  component.Health // Health of running the component
}

var _ BlockNode = (*ImportFileConfigNode)(nil)
var _ RunnableNode = (*ImportFileConfigNode)(nil)

// NewImportFileConfigNode creates a new ImportFileConfigNode from an initial ast.BlockStmt.
// The underlying config isn't applied until Evaluate is called.
func NewImportFileConfigNode(block *ast.BlockStmt, globals ComponentGlobals, onImportContentChange func(importLabel string, newContent string)) *ImportFileConfigNode {
	var (
		id     = BlockComponentID(block)
		nodeID = id.String()
	)

	initHealth := component.Health{
		Health:     component.HealthTypeUnknown,
		Message:    "component created",
		UpdateTime: time.Now(),
	}
	globalID := nodeID
	if globals.ControllerID != "" {
		globalID = path.Join(globals.ControllerID, nodeID)
	}
	cn := &ImportFileConfigNode{
		globalID:              globalID,
		label:                 block.Label,
		nodeID:                BlockComponentID(block).String(),
		componentName:         block.GetBlockName(),
		onImportContentChange: onImportContentChange,

		block:      block,
		eval:       vm.New(block.Body),
		evalHealth: initHealth,
		runHealth:  initHealth,
	}
	cn.managedOpts = getImportManagedOptions(globals, cn)
	return cn
}

func getImportManagedOptions(globals ComponentGlobals, cn *ImportFileConfigNode) component.Options {
	cn.registry = prometheus.NewRegistry()
	return component.Options{
		ID:     cn.globalID,
		Logger: log.With(globals.Logger, "component", cn.globalID),
		Registerer: prometheus.WrapRegistererWith(prometheus.Labels{
			"component_id": cn.globalID,
		}, cn.registry),
		Tracer: tracing.WrapTracer(globals.TraceProvider, cn.globalID),

		DataPath: filepath.Join(globals.DataPath, cn.globalID),

		OnStateChange: cn.UpdateModulesContent,

		GetServiceData: func(name string) (interface{}, error) {
			return globals.GetServiceData(name)
		},
	}
}

type importFileConfigBlock struct {
	LocalFileArguments file.Arguments `river:",squash"`
}

// SetToDefault implements river.Defaulter.
func (a *importFileConfigBlock) SetToDefault() {
	a.LocalFileArguments = file.DefaultArguments
}

// Evaluate implements BlockNode and updates the arguments for the managed config block
// by re-evaluating its River block with the provided scope. The managed config block
// will be built the first time Evaluate is called.
//
// Evaluate will return an error if the River block cannot be evaluated or if
// decoding to arguments fails.
func (cn *ImportFileConfigNode) Evaluate(scope *vm.Scope) error {
	err := cn.evaluate(scope)

	switch err {
	case nil:
		cn.setEvalHealth(component.HealthTypeHealthy, "component evaluated")
	default:
		msg := fmt.Sprintf("component evaluation failed: %s", err)
		cn.setEvalHealth(component.HealthTypeUnhealthy, msg)
	}
	return err
}

func (cn *ImportFileConfigNode) setEvalHealth(t component.HealthType, msg string) {
	cn.healthMut.Lock()
	defer cn.healthMut.Unlock()

	cn.evalHealth = component.Health{
		Health:     t,
		Message:    msg,
		UpdateTime: time.Now(),
	}
}

func (cn *ImportFileConfigNode) evaluate(scope *vm.Scope) error {
	cn.mut.Lock()
	defer cn.mut.Unlock()

	var argument importFileConfigBlock
	if err := cn.eval.Evaluate(scope, &argument); err != nil {
		return fmt.Errorf("decoding River: %w", err)
	}
	if cn.fileComponent == nil {
		var err error
		cn.fileComponent, err = file.New(cn.managedOpts, argument.LocalFileArguments)
		if err != nil {
			return fmt.Errorf("creating file component: %w", err)
		}
		cn.argument = argument
	}

	if reflect.DeepEqual(cn.argument, argument) {
		// Ignore components which haven't changed. This reduces the cost of
		// calling evaluate for components where evaluation is expensive (e.g., if
		// re-evaluating requires re-starting some internal logic).
		return nil
	}

	// Update the existing managed component
	if err := cn.fileComponent.Update(argument); err != nil {
		return fmt.Errorf("updating component: %w", err)
	}

	return nil
}

// Run runs the managed component in the calling goroutine until ctx is
// canceled. Evaluate must have been called at least once without returning an
// error before calling Run.
//
// Run will immediately return ErrUnevaluated if Evaluate has never been called
// successfully. Otherwise, Run will return nil.
func (cn *ImportFileConfigNode) Run(ctx context.Context) error {
	cn.mut.RLock()
	managed := cn.fileComponent
	cn.mut.RUnlock()

	if managed == nil {
		return ErrUnevaluated
	}

	cn.setRunHealth(component.HealthTypeHealthy, "started component")
	err := cn.fileComponent.Run(ctx)

	var exitMsg string
	logger := cn.managedOpts.Logger
	if err != nil {
		level.Error(logger).Log("msg", "component exited with error", "err", err)
		exitMsg = fmt.Sprintf("component shut down with error: %s", err)
	} else {
		level.Info(logger).Log("msg", "component exited")
		exitMsg = "component shut down normally"
	}

	cn.setRunHealth(component.HealthTypeExited, exitMsg)
	return err
}

func (cn *ImportFileConfigNode) UpdateModulesContent(e component.Exports) {
	cn.onImportContentChange(cn.label, e.(file.Exports).Content.Value)
}

func (cn *ImportFileConfigNode) setRunHealth(t component.HealthType, msg string) {
	cn.healthMut.Lock()
	defer cn.healthMut.Unlock()

	cn.runHealth = component.Health{
		Health:     t,
		Message:    msg,
		UpdateTime: time.Now(),
	}
}

func (cn *ImportFileConfigNode) Label() string { return cn.label }

// Block implements BlockNode and returns the current block of the managed config node.
func (cn *ImportFileConfigNode) Block() *ast.BlockStmt {
	cn.mut.RLock()
	defer cn.mut.RUnlock()
	return cn.block
}

// NodeID implements dag.Node and returns the unique ID for the config node.
func (cn *ImportFileConfigNode) NodeID() string { return cn.nodeID }
