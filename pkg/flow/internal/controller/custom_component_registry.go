package controller

import (
	"fmt"

	"github.com/grafana/river/ast"
)

// CustomComponentRegistry holds custom component definitions that are available in the context.
// The definitions are either imported, declared locally, or declared in a parent registry.
// Imported definitions are stored inside of the corresponding import registry.
type CustomComponentRegistry struct {
	parent   *CustomComponentRegistry            // nil if root config
	declares map[string]ast.Body                 // customComponentName: template
	imports  map[string]*CustomComponentRegistry // importNamespace: importScope
}

// NewCustomComponentRegistry creates a new CustomComponentRegistry with a parent.
// parent must be nil or *Scope.
func NewCustomComponentRegistry(parent any) *CustomComponentRegistry {
	s := &CustomComponentRegistry{
		declares: make(map[string]ast.Body),
		imports:  make(map[string]*CustomComponentRegistry),
	}
	if parent != nil {
		s.parent = parent.(*CustomComponentRegistry)
	}
	return s
}

// registerDeclare stores a local declare block.
func (s *CustomComponentRegistry) registerDeclare(declare *ast.BlockStmt) {
	s.declares[declare.Label] = declare.Body
}

// registerImport stores the label of the import.
// The content will be added later during evaluation.
// It's important to register it before populating the component nodes
// (else we don't know which one exists).
func (s *CustomComponentRegistry) registerImport(importNamespace string) {
	s.imports[importNamespace] = nil
}

// updateImportContent updates the content of a registered import.
// The content of an import node can contain other import blocks.
// These are considered as "children" of the root import node.
// Each child has its own CustomComponentRegistry which needs to be updated.
func (s *CustomComponentRegistry) updateImportContent(importNode *ImportConfigNode) error {
	if _, exist := s.imports[importNode.label]; !exist {
		return fmt.Errorf("import %q was not registered", importNode.label)
	}
	importScope := NewCustomComponentRegistry(nil)
	importScope.declares = importNode.importedDeclares
	importScope.updateImportContentChildren(importNode)
	s.imports[importNode.label] = importScope
	return nil
}

// updateImportContentChildren recurse through the children of an import node
// and update their scope with the imported declare blocks.
func (s *CustomComponentRegistry) updateImportContentChildren(importNode *ImportConfigNode) {
	for _, child := range importNode.ImportConfigNodesChildren() {
		childScope := NewCustomComponentRegistry(nil)
		childScope.declares = child.importedDeclares
		childScope.updateImportContentChildren(child)
		s.imports[child.label] = childScope
	}
}

// DeepCopy returns a deep copy of the full scope (including parents and imports).
func (s *CustomComponentRegistry) DeepCopy() *CustomComponentRegistry {
	newScope := NewCustomComponentRegistry(nil)

	if s.parent != nil {
		newScope.parent = s.parent.DeepCopy()
	}

	for k, v := range s.declares {
		if v != nil {
			newScope.declares[k] = v
		}
	}

	for k, v := range s.imports {
		if v != nil {
			newScope.imports[k] = v.DeepCopy()
		}
	}

	return newScope
}
