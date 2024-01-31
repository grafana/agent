package controller

import (
	"fmt"

	"github.com/grafana/agent/component"
	"github.com/grafana/agent/pkg/flow/internal/dag"
	"github.com/grafana/river/ast"
)

// ComponentRegistry is a collection of components.
type ComponentRegistry interface {
	// Get looks up a component by name.
	Get(name string) (Component, bool)
}

// Component is a generic representation of a component.
type Component struct {
	kind    ComponentKind
	builtin component.Registration
	custom  CustomComponent
}

// Kind returns the Kind of component c is.
func (c *Component) Kind() ComponentKind { return c.kind }

// Builtin returns the registration for a built-in component. Builtin panics if
// Kind() is not ComponentKindBuiltin.
func (c *Component) Builtin() component.Registration {
	if c.kind != ComponentKindBuiltin {
		panic("Component.Builtin: component is not a builtin component")
	}
	return c.builtin
}

// Custom returns the custom component. Custom panics if Kind() is not
// ComponentKindCustom.
func (c *Component) Custom() CustomComponent {
	if c.kind != ComponentKindCustom {
		panic("Component.Builtin: component is not a custom component")
	}
	return c.custom
}

// ComponentKind represents a kind of component.
type ComponentKind int

const (
	ComponentKindInvalid ComponentKind = iota // ComponentKindInvalid is an invalid ComponentKind.
	ComponentKindBuiltin                      // ComponentKindBuiltin is a built-in component.
	ComponentKindCustom                       // ComponentKindCustom is a custom component.
)

// String returns the string form of the ComponentKind.
func (kind ComponentKind) String() string {
	switch kind {
	case ComponentKindInvalid:
		return "invalid"
	case ComponentKindBuiltin:
		return "builtin"
	case ComponentKindCustom:
		return "custom"
	default:
		return fmt.Sprintf("ComponentKind(%d)", kind)
	}
}

// CustomComponent represents the definition of a custom component either
// through a declare statment or an import.
type CustomComponent interface {
	// Definition retrieves the definition for a CustomComponent.
	//
	// Definition may lazily retrieve a component definition from an imported
	// module. If the custom component doesn't exist in the imported module,
	// or the imported module hasn't been evaluated yet, Definition returns an
	// error.
	Definition() (ast.Body, error)
}

// DefaultComponentRegistry is the default [ComponentRegistry] which only gets
// builtin components registered to github.com/grafana/agent/component.
type DefaultComponentRegistry struct{}

// Get retrieves a component using [component.Get].
func (reg DefaultComponentRegistry) Get(name string) (Component, bool) {
	builtinReg, ok := component.Get(name)
	if !ok {
		return Component{}, false
	}

	return Component{
		kind:    ComponentKindBuiltin,
		builtin: builtinReg,
	}, true
}

// RegistryMap is a map which implements [ComponentRegistry].
type RegistryMap map[string]component.Registration

// Get retrieves a component using [component.Get].
func (m RegistryMap) Get(name string) (Component, bool) {
	reg, ok := m[name]
	if !ok {
		return Component{}, false
	}

	return Component{
		kind:    ComponentKindBuiltin,
		builtin: reg,
	}, true
}

// customComponentRegistry looks up custom components defined within a graph,
// falling back to a parent registry if provided.
type customComponentRegistry struct {
	parent ComponentRegistry
	graph  *dag.Graph
}

func (reg *customComponentRegistry) Get(name string) (Component, bool) {
	// First look for a custom component.
	if c, ok := reg.getCustomComponent(name); ok {
		return c, ok
	}

	// Fall back to the parent registry if it exists.
	if reg.parent == nil {
		return Component{}, false
	}
	return reg.parent.Get(name)
}

func (reg *customComponentRegistry) getCustomComponent(name string) (Component, bool) {
	node := reg.graph.GetByID("declare." + name)
	switch node := node.(type) {
	case *DeclareNode:
		return Component{
			kind:   ComponentKindCustom,
			custom: node,
		}, true

	default:
		return Component{}, false
	}
}
