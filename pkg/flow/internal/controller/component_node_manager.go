package controller

import (
	"fmt"
	"strings"

	"github.com/grafana/river/ast"
)

// ComponentNodeManager is responsible for creating new component nodes and
// obtaining the necessary information to run them.
type ComponentNodeManager struct {
	globals ComponentGlobals
	// componentReg returns information to build and run built-in components.
	componentReg ComponentRegistry
	// customComponentReg returns information to build and run custom components.
	customComponentReg *CustomComponentRegistry
}

type getCustomComponentConfig func(namespace string, componentName string) (ast.Body, *CustomComponentRegistry, error)

// NewComponentNodeManager creates a new ComponentNodeManager without custom component registry.
func NewComponentNodeManager(globals ComponentGlobals, componentReg ComponentRegistry) *ComponentNodeManager {
	return &ComponentNodeManager{
		globals:      globals,
		componentReg: componentReg,
	}
}

// CreateComponentNode creates a new builtin component or a new custom component.
func (m *ComponentNodeManager) createComponentNode(componentName string, block *ast.BlockStmt) (ComponentNode, error) {
	// firstPart may correspond either to a namespace or to a componentName for custom components
	firstPart := strings.Split(componentName, ".")[0]
	if isCustomComponent(m.customComponentReg, firstPart) {
		return NewCustomComponentNode(m.globals, block, m.getCustomComponentConfig), nil
	}
	registration, exists := m.componentReg.Get(componentName)
	if !exists {
		return nil, fmt.Errorf("unrecognized component name %q", componentName)
	}
	return NewBuiltinComponentNode(m.globals, registration, block), nil
}

// getCustomComponentConfig is used by the custom component to retrieve its template and the customComponentRegistry associated with it.
func (m *ComponentNodeManager) getCustomComponentConfig(namespace string, componentName string) (ast.Body, *CustomComponentRegistry, error) {
	var (
		template                ast.Body
		customComponentRegistry *CustomComponentRegistry
	)

	if namespace == "" {
		template, customComponentRegistry = findLocalDeclare(m.customComponentReg, componentName)
	} else {
		template, customComponentRegistry = findImportedDeclare(m.customComponentReg, namespace, componentName)
	}

	if customComponentRegistry == nil || template == nil {
		return nil, nil, fmt.Errorf("custom component config not found in the registry, namespace: %s, componentName: %s", namespace, componentName)
	}
	return template, customComponentRegistry.deepCopy(), nil
}

// isCustomComponent returns true if the name matches a declare in the provided custom component registry.
func isCustomComponent(reg *CustomComponentRegistry, name string) bool {
	if reg == nil {
		return false
	}
	_, declareExists := reg.declares[name]
	_, importExists := reg.imports[name]
	return declareExists || importExists || isCustomComponent(reg.parent, name)
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

// findImportedDeclare recursively searches for an import matching the provided namespace.
// When the import is found, it will search for a declare matching the componentName within the custom registry of the import.
func findImportedDeclare(reg *CustomComponentRegistry, namespace string, componentName string) (ast.Body, *CustomComponentRegistry) {
	if imported, ok := reg.imports[namespace]; ok {
		if declare, ok := imported.declares[componentName]; ok {
			return declare, imported
		}
	}
	if reg.parent != nil {
		return findImportedDeclare(reg.parent, namespace, componentName)
	}
	return nil, nil
}
