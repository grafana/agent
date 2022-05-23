package controller

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/grafana/agent/component"
	"github.com/grafana/agent/pkg/flow/internal/dag"
	"github.com/hashicorp/hcl/v2"
	"github.com/rfratto/gohcl"
)

// ComponentID is a fully-qualified name of a component. Each element in
// ComponentID corresponds to a fragment of the period-delimited string;
// "remote.http.example" is ComponentID{"remote", "http", "example"}.
type ComponentID []string

// BlockComponentID returns the ComponentID specified by an HCL block.
func BlockComponentID(b *hcl.Block) ComponentID {
	id := make(ComponentID, 0, 1+len(b.Labels)) // add 1 for the block type
	id = append(id, b.Type)
	id = append(id, b.Labels...)
	return id
}

// String returns the string representation of a component ID.
func (id ComponentID) String() string {
	return strings.Join(id, ".")
}

// Equals returns true if id == other.
func (id ComponentID) Equals(other ComponentID) bool {
	if len(id) != len(other) {
		return false
	}
	for i := 0; i < len(id); i++ {
		if id[i] != other[i] {
			return false
		}
	}
	return true
}

// ComponentGlobals are used by ComponentNodes to build managed components. All
// ComponentNodes should use the same ComponentGlobals.
type ComponentGlobals struct {
	Logger          log.Logger              // Logger shared between all managed components.
	DataPath        string                  // Shared directory where component data may be stored
	OnExportsChange func(cn *ComponentNode) // Invoked when the managed component updated its exports
}

// ComponentNode is a controller node which manages a user-defined component.
//
// ComponentNode manages the underlying component and caches its current
// arguments and exports. ComponentNode manages the arguments for the component
// from an HCL block.
type ComponentNode struct {
	id              ComponentID
	nodeID          string // Cached from id.String() to avoid allocating new strings every time NodeID is called.
	reg             component.Registration
	managedOpts     component.Options
	exportsType     reflect.Type
	onExportsChange func(cn *ComponentNode) // Informs controller that we changed our exports

	mut     sync.RWMutex
	block   *hcl.Block          // Current HCL block to derive args from
	managed component.Component // Inner managed component
	args    component.Arguments // Evaluated arguments for the managed component

	// NOTE(rfratto): health and exports have their own mutex because they may be
	// set asynchronously while mut is still being held (i.e., when calling Evaluate
	// and the managed component immediately creates new exports)

	healthMut  sync.RWMutex
	evalHealth component.Health // Health of the last evaluate
	runHealth  component.Health // Health of running the component

	exportsMut sync.RWMutex
	exports    component.Exports // Evaluated exports for the managed component
}

var (
	_ dag.Node = (*ComponentNode)(nil)
)

// NewComponentNode creates a new ComponentNode from an initial hcl.Block. The
// underlying managed component isn't created until Evaluate is called.
func NewComponentNode(globals ComponentGlobals, b *hcl.Block) *ComponentNode {
	var (
		id     = BlockComponentID(b)
		nodeID = id.String()
	)

	reg, ok := getRegistration(id)
	if !ok {
		// NOTE(rfratto): It's normally not possible to get to this point; the
		// HCL schema should be validated in advance to guarantee that b is an
		// expected component.
		panic("NewComponentNode: could not find registration for component " + nodeID)
	}

	initHealth := component.Health{
		Health:     component.HealthTypeUnknown,
		Message:    "component created",
		UpdateTime: time.Now(),
	}

	cn := &ComponentNode{
		id:              id,
		nodeID:          nodeID,
		reg:             reg,
		exportsType:     getExportsType(reg),
		onExportsChange: globals.OnExportsChange,

		block: b,

		// Prepopulate arguments and exports with their zero values.
		args:    reg.Args,
		exports: reg.Exports,

		evalHealth: initHealth,
		runHealth:  initHealth,
	}
	cn.managedOpts = getManagedOptions(globals, cn)

	return cn
}

