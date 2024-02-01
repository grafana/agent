package controller

import (
	"context"
	"fmt"
	"path"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/go-kit/log"
	"github.com/grafana/agent/component"
	"github.com/grafana/agent/pkg/flow/config"
	"github.com/grafana/agent/pkg/flow/logging/level"
	"github.com/grafana/river/ast"
	"github.com/grafana/river/vm"
)

// CustomComponentNode is a controller node which manages a custom component.
//
// CustomComponentNode manages the underlying custom component and caches its current
// arguments and exports.
type CustomComponentNode struct {
	id                ComponentID
	globalID          string
	label             string
	componentName     string
	importLabel       string
	declareLabel      string
	nodeID            string // Cached from id.String() to avoid allocating new strings every time NodeID is called.
	moduleController  ModuleController
	OnBlockNodeUpdate func(cn BlockNode) // Informs controller that we need to reevaluate
	logger            log.Logger

	getConfig getCustomComponentConfig // Retrieve the custom component config.

	mut     sync.RWMutex
	block   *ast.BlockStmt // Current River block to derive args from
	eval    *vm.Evaluator
	managed component.Module    // Inner managed custom component
	args    component.Arguments // Evaluated arguments for the managed component

	// NOTE(rfratto): health and exports have their own mutex because they may be
	// set asynchronously while mut is still being held (i.e., when calling Evaluate
	// and the managed custom component immediately creates new exports)

	healthMut  sync.RWMutex
	evalHealth component.Health // Health of the last evaluate
	runHealth  component.Health // Health of running the component

	exportsMut sync.RWMutex
	exports    component.Exports // Evaluated exports for the managed custom component
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
	// If this is an imported custom component.
	if len(parts) > 1 {
		importLabel = parts[0]
		declareLabel = parts[1]
	}
	return importLabel, declareLabel
}

var _ ComponentNode = (*CustomComponentNode)(nil)

