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
	"github.com/grafana/agent/component/module"
	"github.com/grafana/agent/pkg/flow/logging/level"
	"github.com/grafana/agent/pkg/flow/tracing"
	"github.com/grafana/river/ast"
	"github.com/grafana/river/vm"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/atomic"
)

// DeclareComponentNode is a controller node which manages a module.
//
// DeclareComponentNode manages the underlying module and caches its current
// arguments and exports.
type DeclareComponentNode struct {
	id                         ComponentID
	globalID                   string
	label                      string
	componentName              string
	importLabel                string
	declareLabel               string
	nodeID                     string // Cached from id.String() to avoid allocating new strings every time NodeID is called.
	managedOpts                component.Options
	registry                   *prometheus.Registry
	moduleController           ModuleController
	OnNodeWithDependantsUpdate func(cn NodeWithDependants) // Informs controller that we need to reevaluate

	GetModuleInfo  func(fullName string, importLabel string, declareLabel string) (ModuleInfo, error) // Retrieve the module config.
	lastUpdateTime atomic.Time

	mut     sync.RWMutex
	block   *ast.BlockStmt // Current River block to derive args from
	eval    *vm.Evaluator
	managed *module.ModuleComponent // Inner managed module
	args    component.Arguments     // Evaluated arguments for the managed component

	// NOTE(rfratto): health and exports have their own mutex because they may be
	// set asynchronously while mut is still being held (i.e., when calling Evaluate
	// and the managed module immediately creates new exports)

	healthMut  sync.RWMutex
	evalHealth component.Health // Health of the last evaluate
	runHealth  component.Health // Health of running the component

	exportsMut sync.RWMutex
	exports    component.Exports // Evaluated exports for the managed module
}

// ExtractImportAndDeclareLabels extract an importLabel and a declareLabel from a componentName.
func ExtractImportAndDeclareLabels(componentName string) (string, string) {
	parts := strings.Split(componentName, ".")
	if len(parts) == 0 {
		return "", ""
	}
	// If this is a local declare.
	importLabel := ""
	declareLabel := parts[0]
	// If this is an imported module.
	if len(parts) > 1 {
		importLabel = parts[0]
		declareLabel = parts[1]
	}
	return importLabel, declareLabel
}

var _ NodeWithDependants = (*DeclareComponentNode)(nil)
var _ ComponentNode = (*DeclareComponentNode)(nil)

// NewDeclareComponentNode creates a new DeclareComponentNode from an initial ast.BlockStmt.
// The underlying managed module isn't created until Evaluate is called.
func NewDeclareComponentNode(globals ComponentGlobals, b *ast.BlockStmt, GetModuleInfo func(string, string, string) (ModuleInfo, error)) *DeclareComponentNode {
	var (
		id     = BlockComponentID(b)
		nodeID = id.String()
	)

	initHealth := component.Health{
		Health:     component.HealthTypeUnknown,
		Message:    "node declare component created",
		UpdateTime: time.Now(),
	}

	// We need to generate a globally unique component ID to give to the
	// component and for use with telemetry data which doesn't support
	// reconstructing the global ID. For everything else (HTTP, data), we can
	// just use the controller-local ID as those values are guaranteed to be
	// globally unique.
	globalID := nodeID
	if globals.ControllerID != "" {
		globalID = path.Join(globals.ControllerID, nodeID)
	}

	componentName := b.GetBlockName()

	importLabel, declareLabel := ExtractImportAndDeclareLabels(componentName)

	cn := &DeclareComponentNode{
		id:                         id,
		globalID:                   globalID,
		label:                      b.Label,
		nodeID:                     nodeID,
		componentName:              componentName,
		importLabel:                importLabel,
		declareLabel:               declareLabel,
		moduleController:           globals.NewModuleController(globalID),
		OnNodeWithDependantsUpdate: globals.OnNodeWithDependantsUpdate,
		GetModuleInfo:              GetModuleInfo,

		block: b,
		eval:  vm.New(b.Body),

		evalHealth: initHealth,
		runHealth:  initHealth,
	}
	cn.managedOpts = getDeclareManagedOptions(globals, cn)

	return cn
}

