package controller

import (
	"fmt"
	"sync"

	"github.com/grafana/river/ast"
)

// ComponentNodeManager is responsible for creating new component nodes and
// obtaining the necessary information to run them.
type ComponentNodeManager struct {
	globals ComponentGlobals
	// builtinComponentReg returns information to build and run built-in components.
	builtinComponentReg ComponentRegistry

	mut sync.RWMutex
	// customComponentReg returns information to build and run custom components.
	customComponentReg *CustomComponentRegistry
}

// NewComponentNodeManager creates a new ComponentNodeManager without custom component registry.
func NewComponentNodeManager(globals ComponentGlobals, componentReg ComponentRegistry) *ComponentNodeManager {
	return &ComponentNodeManager{
		globals:             globals,
		builtinComponentReg: componentReg,
	}
}

// CreateComponentNode creates a new builtin component or a new custom component.
func (m *ComponentNodeManager) createComponentNode(componentName string, block *ast.BlockStmt) (ComponentNode, error) {
	if isCustomComponent(m.customComponentReg, block.Name[0]) {
		return NewCustomComponentNode(m.globals, block, m.getCustomComponentConfig), nil
	}
	registration, exists := m.builtinComponentReg.Get(componentName)
	if !exists {
		return nil, fmt.Errorf("cannot retrieve the definition of component name %q", componentName)
	}
	if block.Label == "" {
		return nil, fmt.Errorf("component %q must have a label", componentName)
	}
	return NewBuiltinComponentNode(m.globals, registration, block), nil
}

// getCustomComponentConfig is used by the custom component to retrieve its template and the customComponentRegistry associated with it.
func (m *ComponentNodeManager) getCustomComponentConfig(componentName string) (ast.Body, *CustomComponentRegistry, error) {
	m.mut.Lock()
	defer m.mut.Unlock()

	template, customComponentRegistry := findLocalDeclare(m.customComponentReg, componentName)

	if customComponentRegistry == nil || template == nil {
		return nil, nil, fmt.Errorf("custom component config not found in the registry, componentName: %s", componentName)
	}
	return template, customComponentRegistry, nil
}

// isCustomComponent returns true if the name matches a declare in the provided custom component registry.
func isCustomComponent(reg *CustomComponentRegistry, name string) bool {
	if reg == nil {
		return false
	}
	_, declareExists := reg.declares[name]
	return declareExists || isCustomComponent(reg.parent, name)
}

// findLocalDeclare recursively searches for a declare definition in the custom component registry.
func findLocalDeclare(reg *CustomComponentRegistry, componentName string) (ast.Body, *CustomComponentRegistry) {
	if declare, ok := reg.declares[componentName]; ok {
		return declare, reg
	}
	if reg.parent != nil {
		return findLocalDeclare(reg.parent, componentName)
	}
	return nil, nil
}

func (m *ComponentNodeManager) SetCustomComponentRegistry(reg *CustomComponentRegistry) {
	m.mut.Lock()
	defer m.mut.Unlock()

	m.customComponentReg = reg
}
