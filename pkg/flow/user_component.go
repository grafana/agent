package flow

import (
	"fmt"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/go-kit/log"
	"github.com/grafana/agent/pkg/flow/component"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/rfratto/gohcl"
)

// userComponentName is the fully-qualified name of a userComponent.
type userComponentName []string

// blockToComponentName gets the userComponentName from an HCL block.
func blockToComponentName(b *hcl.Block) userComponentName {
	name := make(userComponentName, 0, 1+len(b.Labels))
	name = append(name, b.Type)
	name = append(name, b.Labels...)
	return name
}

func (r userComponentName) String() string {
	return strings.Join(r, ".")
}

func (r userComponentName) Equals(other userComponentName) bool {
	if len(r) != len(other) {
		return false
	}
	for i := 0; i < len(r); i++ {
		if r[i] != other[i] {
			return false
		}
	}
	return true
}

// userComponentOptions are options shared between all components.
type userComponentOptions struct {
	Logger        log.Logger              // Base logger shared between components
	DataPath      string                  // Path where components may store data
	OnStateChange func(uc *userComponent) // Invoked to propagate state change
}

// userComponent is a node in controller's graph which corresponds to a
// component defined by a user.
type userComponent struct {
	name userComponentName
	reg  component.Registration
	opts component.Options

	mut    sync.RWMutex
	block  *hcl.Block
	inner  component.Component // Inner component
	config component.Config    // Current evaluted config for the inner component
	health component.Health    // Current reported component health

	exportsMut sync.RWMutex
	exports    component.Exports // Most reset exports from the inner component. May be nil.
}

// newUserComponent constructs a blank userComponent from an hcl.Block.
func newUserComponent(opts userComponentOptions, b *hcl.Block) *userComponent {
	name := blockToComponentName(b)

	// We need to find the registration definition of the component. We don't
	// know if the final element in our name is a user identifier (indicating a
	// non-singleton component), so we have to try to look up the component name
	// twice, once without and once with the final label.
	reg, ok := component.Get(name[:len(name)-1].String())
	if !ok {
		// Nope: we're probably representing a singleton component then. Use the
		// full name for the component ID.
		reg, ok = component.Get(name.String())
		if !ok {
			// We can only reach this point after validating the HCL schema which
			// contains all known blocks. Getting here indicates a bug in mapping the
			// block labels to a component and not that there's a missing component.
			panic("Could not find registration for component " + name.String())
		}
	}

	var exportsType reflect.Type
	if reg.Exports != nil {
		exportsType = reflect.TypeOf(reg.Exports)
	}

	uc := &userComponent{
		name: blockToComponentName(b),
		reg:  reg,
		opts: component.Options{
			ID:       name.String(),
			Logger:   log.With(opts.Logger, "component", name.String()),
			DataPath: filepath.Join(opts.DataPath, name.String()),
		},

		block: b,

		// Pre-populate the current config and exports using the registration.
		config:  reg.Config,
		exports: reg.Exports,

		health: component.Health{
			Health:     component.HealthTypeUnkown,
			Message:    "component created",
			UpdateTime: time.Now(),
		},
	}

	// Wire up the state change handler.
	uc.opts.OnStateChange = func(e component.Exports) {
		if exportsType == nil {
			panic(fmt.Sprintf("Component %s called OnStateChange but never registered an Exports type", name))
		}
		if reflect.TypeOf(e) != exportsType {
			panic(fmt.Sprintf("Component %s changed Exports types from %T to %T", name, reg.Exports, e))
		}

		uc.exportsMut.Lock()
		uc.exports = e
		uc.exportsMut.Unlock()

		// Propagate to the parent.
		opts.OnStateChange(uc)
	}

	return uc
}

// Name returns the fully-qualified component name.
func (uc *userComponent) Name() userComponentName { return uc.name }

// NodeID implements dag.Node, returning the fully-qualified name of the
// component as a string.
func (uc *userComponent) NodeID() string { return uc.name.String() }