func getDeclareManagedOptions(globals ComponentGlobals, cn *DeclareComponentNode) component.Options {
	cn.registry = prometheus.NewRegistry()
	return component.Options{
		ID:     cn.globalID,
		Logger: log.With(globals.Logger, "component", cn.globalID),
		Registerer: prometheus.WrapRegistererWith(prometheus.Labels{
			"component_id": cn.globalID,
		}, cn.registry),
		Tracer: tracing.WrapTracer(globals.TraceProvider, cn.globalID),

		DataPath: filepath.Join(globals.DataPath, cn.globalID),

		OnStateChange:    cn.setExports,
		ModuleController: cn.moduleController,

		GetServiceData: func(name string) (interface{}, error) {
			return globals.GetServiceData(name)
		},
	}
}

// ID returns the component ID of the managed component from its River block.
func (cn *DeclareComponentNode) ID() ComponentID { return cn.id }

// Label returns the label for the block or "" if none was specified.
func (cn *DeclareComponentNode) Label() string { return cn.label }

// NodeID implements dag.Node and returns the unique ID for this node. The
// NodeID is the string representation of the component's ID from its River
// block.
func (cn *DeclareComponentNode) NodeID() string { return cn.nodeID }

// UpdateBlock updates the River block used to construct arguments for the
// managed component. The new block isn't used until the next time Evaluate is
// invoked.
//
// UpdateBlock will panic if the block does not match the component ID of the
// DeclareComponentNode.
func (cn *DeclareComponentNode) UpdateBlock(b *ast.BlockStmt) {
	if !BlockComponentID(b).Equals(cn.id) {
		panic("UpdateBlock called with an River block with a different component ID")
	}

	cn.mut.Lock()
	defer cn.mut.Unlock()
	cn.block = b
	cn.eval = vm.New(b.Body)
}

