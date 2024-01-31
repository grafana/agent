package controller

import (
	"context"
	"fmt"
	"path"
	"reflect"
	"sync"
	"time"

	"github.com/go-kit/log"
	"github.com/grafana/agent/component"
	"github.com/grafana/agent/pkg/flow/logging/level"
	"github.com/grafana/river/ast"
	"github.com/grafana/river/vm"
)

// CustomComponentNode is a controller node which manages a custom component.
//
// CustomComponentNode manages the underlying custom component and caches its
// current arguments and exports.
type CustomComponentNode struct {
	id                ComponentID
	label             string
	componentName     string
	nodeID            string // Cached from id.String() to avoid allocating new strings every time NodeID is called.
	customComponent   CustomComponent
	exportsType       reflect.Type
	moduleController  ModuleController
	onBlockNodeUpdate func(cn BlockNode) // Informs controller that we need to reevaluate dependants.
	logger            log.Logger

	mut    sync.RWMutex
	block  *ast.BlockStmt
	eval   *vm.Evaluator
	module component.Module    // Inner managed component.
	args   component.Arguments // Evaluated arguments for the managed component.

	// NOTE(rfratto): health and exports have their own mutex because they may be
	// set asynchronously while mut is still being held (i.e., when calling
	// Evaluate and the managed component immediately creates new exports).

	healthMut  sync.RWMutex
	evalHealth component.Health // Health of the last evaluation.
	runHealth  component.Health // Health of running the component.

	exportsMut sync.RWMutex
	exports    component.Exports // Evaluated exports for the managed component.
}

var _ ComponentNode = (*CustomComponentNode)(nil)

