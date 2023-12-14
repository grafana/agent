package controller

import (
	"context"
	"fmt"
	"path"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/go-kit/log"
	"github.com/grafana/agent/component"
	"github.com/grafana/agent/component/local/file"
	"github.com/grafana/agent/pkg/flow/logging/level"
	"github.com/grafana/agent/pkg/flow/tracing"
	"github.com/grafana/river/ast"
	"github.com/grafana/river/parser"
	"github.com/grafana/river/vm"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/atomic"
)

type ImportFileConfigNode struct {
	id                ComponentID
	label             string
	nodeID            string
	componentName     string
	globalID          string
	fileComponent     *file.Component
	managedOpts       component.Options
	registry          *prometheus.Registry
	importedContent   map[string]string
	OnComponentUpdate func(cn NodeWithDependants) // Informs controller that we need to reevaluate

	mut                sync.RWMutex
	importedContentMut sync.RWMutex
	block              *ast.BlockStmt // Current River blocks to derive config from
	eval               *vm.Evaluator
	argument           component.Arguments
	lastUpdateTime     atomic.Time

	healthMut  sync.RWMutex
	evalHealth component.Health // Health of the last evaluate
	runHealth  component.Health // Health of running the component
}

var _ NodeWithDependants = (*ImportFileConfigNode)(nil)
var _ RunnableNode = (*ImportFileConfigNode)(nil)
var _ UINode = (*ComponentNode)(nil)

// NewImportFileConfigNode creates a new ImportFileConfigNode from an initial ast.BlockStmt.
// The underlying config isn't applied until Evaluate is called.
func NewImportFileConfigNode(block *ast.BlockStmt, globals ComponentGlobals) *ImportFileConfigNode {
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
		id:                id,
		globalID:          globalID,
		label:             block.Label,
		nodeID:            BlockComponentID(block).String(),
		componentName:     block.GetBlockName(),
		importedContent:   make(map[string]string),
		OnComponentUpdate: globals.OnComponentUpdate,
		block:             block,
		eval:              vm.New(block.Body),
		evalHealth:        initHealth,
		runHealth:         initHealth,
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

		OnStateChange: cn.onFileContentUpdate,

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

func (cn *ImportFileConfigNode) onFileContentUpdate(e component.Exports) {
	cn.importedContentMut.Lock()
	defer cn.importedContentMut.Unlock()
	fileContent := e.(file.Exports).Content.Value
	cn.importedContent = make(map[string]string)
	node, err := parser.ParseFile(cn.label, []byte(fileContent))
	logger := cn.managedOpts.Logger
	if err != nil {
		level.Error(logger).Log("msg", "failed to parse file on update", "err", err)
		return
	}
	for _, stmt := range node.Body {
		switch stmt := stmt.(type) {
		case *ast.BlockStmt:
			fullName := strings.Join(stmt.Name, ".")
			switch fullName {
			case "declare":
				if _, ok := cn.importedContent[stmt.Label]; ok {
					level.Error(logger).Log("msg", "declare block redefined", "name", stmt.Label)
					continue
				}
				cn.importedContent[stmt.Label] = string(fileContent[stmt.LCurlyPos.Position().Offset+1 : stmt.RCurlyPos.Position().Offset-1])
			default:
				level.Error(logger).Log("msg", "only declare blocks are allowed in a module", "forbidden", fullName)
			}
		default:
			level.Error(logger).Log("msg", "only declare blocks are allowed in a module")
		}
	}
	cn.lastUpdateTime.Store(time.Now())
	cn.OnComponentUpdate(cn)
}

func (cn *ImportFileConfigNode) ModuleContent(module string) (string, error) {
	cn.importedContentMut.Lock()
	defer cn.importedContentMut.Unlock()
	if content, ok := cn.importedContent[module]; ok {
		return content, nil
	}
	return "", fmt.Errorf("module %s not found in imported node %s", module, cn.label)
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

// This node has no exports.
func (cn *ImportFileConfigNode) Exports() component.Exports {
	return nil
}

func (cn *ImportFileConfigNode) ID() ComponentID { return cn.id }

func (cn *ImportFileConfigNode) LastUpdateTime() time.Time {
	return cn.lastUpdateTime.Load()
}

// Arguments returns the current arguments of the managed component.
func (cn *ImportFileConfigNode) Arguments() component.Arguments {
	cn.mut.RLock()
	defer cn.mut.RUnlock()
	return cn.argument
}

// Component returns the instance of the managed component. Component may be
// nil if the ComponentNode has not been successfully evaluated yet.
func (cn *ImportFileConfigNode) Component() component.Component {
	cn.mut.RLock()
	defer cn.mut.RUnlock()
	return cn.fileComponent
}

// CurrentHealth returns the current health of the ComponentNode.
//
// The health of a ComponentNode is determined by combining:
//
//  1. Health from the call to Run().
//  2. Health from the last call to Evaluate().
//  3. Health reported from the component.
func (cn *ImportFileConfigNode) CurrentHealth() component.Health {
	cn.healthMut.RLock()
	defer cn.healthMut.RUnlock()
	return component.LeastHealthy(cn.runHealth, cn.evalHealth, cn.fileComponent.CurrentHealth())
}

// FileComponent does not have DebugInfo
func (cn *ImportFileConfigNode) DebugInfo() interface{} {
	return nil
}

// This component does not manage modules.
func (cn *ImportFileConfigNode) ModuleIDs() []string {
	return nil
}

// BlockName returns the name of the block.
func (cn *ImportFileConfigNode) BlockName() string {
	return cn.componentName
}