// Evaluate implements BlockNode and updates the arguments by re-evaluating its River block with the provided scope and the module content by
// retrieving it from the corresponding import or declare node for the managed module.
// The managed module will be built the first time Evaluate is called.
//
// Evaluate will return an error if the River block cannot be evaluated, if
// decoding to arguments fails or if the module content cannot be retrieved.
func (cn *DeclareComponentNode) Evaluate(scope *vm.Scope) error {
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

func (cn *DeclareComponentNode) evaluate(scope *vm.Scope) error {
	cn.mut.Lock()
	defer cn.mut.Unlock()

	var args map[string]any
	if err := cn.eval.Evaluate(scope, &args); err != nil {
		return fmt.Errorf("decoding River: %w", err)
	}

	if cn.managed == nil {
		// We haven't built the managed module successfully yet.
		managed, err := module.NewModuleComponent(cn.managedOpts)
		if err != nil {
			return fmt.Errorf("building module: %w", err)
		}
		cn.managed = managed
	}

	moduleInfo, err := cn.GetModuleInfo(cn.componentName, cn.importLabel, cn.declareLabel)
	if err != nil {
		return fmt.Errorf("retrieving module info: %w", err)
	}

	// Reload the module with new config
	if err := cn.managed.LoadFlowSource(args, moduleInfo.content, moduleInfo.moduleDefinitions); err != nil {
		return fmt.Errorf("updating component: %w", err)
	}
	return nil
}

func (cn *DeclareComponentNode) Run(ctx context.Context) error {
	cn.mut.RLock()
	managed := cn.managed
	logger := cn.managedOpts.Logger
	cn.mut.RUnlock()

	if managed == nil {
		return ErrUnevaluated
	}

	cn.setRunHealth(component.HealthTypeHealthy, "started module")
	cn.managed.RunFlowController(ctx)

	level.Info(logger).Log("msg", "module exited")
	cn.setRunHealth(component.HealthTypeExited, "module shut down")
	return nil
}

// Arguments returns the current arguments of the managed module.
func (cn *DeclareComponentNode) Arguments() component.Arguments {
	cn.mut.RLock()
	defer cn.mut.RUnlock()
	return cn.args
}

// Block implements BlockNode and returns the current block of the managed module.
func (cn *DeclareComponentNode) Block() *ast.BlockStmt {
	cn.mut.RLock()
	defer cn.mut.RUnlock()
	return cn.block
}

// Exports returns the current set of exports from the managed module.
// Exports returns nil if the managed module does not have exports.
func (cn *DeclareComponentNode) Exports() component.Exports {
	cn.exportsMut.RLock()
	defer cn.exportsMut.RUnlock()
	return cn.exports
}

func (cn *DeclareComponentNode) LastUpdateTime() time.Time {
	return cn.lastUpdateTime.Load()
}

// setExports is called whenever the managed module updates. e must be the
// same type as the registered exports type of the managed module.
func (cn *DeclareComponentNode) setExports(e component.Exports) {
	// Some components may aggressively reexport values even though no exposed
	// state has changed. This may be done for components which always supply
	// exports whenever their arguments are evaluated without tracking internal
	// state to see if anything actually changed.
	//
	// To avoid needlessly reevaluating components we'll ignore unchanged
	// exports.
	var changed bool

	cn.exportsMut.Lock()
	if !reflect.DeepEqual(cn.exports, e) {
		changed = true
		cn.exports = e
	}
	cn.exportsMut.Unlock()

	if changed {
		// Inform the controller that we have new exports.
		cn.lastUpdateTime.Store(time.Now())
		cn.OnNodeWithDependantsUpdate(cn)
	}
}

// CurrentHealth returns the current health of the DeclareComponentNode.
//
// The health of a DeclareComponentNode is determined by combining:
//
//  1. Health from the call to Run().
//  2. Health from the last call to Evaluate().
//  3. Health reported from the module.
func (cn *DeclareComponentNode) CurrentHealth() component.Health {
	cn.healthMut.RLock()
	defer cn.healthMut.RUnlock()
	return component.LeastHealthy(cn.runHealth, cn.evalHealth, cn.managed.CurrentHealth())
}

// TODO implement debugInfo?
func (cn *DeclareComponentNode) DebugInfo() interface{} {
	cn.mut.RLock()
	defer cn.mut.RUnlock()
	return nil
}

// setEvalHealth sets the internal health from a call to Evaluate. See Health
// for information on how overall health is calculated.
func (cn *DeclareComponentNode) setEvalHealth(t component.HealthType, msg string) {
	cn.healthMut.Lock()
	defer cn.healthMut.Unlock()

	cn.evalHealth = component.Health{
		Health:     t,
		Message:    msg,
		UpdateTime: time.Now(),
	}
}

// setRunHealth sets the internal health from a call to Run. See Health for
// information on how overall health is calculated.
func (cn *DeclareComponentNode) setRunHealth(t component.HealthType, msg string) {
	cn.healthMut.Lock()
	defer cn.healthMut.Unlock()

	cn.runHealth = component.Health{
		Health:     t,
		Message:    msg,
		UpdateTime: time.Now(),
	}
}

// ModuleIDs returns the current list of modules that this component is
// managing.
func (cn *DeclareComponentNode) ModuleIDs() []string {
	return cn.moduleController.ModuleIDs()
}

// BlockName returns the name of the block.
func (cn *DeclareComponentNode) BlockName() string {
	return cn.componentName
}

// This node does not manage any component.
func (cn *DeclareComponentNode) Component() component.Component {
	return nil
}

// Registry returns the prometheus registry of the component.
func (cn *DeclareComponentNode) Registry() *prometheus.Registry {
	return cn.registry
}