// SetBlock updates the hcl.Block used by the userComponent to construct its
// config.
//
// SetBlock will panic if the component name specified by b does not match the
// userComponent's current name.
func (uc *userComponent) SetBlock(b *hcl.Block) {
	name := blockToComponentName(b)
	if !uc.name.Equals(name) {
		panic("SetBlock called with an HCL block that has a different component name")
	}

	uc.mut.Lock()
	defer uc.mut.Unlock()
	uc.block = b
}

// Traversals returns the variable references this user component makes.
func (uc *userComponent) Traversals() []hcl.Traversal {
	uc.mut.RLock()
	defer uc.mut.RUnlock()
	return expressionsFromSyntaxBody(uc.block.Body.(*hclsyntax.Body))
}

// expressionsFromSyntaxBody recurses through body and finds all variable
// references.
func expressionsFromSyntaxBody(body *hclsyntax.Body) []hcl.Traversal {
	var exprs []hcl.Traversal

	for _, attrib := range body.Attributes {
		exprs = append(exprs, attrib.Expr.Variables()...)
	}
	for _, block := range body.Blocks {
		exprs = append(exprs, expressionsFromSyntaxBody(block.Body)...)
	}

	return exprs
}

// Evaluate evaluates the current block for the component using the provided
// EvalContext and synchronizes state with the inner managed component.
//
// The inner managed component will be built the first time calling Evaluate.
// On subsequent calls, the inner managed component will be updated instead.
func (uc *userComponent) Evaluate(ectx *hcl.EvalContext) error {
	uc.mut.Lock()
	defer uc.mut.Unlock()

	cfg := uc.reg.CloneConfig()
	diags := gohcl.DecodeBody(uc.block.Body, ectx, cfg)
	if diags.HasErrors() {
		return diags
	}

	// cfg is always a pointer to the struct type, so we want to dereference it
	// since components expect a non-pointer.
	cfgCopy := reflect.ValueOf(cfg).Elem().Interface()

	if uc.inner == nil {
		// We've never built the component before.
		inner, err := uc.reg.Build(uc.opts, cfgCopy)
		if err != nil {
			return err
		}
		uc.inner = inner
		uc.config = cfgCopy
		return nil
	}

	// Update the existing component.
	if err := uc.inner.Update(cfgCopy); err != nil {
		return err
	}

	uc.config = cfgCopy
	return nil
}

// Get returns the inner managed component. This will return nil until Evaluate
// has been called successfully at least once.
func (uc *userComponent) Get() component.Component {
	uc.mut.RLock()
	defer uc.mut.RUnlock()
	return uc.inner
}

// CurrentConfig returns the current evaluated config of the component.
func (uc *userComponent) CurrentConfig() component.Config {
	uc.mut.RLock()
	defer uc.mut.RUnlock()
	return uc.config
}

// CurrentExports returns the current set of exports for the component.
func (uc *userComponent) CurrentExports() component.Exports {
	uc.exportsMut.RLock()
	defer uc.exportsMut.RUnlock()
	return uc.exports
}

// CurrentHealth gets the current health of the userComponent or inner managed
// component's health. An unhealthy userComponent takes precedence over the
// health of the inner component.
//
// If the inner manager component doesn't export health, only the
// userComponent's health is returned.
func (uc *userComponent) CurrentHealth() component.Health {
	uc.mut.RLock()
	defer uc.mut.RUnlock()

	if uc.health.Health == component.HealthTypeUnhealthy {
		return uc.health
	}

	hc, _ := uc.inner.(component.HealthComponent)
	if hc == nil {
		// Inner component doesn't export health information
		return uc.health
	}

	return hc.CurrentHealth()
}

// SetHealth sets the health of the userComponent, which represents the health
// of the component at a graph level. This will be merged with the health of
// the inner component (the health of the business logic) to determine overall
// health.
func (uc *userComponent) SetHealth(h component.Health) {
	uc.mut.Lock()
	defer uc.mut.Unlock()
	uc.health = h
}

// setComponentError is a convenience function which sets the component to
// unhealthy using err as the error message.
func setComponentError(uc *userComponent, err error) {
	uc.SetHealth(component.Health{
		Health:     component.HealthTypeUnhealthy,
		Message:    err.Error(),
		UpdateTime: time.Now(),
	})
}