func getRegistration(id ComponentID) (component.Registration, bool) {
	// id is the fully qualfied name of the component, including the custom user
	// identifier, if supported by the component. We don't know if the component
	// we're looking up is a singleton or not, so we have to try the lookup
	// twice: once with all name fragments in the ID and once without the last
	// one (e.g., the one that represents the user ID).
	reg, ok := component.Get(id.String())
	if ok {
		return reg, ok
	}

	reg, ok = component.Get(id[:len(id)-1].String())
	return reg, ok
}

func getManagedOptions(globals ComponentGlobals, cn *ComponentNode) component.Options {
	return component.Options{
		ID:            cn.nodeID,
		Logger:        log.With(globals.Logger, "component", cn.nodeID),
		DataPath:      filepath.Join(globals.DataPath, cn.nodeID),
		OnStateChange: cn.setExports,
	}
}

func getExportsType(reg component.Registration) reflect.Type {
	if reg.Exports != nil {
		return reflect.TypeOf(reg.Exports)
	}
	return nil
}

// ID returns the component ID of the managed component from its HCL block.
func (cn *ComponentNode) ID() ComponentID { return cn.id }

// NodeID implements dag.Node and returns the unique ID for this node. The
// NodeID is the string reprsentation of the component's ID from its HCL block.
func (cn *ComponentNode) NodeID() string { return cn.nodeID }

// UpdateBlock updates the HCL block used to construct arguments for the
// managed component. The new block isn't used until the next time Evaluate is
// invoked.
//
// UpdateBlock will panic if the block does not match the component ID of the
// ComponentNode.
func (cn *ComponentNode) UpdateBlock(b *hcl.Block) {
	if !BlockComponentID(b).Equals(cn.id) {
		panic("UpdateBlock called with an HCL block with a different component ID")
	}

	cn.mut.Lock()
	defer cn.mut.Unlock()
	cn.block = b
}

// Evaluate updates the arguments for the managed component by re-evaluating
// its HCL block with the provided evaluation context. The managed component
// will be built the first time Evaluate is called.
//
// Evaluate will return an error if the HCL block cannot be evaluated or if
// decoding to arguments fails.
func (cn *ComponentNode) Evaluate(ectx *hcl.EvalContext) error {
	err := cn.evaluate(ectx)

	switch err {
	case nil:
		cn.setEvalHealth(component.HealthTypeHealthy, "component evaluated")
	default:
		msg := fmt.Sprintf("component evaluation failed: %s", err)
		cn.setEvalHealth(component.HealthTypeUnhealthy, msg)
	}

	return err
}

func (cn *ComponentNode) evaluate(ectx *hcl.EvalContext) error {
	cn.mut.Lock()
	defer cn.mut.Unlock()

	args := cn.reg.CloneArguments()
	diags := gohcl.DecodeBody(cn.block.Body, ectx, args)
	if diags.HasErrors() {
		return fmt.Errorf("decoding HCL: %w", diags)
	}

	// args is always a pointer to the args type, so we want to deference it since
	// components expect a non-pointer.
	argsCopy := reflect.ValueOf(args).Elem().Interface()

	if cn.managed == nil {
		// We haven't built the managed component successfully yet.
		managed, err := cn.reg.Build(cn.managedOpts, argsCopy)
		if err != nil {
			return fmt.Errorf("building component: %w", diags)
		}
		cn.managed = managed
		cn.args = argsCopy

		return nil
	}

	// Update the existing managed component
	if err := cn.managed.Update(argsCopy); err != nil {
		return fmt.Errorf("updating component: %w", err)
	}

	cn.args = argsCopy
	return nil
}

