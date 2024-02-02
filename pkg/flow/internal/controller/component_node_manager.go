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
	// ComponentRegistry returns information to build and run built-in components.
	componentReg ComponentRegistry
	// Scope returns information to build and run custom components.
	scope *CustomComponentRegistry
}

type getCustomComponentConfig func(namespace string, componentName string) (ast.Body, *CustomComponentRegistry, error)

// NewComponentNodeManager creates a new ComponentNodeManager without Scope.
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
	if isCustomComponent(m.scope, firstPart) {
		return NewCustomComponentNode(m.globals, block, m.getCustomComponentConfig), nil
	}
	registration, exists := m.componentReg.Get(componentName)
	if !exists {
		return nil, fmt.Errorf("unrecognized component name %q", componentName)
	}
	return NewBuiltinComponentNode(m.globals, registration, block), nil
}

// getCustomComponentConfig is used by the custom component to retrieve its template and the scope associated with it.
func (m *ComponentNodeManager) getCustomComponentConfig(namespace string, componentName string) (ast.Body, *CustomComponentRegistry, error) {
	var (
		template ast.Body
		scope    *CustomComponentRegistry
	)

	if namespace == "" {
		template, scope = findLocalDeclare(m.scope, componentName)
	} else {
		template, scope = findImportedDeclare(m.scope, namespace, componentName)
	}

	if scope == nil || template == nil {
		return nil, nil, fmt.Errorf("custom component config not found in the registry, namespace: %s, componentName: %s", namespace, componentName)
	}
	return template, scope.DeepCopy(), nil
}

// isCustomComponent returns true if the name matches a declare in the provided scope.
func isCustomComponent(scope *CustomComponentRegistry, name string) bool {
	if scope == nil {
		return false
	}
	_, declareExists := scope.declares[name]
	_, importExists := scope.imports[name]
	return declareExists || importExists || isCustomComponent(scope.parent, name)
}

// findLocalDeclare recursively searches for a declare definition in the scope.
func findLocalDeclare(scope *CustomComponentRegistry, componentName string) (ast.Body, *CustomComponentRegistry) {
	if declare, ok := scope.declares[componentName]; ok {
		return declare, scope
	}
	if scope.parent != nil {
		return findLocalDeclare(scope.parent, componentName)
	}
	return nil, nil
}

// findImportedDeclare recursively searches for an import matching the provided namespace.
// When the import is found, it will search for a declare matching the componentName within the Scope of the import.
func findImportedDeclare(scope *CustomComponentRegistry, namespace string, componentName string) (ast.Body, *CustomComponentRegistry) {
	if imported, ok := scope.imports[namespace]; ok {
		if declare, ok := imported.declares[componentName]; ok {
			return declare, imported
		}
	}
	if scope.parent != nil {
		return findImportedDeclare(scope.parent, namespace, componentName)
	}
	return nil, nil
}
