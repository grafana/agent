package controller

import "fmt"

type Scope struct {
	parent   *Scope
	declares map[string]*Declare // customComponentName: template
	imports  map[string]*Scope   // importNamespace: importScope
}

func NewScope(parent *Scope) *Scope {
	return &Scope{
		parent:   parent,
		declares: make(map[string]*Declare),
		imports:  make(map[string]*Scope),
	}
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

func (s *Scope) registerDeclare(declare *Declare) {
	s.declares[declare.block.Label] = declare
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