// Run runs the managed component in the calling goroutine until ctx is
// canceled. Evaluate must have been called at least once without retuning an
// error before calling Run.
//
// Run will immediately return ErrUnevaluated if Evaluate has never been called
// successfully. Otherwise, Run will return nil.
func (cn *ComponentNode) Run(ctx context.Context) error {
	cn.mut.RLock()
	managed := cn.managed
	cn.mut.RUnlock()

	if managed == nil {
		return ErrUnevaluated
	}

	cn.setRunHealth(component.HealthTypeHealthy, "started component")
	err := cn.managed.Run(ctx)

	var exitMsg string
	log := cn.managedOpts.Logger
	if err != nil {
		level.Error(log).Log("msg", "component exited with error", "err", err)
		exitMsg = fmt.Sprintf("component shut down with error: %s", err)
	} else {
		level.Info(log).Log("msg", "component exited")
		exitMsg = "component shut down normally"
	}

	cn.setRunHealth(component.HealthTypeExited, exitMsg)
	return err
}

// ErrUnevaluated is returned if ComponentNode.Run is called before a managed
// component is built.
var ErrUnevaluated = errors.New("managed component not built")

// Arguments returns the current arguments of the managed component.
func (cn *ComponentNode) Arguments() component.Arguments {
	cn.mut.RLock()
	defer cn.mut.RUnlock()
	return cn.args
}

// Exports returns the current set of exports from the managed component.
// Exports returns nil if the managed component does not have exports.
func (cn *ComponentNode) Exports() component.Exports {
	cn.exportsMut.RLock()
	defer cn.exportsMut.RUnlock()
	return cn.exports
}

// setExports is called whenever the managed component updates. e must be the
// same type as the registered exports type of the managed component.
func (cn *ComponentNode) setExports(e component.Exports) {
	if cn.exportsType == nil {
		panic(fmt.Sprintf("Component %s called OnStateChange but never registered an Exports type", cn.nodeID))
	}
	if reflect.TypeOf(e) != cn.exportsType {
		panic(fmt.Sprintf("Component %s changed Exports types from %T to %T", cn.nodeID, cn.reg.Exports, e))
	}

	cn.exportsMut.Lock()
	cn.exports = e
	cn.exportsMut.Unlock()

	// Inform the controller that we have new exports.
	cn.onExportsChange(cn)
}

// CurrentHealth returns the current health of the ComponentNode.
//
// The health of a ComponentNode is tracked from three parts, in descending
// prescedence order:
//
//     1. Exited health from a call to Run()
//     2. Unhealthy status from last call to Evaluate
//     3. Health reported by the managed component (if any)
//     4. Latest health from Run() or Evaluate(), if the managed component does not
//        report health.
func (cn *ComponentNode) CurrentHealth() component.Health {
	cn.healthMut.RLock()
	defer cn.healthMut.RUnlock()

	// A component which stopped running takes predence over all other health states
	if cn.runHealth.Health == component.HealthTypeExited {
		return cn.runHealth
	}

	// Next, an unhealthy evaluate takes precedence over the real health of a
	// component.
	if cn.evalHealth.Health != component.HealthTypeHealthy {
		return cn.evalHealth
	}

	// Then, the health of a managed component takes precedence if it is exposed.
	hc, _ := cn.managed.(component.HealthComponent)
	if hc != nil {
		return hc.CurrentHealth()
	}

	// Finally, we return the newer health between eval and run
	latestHealth := cn.evalHealth
	if cn.runHealth.UpdateTime.After(latestHealth.UpdateTime) {
		latestHealth = cn.runHealth
	}
	return latestHealth
}

// DebugInfo returns debugging information from the managed component (if any).
func (cn *ComponentNode) DebugInfo() interface{} {
	cn.mut.RLock()
	defer cn.mut.RUnlock()

	if dc, ok := cn.managed.(component.DebugComponent); ok {
		return dc.DebugInfo()
	}
	return nil
}

// setEvalHealth sets the internal health from a call to Evaluate. See Health
// for information on how overall health is calculated.
func (cn *ComponentNode) setEvalHealth(t component.HealthType, msg string) {
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
func (cn *ComponentNode) setRunHealth(t component.HealthType, msg string) {
	cn.healthMut.Lock()
	defer cn.healthMut.Unlock()

	cn.runHealth = component.Health{
		Health:     t,
		Message:    msg,
		UpdateTime: time.Now(),
	}
}
