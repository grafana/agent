package controller

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"path"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/go-kit/log/level"
	"github.com/grafana/agent/component"
	"github.com/grafana/agent/pkg/cluster"
	"github.com/grafana/agent/pkg/flow/logging"
	"github.com/grafana/agent/pkg/river/ast"
	"github.com/grafana/agent/pkg/river/vm"
	"github.com/prometheus/client_golang/prometheus"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/atomic"
)

// ComponentID is a fully-qualified name of a component. Each element in
// ComponentID corresponds to a fragment of the period-delimited string;
// "remote.http.example" is ComponentID{"remote", "http", "example"}.
type ComponentID []string

// BlockComponentID returns the ComponentID specified by an River block.
func BlockComponentID(b *ast.BlockStmt) ComponentID {
	id := make(ComponentID, 0, len(b.Name)+1) // add 1 for the optional label
	id = append(id, b.Name...)
	if b.Label != "" {
		id = append(id, b.Label)
	}
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

// DialFunc is a function to establish a network connection.
type DialFunc func(ctx context.Context, network, address string) (net.Conn, error)

// ComponentGlobals are used by ComponentNodes to build managed components. All
// ComponentNodes should use the same ComponentGlobals.
type ComponentGlobals struct {
	LogSink           *logging.Sink                // Sink used for Logging.
	Logger            *logging.Logger              // Logger shared between all managed components.
	TraceProvider     trace.TracerProvider         // Tracer shared between all managed components.
	Clusterer         *cluster.Clusterer           // Clusterer shared between all managed components.
	DataPath          string                       // Shared directory where component data may be stored
	OnComponentUpdate func(cn *ComponentNode)      // Informs controller that we need to reevaluate
	OnExportsChange   func(exports map[string]any) // Invoked when the managed component updated its exports
	Registerer        prometheus.Registerer        // Registerer for serving agent and component metrics
	HTTPPathPrefix    string                       // HTTP prefix for components.
	HTTPListenAddr    string                       // Base address for server
	DialFunc          DialFunc                     // Function to connect to HTTPListenAddr.
	ControllerID      string                       // ID of controller.
}

// ComponentNode is a controller node which manages a user-defined component.
//
// ComponentNode manages the underlying component and caches its current
// arguments and exports. ComponentNode manages the arguments for the component
// from a River block.
type ComponentNode struct {
	id                ComponentID
	label             string
	componentName     string
	nodeID            string // Cached from id.String() to avoid allocating new strings every time NodeID is called.
	reg               component.Registration
	managedOpts       component.Options
	registry          *prometheus.Registry
	exportsType       reflect.Type
	OnComponentUpdate func(cn *ComponentNode) // Informs controller that we need to reevaluate

	mut     sync.RWMutex
	block   *ast.BlockStmt // Current River block to derive args from
	eval    *vm.Evaluator
	managed component.Component // Inner managed component
	args    component.Arguments // Evaluated arguments for the managed component

	doingEval atomic.Bool

	// NOTE(rfratto): health and exports have their own mutex because they may be
	// set asynchronously while mut is still being held (i.e., when calling Evaluate
	// and the managed component immediately creates new exports)

	healthMut  sync.RWMutex
	evalHealth component.Health // Health of the last evaluate
	runHealth  component.Health // Health of running the component

	exportsMut sync.RWMutex
	exports    component.Exports // Evaluated exports for the managed component
}

var _ BlockNode = (*ComponentNode)(nil)

// NewComponentNode creates a new ComponentNode from an initial ast.BlockStmt.
// The underlying managed component isn't created until Evaluate is called.
func NewComponentNode(globals ComponentGlobals, b *ast.BlockStmt) *ComponentNode {
	var (
		id     = BlockComponentID(b)
		nodeID = id.String()
	)

	reg, ok := component.Get(ComponentID(b.Name).String())
	if !ok {
		// NOTE(rfratto): It's normally not possible to get to this point; the
		// blocks should have been validated by the graph loader in advance to
		// guarantee that b is an expected component.
		panic("NewComponentNode: could not find registration for component " + nodeID)
	}

	initHealth := component.Health{
		Health:     component.HealthTypeUnknown,
		Message:    "component created",
		UpdateTime: time.Now(),
	}

	cn := &ComponentNode{
		id:                id,
		label:             b.Label,
		nodeID:            nodeID,
		componentName:     strings.Join(b.Name, "."),
		reg:               reg,
		exportsType:       getExportsType(reg),
		OnComponentUpdate: globals.OnComponentUpdate,

		block: b,
		eval:  vm.New(b.Body),

		// Prepopulate arguments and exports with their zero values.
		args:    reg.Args,
		exports: reg.Exports,

		evalHealth: initHealth,
		runHealth:  initHealth,
	}
	cn.managedOpts = getManagedOptions(globals, cn)

	return cn
}

func getManagedOptions(globals ComponentGlobals, cn *ComponentNode) component.Options {
	// Make sure the prefix is always absolute.
	prefix := globals.HTTPPathPrefix
	if !strings.HasPrefix(prefix, "/") {
		prefix = "/" + prefix
	}

	// We need to generate a globally unique component ID to give to the
	// component and for use with telemetry data which doesn't support
	// reconstructing the global ID. For everything else (HTTP, data), we can
	// just use the controller-local ID as those values are guaranteed to be
	// globally unique.
	globalID := cn.nodeID
	if globals.ControllerID != "" {
		globalID = path.Join(globals.ControllerID, cn.nodeID)
	}

	cn.registry = prometheus.NewRegistry()
	return component.Options{
		ID:     globalID,
		Logger: logging.New(logging.LoggerSink(globals.Logger), logging.WithComponentID(cn.nodeID)),
		Registerer: prometheus.WrapRegistererWith(prometheus.Labels{
			"component_id": globalID,
		}, cn.registry),
		Tracer:    wrapTracer(globals.TraceProvider, globalID),
		Clusterer: globals.Clusterer,

		DataPath:       filepath.Join(globals.DataPath, cn.nodeID),
		HTTPListenAddr: globals.HTTPListenAddr,
		DialFunc:       globals.DialFunc,
		HTTPPath:       path.Join(prefix, cn.nodeID) + "/",

		OnStateChange: cn.setExports,
	}
}

func getExportsType(reg component.Registration) reflect.Type {
	if reg.Exports != nil {
		return reflect.TypeOf(reg.Exports)
	}
	return nil
}

// ID returns the component ID of the managed component from its River block.
func (cn *ComponentNode) ID() ComponentID { return cn.id }

// Label returns the label for the block or "" if none was specified.
func (cn *ComponentNode) Label() string { return cn.label }

// ComponentName returns the component's type, i.e. `local.file.test` returns `local.file`.
func (cn *ComponentNode) ComponentName() string { return cn.componentName }

// NodeID implements dag.Node and returns the unique ID for this node. The
// NodeID is the string representation of the component's ID from its River
// block.
func (cn *ComponentNode) NodeID() string { return cn.nodeID }

// UpdateBlock updates the River block used to construct arguments for the
// managed component. The new block isn't used until the next time Evaluate is
// invoked.
//
// UpdateBlock will panic if the block does not match the component ID of the
// ComponentNode.
func (cn *ComponentNode) UpdateBlock(b *ast.BlockStmt) {
	if !BlockComponentID(b).Equals(cn.id) {
		panic("UpdateBlock called with an River block with a different component ID")
	}

	cn.mut.Lock()
	defer cn.mut.Unlock()
	cn.block = b
	cn.eval = vm.New(b.Body)
}

// Evaluate implements BlockNode and updates the arguments for the managed component
// by re-evaluating its River block with the provided scope. The managed component
// will be built the first time Evaluate is called.
//
// Evaluate will return an error if the River block cannot be evaluated or if
// decoding to arguments fails.
func (cn *ComponentNode) Evaluate(scope *vm.Scope) error {
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

// Reevaluate calls Update on the managed component with its last used
// arguments.Reevaluate does not build the component if it is not already built
// and does not re-evaluate the River block itself.
// Its only use case is for components opting-in to clustering where calling
// Update with the same Arguments may result in different functionality.
func (cn *ComponentNode) Reevaluate() error {
	cn.mut.Lock()
	defer cn.mut.Unlock()

	cn.doingEval.Store(true)
	defer cn.doingEval.Store(false)

	if cn.managed == nil {
		// We haven't built the managed component successfully yet.
		return nil
	}

	// Update the existing managed component with the same arguments.
	err := cn.managed.Update(cn.args)

	switch err {
	case nil:
		cn.setEvalHealth(component.HealthTypeHealthy, "component evaluated")
		return nil
	default:
		msg := fmt.Sprintf("component evaluation failed: %s", err)
		cn.setEvalHealth(component.HealthTypeUnhealthy, msg)
		return err
	}
}

func (cn *ComponentNode) evaluate(scope *vm.Scope) error {
	cn.mut.Lock()
	defer cn.mut.Unlock()

	cn.doingEval.Store(true)
	defer cn.doingEval.Store(false)

	argsPointer := cn.reg.CloneArguments()
	if err := cn.eval.Evaluate(scope, argsPointer); err != nil {
		return fmt.Errorf("decoding River: %w", err)
	}

	// args is always a pointer to the args type, so we want to deference it since
	// components expect a non-pointer.
	argsCopyValue := reflect.ValueOf(argsPointer).Elem().Interface()

	if cn.managed == nil {
		// We haven't built the managed component successfully yet.
		managed, err := cn.reg.Build(cn.managedOpts, argsCopyValue)
		if err != nil {
			return fmt.Errorf("building component: %w", err)
		}
		cn.managed = managed
		cn.args = argsCopyValue

		return nil
	}

	if reflect.DeepEqual(cn.args, argsCopyValue) {
		// Ignore components which haven't changed. This reduces the cost of
		// calling evaluate for components where evaluation is expensive (e.g., if
		// re-evaluating requires re-starting some internal logic).
		return nil
	}

	// Update the existing managed component
	if err := cn.managed.Update(argsCopyValue); err != nil {
		return fmt.Errorf("updating component: %w", err)
	}

	cn.args = argsCopyValue
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

// ErrUnevaluated is returned if ComponentNode.Run is called before a managed
// component is built.
var ErrUnevaluated = errors.New("managed component not built")

// Arguments returns the current arguments of the managed component.
func (cn *ComponentNode) Arguments() component.Arguments {
	cn.mut.RLock()
	defer cn.mut.RUnlock()
	return cn.args
}

// Block implements BlockNode and returns the current block of the managed component.
func (cn *ComponentNode) Block() *ast.BlockStmt {
	cn.mut.RLock()
	defer cn.mut.RUnlock()
	return cn.block
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

	if cn.doingEval.Load() {
		// Optimization edge case: some components supply exports when they're
		// being evaluated.
		//
		// Since components that are being evaluated will always cause their
		// dependencies to also be evaluated, there's no reason to call
		// onExportsChange here.
		return
	}

	if changed {
		// Inform the controller that we have new exports.
		cn.OnComponentUpdate(cn)
	}
}

// CurrentHealth returns the current health of the ComponentNode.
//
// The health of a ComponentNode is determined by combining:
//
//  1. Health from the call to Run().
//  2. Health from the last call to Evaluate().
//  3. Health reported from the component.
func (cn *ComponentNode) CurrentHealth() component.Health {
	cn.healthMut.RLock()
	defer cn.healthMut.RUnlock()

	var (
		runHealth  = cn.runHealth
		evalHealth = cn.evalHealth
	)

	if hc, ok := cn.managed.(component.HealthComponent); ok {
		componentHealth := hc.CurrentHealth()
		return component.LeastHealthy(runHealth, evalHealth, componentHealth)
	}

	return component.LeastHealthy(runHealth, evalHealth)
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

// HTTPHandler returns an http handler for a component IF it implements HTTPComponent.
// otherwise it will return nil.
func (cn *ComponentNode) HTTPHandler() http.Handler {
	handler, ok := cn.managed.(component.HTTPComponent)
	if !ok {
		return nil
	}
	return handler.Handler()
}
