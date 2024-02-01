package controller

import (
	"fmt"

	"github.com/grafana/river/ast"
)

type Scope struct {
	parent   *Scope
	declares map[string]ast.Body // customComponentName: template
	imports  map[string]*Scope   // importNamespace: importScope
}

func NewScope(parent any) *Scope {
	s := &Scope{
		declares: make(map[string]ast.Body),
		imports:  make(map[string]*Scope),
	}
	if parent != nil {
		s.parent = parent.(*Scope)
	}
	return s
}

func (s *Scope) DeepCopy() *Scope {
	newScope := NewScope(nil)

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

func (s *Scope) registerDeclare(declare *ast.BlockStmt) {
	s.declares[declare.Label] = declare.Body
}

func (s *Scope) registerImport(importLabel string) {
	s.imports[importLabel] = nil
}

func (s *Scope) updateImportContent(importNode *ImportConfigNode) error {
	if _, exist := s.imports[importNode.label]; !exist {
		return fmt.Errorf("import %q was not registered", importNode.label)
	}
	importScope := NewScope(nil)
	importScope.declares = importNode.importedDeclares
	importScope.updateImportContentChildren(importNode)
	s.imports[importNode.label] = importScope
	return nil
}

func (s *Scope) updateImportContentChildren(importNode *ImportConfigNode) {
	for _, child := range importNode.ImportConfigNodesChildren() {
		childScope := NewScope(nil)
		childScope.declares = child.importedDeclares
		childScope.updateImportContentChildren(child)
		s.imports[child.label] = childScope
	}
}
