package controller

import (
	"fmt"
	"strings"

	"github.com/grafana/river/ast"
)

type ComponentNodeManager struct {
	globals      ComponentGlobals
	componentReg ComponentRegistry
	scope        *Scope
}

type getCustomComponentConfig func(namespace string, componentName string) (*Declare, *Scope)

// NewComponentNodeManager creates a new ComponentNodeManager.
func NewComponentNodeManager(globals ComponentGlobals, componentReg ComponentRegistry) *ComponentNodeManager {
	return &ComponentNodeManager{
		globals:      globals,
		componentReg: componentReg,
	}
}

// CreateComponentNode creates a new builtin component or a new custom component.
func (m *ComponentNodeManager) createComponentNode(componentName string, block *ast.BlockStmt) (ComponentNode, error) {
	namespace := strings.Split(componentName, ".")[0]
	if isCustomComponent(m.scope, namespace) {
		return NewCustomComponentNode(m.globals, block, m.getCustomComponentConfig), nil
	} else {
		registration, exists := m.componentReg.Get(componentName)
		if !exists {
			return nil, fmt.Errorf("unrecognized component name %q", componentName)
		}
		return NewBuiltinComponentNode(m.globals, registration, block), nil
	}
}

func (m *ComponentNodeManager) getCustomComponentConfig(namespace string, componentName string) (*Declare, *Scope) {
	var (
		template *Declare
		scope    *Scope
	)

	if namespace == "" {
		template, scope = findLocalDeclare(m.scope, componentName)
	} else {
		template, scope = findImportedDeclare(m.scope, namespace, componentName)
	}

	if scope == nil || template == nil {
		return nil, nil
	}
	return template, scope.DeepCopy()
}

func isCustomComponent(scope *Scope, namespace string) bool {
	if scope == nil {
		return false
	}
	_, declareExists := scope.declares[namespace]
	_, importExists := scope.imports[namespace]
	return declareExists || importExists || isCustomComponent(scope.parent, namespace)
}

func findLocalDeclare(scope *Scope, componentName string) (*Declare, *Scope) {
	if declare, ok := scope.declares[componentName]; ok {
		return declare, scope
	}
	if scope.parent != nil {
		return findLocalDeclare(scope.parent, componentName)
	}
	return nil, nil
}

func findImportedDeclare(scope *Scope, namespace string, componentName string) (*Declare, *Scope) {
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
