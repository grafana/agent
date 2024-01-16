package flow

import (
	"fmt"

	"github.com/grafana/agent/component"
	"github.com/grafana/agent/pkg/flow/internal/controller"
	"github.com/grafana/agent/pkg/flow/internal/dag"
)

// GetComponent implements [component.Provider].
func (f *Flow) GetComponent(id component.ID, opts component.InfoOptions) (*component.Info, error) {
	f.loadMut.RLock()
	defer f.loadMut.RUnlock()

	if id.ModuleID != "" {
		mod, ok := f.modules.Get(id.ModuleID)
		if !ok {
			return nil, component.ErrComponentNotFound
		}

		return mod.f.GetComponent(component.ID{LocalID: id.LocalID}, opts)
	}

	graph := f.loader.OriginalGraph()

	node := graph.GetByID(id.LocalID)
	if node == nil {
		return nil, component.ErrComponentNotFound
	}

	cn, ok := node.(controller.NodeWithComponent)
	if !ok {
		return nil, fmt.Errorf("%q is not a ComponentNode", id)
	}

	return f.getComponentDetail(cn, graph, opts), nil
}

// ListComponents implements [component.Provider].
func (f *Flow) ListComponents(moduleID string, opts component.InfoOptions) ([]*component.Info, error) {
	f.loadMut.RLock()
	defer f.loadMut.RUnlock()

	if moduleID != "" {
		mod, ok := f.modules.Get(moduleID)
		if !ok {
			return nil, component.ErrModuleNotFound
		}

		return mod.f.ListComponents("", opts)
	}

	var (
		components = f.loader.Components()
		imports    = f.loader.Imports()
		graph      = f.loader.OriginalGraph()
	)

	detail := make([]*component.Info, 0, len(components)+len(imports))
	for _, component := range components {
		detail = append(detail, f.getComponentDetail(component, graph, opts))
	}
	for _, importNode := range imports {
		detail = append(detail, f.getComponentDetail(importNode, graph, opts))
	}
	return detail, nil
}

func (f *Flow) getComponentDetail(cn controller.NodeWithComponent, graph *dag.Graph, opts component.InfoOptions) *component.Info {
	var references, referencedBy []string

	// Skip over any edge which isn't between two component nodes. This is a
	// temporary workaround needed until there's athe concept of configuration
	// blocks in the API.
	//
	// Without this change, the graph fails to render when a configuration
	// block is referenced in the graph.
	//
	// TODO(rfratto): add support for config block nodes in the API and UI.
	for _, dep := range graph.Dependencies(cn) {
		if _, ok := dep.(controller.NodeWithComponent); ok {
			references = append(references, dep.NodeID())
		}
	}
	for _, dep := range graph.Dependants(cn) {
		if _, ok := dep.(controller.NodeWithComponent); ok {
			referencedBy = append(referencedBy, dep.NodeID())
		}
	}

	// Fields which are optional to set.
	var (
		health    component.Health
		arguments component.Arguments
		exports   component.Exports
		debugInfo interface{}
	)

	if opts.GetHealth {
		health = cn.CurrentHealth()
	}
	if opts.GetArguments {
		arguments = cn.Arguments()
	}
	if opts.GetExports {
		exports = cn.Exports()
	}
	if opts.GetDebugInfo {
		debugInfo = cn.DebugInfo()
	}

	return &component.Info{
		Component: cn.Component(),
		ModuleIDs: cn.ModuleIDs(),

		ID: component.ID{
			ModuleID: f.opts.ControllerID,
			LocalID:  cn.NodeID(),
		},
		Label: cn.Label(),

		References:   references,
		ReferencedBy: referencedBy,

		BlockName: cn.BlockName(),
		Health:    health,

		Arguments: arguments,
		Exports:   exports,
		DebugInfo: debugInfo,
	}
}