// NewCustomComponentNode creates a new CustomComponentNode from an initial ast.BlockStmt.
// The underlying managed custom component isn't created until Evaluate is called.
func NewCustomComponentNode(globals ComponentGlobals, b *ast.BlockStmt, getConfig getCustomComponentConfig) *CustomComponentNode {
	var (
		id     = BlockComponentID(b)
		nodeID = id.String()
	)

	initHealth := component.Health{
		Health:     component.HealthTypeUnknown,
		Message:    "node custom component created",
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

	cn := &CustomComponentNode{
		id:                id,
		globalID:          globalID,
		label:             b.Label,
		nodeID:            nodeID,
		componentName:     componentName,
		importLabel:       importLabel,
		declareLabel:      declareLabel,
		moduleController:  globals.NewModuleController(globalID),
		OnBlockNodeUpdate: globals.OnBlockNodeUpdate,
		logger:            log.With(globals.Logger, "component", globalID),
		getConfig:         getConfig,

		block: b,
		eval:  vm.New(b.Body),

		evalHealth: initHealth,
		runHealth:  initHealth,
	}

	return cn
}

// ID returns the component ID of the managed component from its River block.
func (cn *CustomComponentNode) ID() ComponentID { return cn.id }

// Label returns the label for the block or "" if none was specified.
func (cn *CustomComponentNode) Label() string { return cn.label }

// NodeID implements dag.Node and returns the unique ID for this node. The
// NodeID is the string representation of the component's ID from its River
// block.
func (cn *CustomComponentNode) NodeID() string { return cn.nodeID }

// UpdateBlock updates the River block used to construct arguments for the
// managed component. The new block isn't used until the next time Evaluate is
// invoked.
//
// UpdateBlock will panic if the block does not match the component ID of the
// CustomComponentNode.
func (cn *CustomComponentNode) UpdateBlock(b *ast.BlockStmt) {
	if !BlockComponentID(b).Equals(cn.id) {
		panic("UpdateBlock called with an River block with a different component ID")
	}

	cn.mut.Lock()
	defer cn.mut.Unlock()
	cn.block = b
	cn.eval = vm.New(b.Body)
}

// Evaluate implements BlockNode and updates the arguments by re-evaluating its River block with the provided scope and the custom component by
// retrieving the component definition from the corresponding import or declare node.
// The managed custom component will be built the first time Evaluate is called.
//
// Evaluate will return an error if the River block cannot be evaluated, if
// decoding to arguments fails or if the custom component definition cannot be retrieved.
func (cn *CustomComponentNode) Evaluate(evalScope *vm.Scope) error {
	err := cn.evaluate(evalScope)

	switch err {
	case nil:
		cn.setEvalHealth(component.HealthTypeHealthy, "component evaluated")
	default:
		msg := fmt.Sprintf("component evaluation failed: %s", err)
		cn.setEvalHealth(component.HealthTypeUnhealthy, msg)
	}
	return err
}

func (cn *CustomComponentNode) evaluate(evalScope *vm.Scope) error {
	cn.mut.Lock()
	defer cn.mut.Unlock()

	var args map[string]any
	if err := cn.eval.Evaluate(evalScope, &args); err != nil {
		return fmt.Errorf("decoding River: %w", err)
	}

	cn.args = args

	if cn.managed == nil {
		// We haven't built the managed custom component successfully yet.
		mod, err := cn.moduleController.NewModule("", func(exports map[string]any) { cn.setExports(exports) })
		if err != nil {
			return fmt.Errorf("creating custom component controller: %w", err)
		}
		cn.managed = mod
	}

	template, scope := cn.getConfig(cn.importLabel, cn.declareLabel)
	if template == nil || scope == nil {
		return fmt.Errorf("could not retrieve custom component config")
	}

	loaderConfig := config.LoaderConfigOptions{
		Scope: scope,
	}

	// Reload the custom component with new config
	if err := cn.managed.LoadBody(template, args, loaderConfig); err != nil {
		return fmt.Errorf("updating custom component: %w", err)
	}
	return nil
}

func (cn *CustomComponentNode) Run(ctx context.Context) error {
	cn.mut.RLock()
	managed := cn.managed
	logger := cn.logger
	cn.mut.RUnlock()

	if managed == nil {
		return ErrUnevaluated
	}

	cn.setRunHealth(component.HealthTypeHealthy, "started custom component")
	err := managed.Run(ctx)
	if err != nil {
		level.Error(logger).Log("msg", "error running custom component", "id", cn.nodeID, "err", err)
	}

	level.Info(logger).Log("msg", "custom component exited")
	cn.setRunHealth(component.HealthTypeExited, "custom component shut down")
	return err
}

// Arguments returns the current arguments of the managed custom component.
func (cn *CustomComponentNode) Arguments() component.Arguments {
	cn.mut.RLock()
	defer cn.mut.RUnlock()
	return cn.args
}

// Block implements BlockNode and returns the current block of the managed custom component.
func (cn *CustomComponentNode) Block() *ast.BlockStmt {
	cn.mut.RLock()
	defer cn.mut.RUnlock()
	return cn.block
}

// Exports returns the current set of exports from the managed custom component.
// Exports returns nil if the managed custom component does not have exports.
func (cn *CustomComponentNode) Exports() component.Exports {
	cn.exportsMut.RLock()
	defer cn.exportsMut.RUnlock()
	return cn.exports
}

// setExports is called whenever the managed custom component updates. e must be the
// same type as the registered exports type of the managed custom component.
func (cn *CustomComponentNode) setExports(e component.Exports) {
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
		cn.OnBlockNodeUpdate(cn)
	}
}

// CurrentHealth returns the current health of the CustomComponentNode.
//
// The health of a CustomComponentNode is determined by combining:
//
//  1. Health from the call to Run().
//  2. Health from the last call to Evaluate().
//  3. Health reported from the custom component.
func (cn *CustomComponentNode) CurrentHealth() component.Health {
	cn.healthMut.RLock()
	defer cn.healthMut.RUnlock()
	return component.LeastHealthy(cn.runHealth, cn.evalHealth)
}

// setEvalHealth sets the internal health from a call to Evaluate. See Health
// for information on how overall health is calculated.
func (cn *CustomComponentNode) setEvalHealth(t component.HealthType, msg string) {
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
func (cn *CustomComponentNode) setRunHealth(t component.HealthType, msg string) {
	cn.healthMut.Lock()
	defer cn.healthMut.Unlock()

	cn.runHealth = component.Health{
		Health:     t,
		Message:    msg,
		UpdateTime: time.Now(),
	}
}

// ComponentName returns the name of the component.
func (cn *CustomComponentNode) ComponentName() string {
	return cn.componentName
}

// TODO: currently used by the component provider to access the components running within
// the custom components. Change it when getting rid of old modules.
func (cn *CustomComponentNode) ModuleIDs() []string {
	return cn.moduleController.ModuleIDs()
}