// NewCustomComponentNode creates a new CustomComponentNode from an initial
// ast.BlockStmt.
//
// The underlying managed custom component isn't created until Evaluate is called.
func NewCustomComponentNode(globals ComponentGlobals, reg CustomComponent, b *ast.BlockStmt) *CustomComponentNode {
	var (
		id     = BlockComponentID(b)
		nodeID = id.String()
	)

	initHealth := component.Health{
		Health:     component.HealthTypeUnknown,
		Message:    "custom component created",
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

	cn := &CustomComponentNode{
		id:                id,
		label:             b.Label,
		nodeID:            nodeID,
		componentName:     b.GetBlockName(),
		customComponent:   reg,
		exportsType:       reflect.TypeOf(map[string]any{}),
		moduleController:  globals.NewModuleController(globalID),
		onBlockNodeUpdate: globals.OnBlockNodeUpdate,
		logger:            log.With(globals.Logger, "component", globalID),

		block: b,
		eval:  vm.New(b.Body),

		// Prepopulate arguments and exports with their zero values.
		args:    map[string]any{},
		exports: map[string]any{},

		evalHealth: initHealth,
		runHealth:  initHealth,
	}

	return cn
}

// NodeID implements [dag.Node] and returns the unique ID for this node. The
// NodeID is the string representation of the component's ID from its River
// block.
func (c *CustomComponentNode) NodeID() string { return c.nodeID }

// Block implements [BlockNode] and returns the current block of the custom
// component.
func (c *CustomComponentNode) Block() *ast.BlockStmt {
	c.mut.RLock()
	defer c.mut.RUnlock()
	return c.block
}

// Evaluate implements [BlockNode] and updates the arguments for the custom
// component by re-evaluating its River block with the provided scope. A
// controller for the custom component will be built the first time Evaluate is
// called.
//
// Evaluate will return an error if the River block cannot be evaluated, if
// decoding to arguments fails, or if the definition for the custom component
// cannot be retrieved.
func (c *CustomComponentNode) Evaluate(scope *vm.Scope) error {
	err := c.evaluate(scope)

	switch err {
	case nil:
		c.setEvalHealth(component.HealthTypeHealthy, "component evaluated")
	default:
		msg := fmt.Sprintf("component evaluation failed: %s", err)
		c.setEvalHealth(component.HealthTypeUnhealthy, msg)
	}

	return err
}

func (c *CustomComponentNode) evaluate(scope *vm.Scope) error {
	c.mut.Lock()
	defer c.mut.Unlock()

	// BUG(rfratto): c.customComponent here relies on a single version of the
	// graph and does not respect updates to the Loader's graph.
	//
	// If the definition of a component changes, we will not get the new change
	// as c.customComponent is kept for the lifetime of our component node.
	body, err := c.customComponent.Definition()
	if err != nil {
		return fmt.Errorf("failed to get definition of %s: %w", c.componentName, err)
	}

	var args map[string]any
	if err := c.eval.Evaluate(scope, &args); err != nil {
		return fmt.Errorf("decoding River: %w", err)
	}

	if c.module == nil {
		mod, err := c.moduleController.NewModule("", func(exports map[string]any) { c.setExports(exports) })
		if err != nil {
			return fmt.Errorf("creating custom component controller: %w", err)
		}
		c.module = mod
	}

	if err := c.module.LoadBody(body, args); err != nil {
		return fmt.Errorf("updating component: %w", err)
	}

	c.args = args
	return nil
}

func (c *CustomComponentNode) Run(ctx context.Context) error {
	c.mut.RLock()
	managed := c.module
	c.mut.RUnlock()

	if managed == nil {
		return ErrUnevaluated
	}

	c.setRunHealth(component.HealthTypeHealthy, "started component")
	err := c.module.Run(ctx)

	var exitMsg string
	if err != nil {
		level.Error(c.logger).Log("msg", "component exited with error", "err", err)
		exitMsg = fmt.Sprintf("component shut down with error: %s", err)
	} else {
		level.Info(c.logger).Log("msg", "component exited")
		exitMsg = "component shut down normally"
	}

	c.setRunHealth(component.HealthTypeExited, exitMsg)
	return err
}

// CurrentHealth returns the current health of the CustomComponentNode.
//
// The health of a CustomComponent node is determined by combining:
//
//  1. Health from the call to Run().
//  2. Health from the last call to Evaluate().
func (c *CustomComponentNode) CurrentHealth() component.Health {
	c.healthMut.RLock()
	defer c.healthMut.RUnlock()

	var (
		runHealth  = c.runHealth
		evalHealth = c.evalHealth
	)

	return component.LeastHealthy(runHealth, evalHealth)
}

// Arguments returns the current arguments of the custom component.
func (c *CustomComponentNode) Arguments() component.Arguments {
	c.mut.RLock()
	defer c.mut.RUnlock()
	return c.args
}

// Exports returns the current set of exports from the custom component.
// Exports returns nil if the custom component does not have exports.
func (c *CustomComponentNode) Exports() component.Exports {
	c.exportsMut.RLock()
	defer c.exportsMut.RUnlock()
	return c.exports
}

// setExports is called whenever the custom module updates. e must be a
// map[string]any.
func (c *CustomComponentNode) setExports(e component.Exports) {
	if reflect.TypeOf(e) != c.exportsType {
		panic(fmt.Sprintf("Component %s changed Exports types from %T to %T", c.nodeID, c.exportsType, e))
	}

	// Some components may aggressively reexport values even though no exposed
	// state has changed. This may be done for components which always supply
	// exports whenever their arguments are evaluated without tracking internal
	// state to see if anything actually changed.
	//
	// To avoid needlessly reevaluating components we'll ignore unchanged
	// exports.
	var changed bool

	c.exportsMut.Lock()
	if !reflect.DeepEqual(c.exports, e) {
		changed = true
		c.exports = e
	}
	c.exportsMut.Unlock()

	if changed {
		// Inform the controller that we have new exports.
		c.onBlockNodeUpdate(c)
	}
}

// Label returns the label for the custom component.
func (c *CustomComponentNode) Label() string { return c.label }

// ComponentName returns the component's name, corresponding to the River block
// name without the label.
func (c *CustomComponentNode) ComponentName() string { return c.componentName }

// ID returns the component ID of the custom component from its River block.
func (c *CustomComponentNode) ID() ComponentID { return c.id }

// UpdateBlock updates the River block used to construct arguments for the
// custom component. The new block isn't used until the next time Evaluate is
// invoked.
//
// UpdateBlock will panic if the block does not match the component ID of the
// CustomComponentNode assigned on creation.
func (c *CustomComponentNode) UpdateBlock(b *ast.BlockStmt) {
	if !BlockComponentID(b).Equals(c.id) {
		panic("UpdateBlock called with a River block with a different component ID")
	}

	c.mut.Lock()
	defer c.mut.Unlock()
	c.block = b
	c.eval = vm.New(b.Body)
}

// setEvalHealth sets the internal health from a call to Evaluate. See Health
// for information on how overall health is calculated.
func (c *CustomComponentNode) setEvalHealth(t component.HealthType, msg string) {
	c.healthMut.Lock()
	defer c.healthMut.Unlock()

	c.evalHealth = component.Health{
		Health:     t,
		Message:    msg,
		UpdateTime: time.Now(),
	}
}

// setRunHealth sets the internal health from a call to Run. See Health for
// information on how overall health is calculated.
func (c *CustomComponentNode) setRunHealth(t component.HealthType, msg string) {
	c.healthMut.Lock()
	defer c.healthMut.Unlock()

	c.runHealth = component.Health{
		Health:     t,
		Message:    msg,
		UpdateTime: time.Now(),
	}
}
