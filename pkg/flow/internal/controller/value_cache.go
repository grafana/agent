package controller

import (
	"sync"

	"github.com/grafana/agent/component"
	"github.com/grafana/agent/pkg/river/vm"
)

// valueCache caches component arguments and exports to expose as variables for
// River expressions.
//
// The current state of valueCache can then be built into a *vm.Scope for other
// components to be evaluated.
type valueCache struct {
	mut           sync.RWMutex
	components    map[string]ComponentID // NodeID -> ComponentID
	args          map[string]interface{} // NodeID -> component arguments value
	exports       map[string]interface{} // NodeID -> component exports value
	moduleExports map[string]any         // name -> value for the value of module exports
}

// newValueCache cretes a new ValueCache.
func newValueCache() *valueCache {
	return &valueCache{
		components:    make(map[string]ComponentID),
		args:          make(map[string]interface{}),
		exports:       make(map[string]interface{}),
		moduleExports: make(map[string]any),
	}
}

// CacheArguments will cache the provided arguments by the given id. args may
// be nil to store an empty object.
func (vc *valueCache) CacheArguments(id ComponentID, args component.Arguments) {
	vc.mut.Lock()
	defer vc.mut.Unlock()

	nodeID := id.String()
	vc.components[nodeID] = id

	var argsVal interface{} = make(map[string]interface{})
	if args != nil {
		argsVal = args
	}
	vc.args[nodeID] = argsVal
}

// CacheExports will cache the provided exports using the given id. exports may
// be nil to store an empty object.
func (vc *valueCache) CacheExports(id ComponentID, exports component.Exports) {
	vc.mut.Lock()
	defer vc.mut.Unlock()

	nodeID := id.String()
	vc.components[nodeID] = id

	var exportsVal interface{} = make(map[string]interface{})
	if exports != nil {
		exportsVal = exports
	}
	vc.exports[nodeID] = exportsVal
}

// CacheModuleExportValue saves the value to the map
func (vc *valueCache) CacheModuleExportValue(name string, value any) {
	vc.mut.Lock()
	defer vc.mut.Unlock()

	vc.moduleExports[name] = value
}

// CreateModuleExports creates a map for usage on OnExportsChanged
func (vc *valueCache) CreateModuleExports() map[string]any {
	vc.mut.RLock()
	defer vc.mut.RUnlock()

	exports := make(map[string]any)
	for k, v := range vc.moduleExports {
		exports[k] = v
	}
	return exports
}

func (vc *valueCache) ClearModuleExports() {
	vc.mut.Lock()
	defer vc.mut.Unlock()

	vc.moduleExports = make(map[string]any)
}

// SyncIDs will removed any cached values for any Component ID which is not in
// ids. SyncIDs should be called with the current set of components after the
// graph is updated.
func (vc *valueCache) SyncIDs(ids []ComponentID) {
	expectMap := make(map[string]ComponentID, len(ids))
	for _, id := range ids {
		expectMap[id.String()] = id
	}

	vc.mut.Lock()
	defer vc.mut.Unlock()

	for id := range vc.components {
		if _, keep := expectMap[id]; keep {
			continue
		}
		delete(vc.components, id)
		delete(vc.args, id)
		delete(vc.exports, id)
	}
}

// BuildContext builds a vm.Scope based on the current set of cached values.
// The arguments and exports for the same ID are merged into one object.
func (vc *valueCache) BuildContext(parent *vm.Scope) *vm.Scope {
	vc.mut.RLock()
	defer vc.mut.RUnlock()

	scope := &vm.Scope{
		Parent:    parent,
		Variables: make(map[string]interface{}),
	}

	// First, partition components by River block name.
	var componentsByBlockName = make(map[string][]ComponentID)
	for _, id := range vc.components {
		blockName := id[0]
		componentsByBlockName[blockName] = append(componentsByBlockName[blockName], id)
	}

	// Then, convert each partition into a single value.
	for blockName, ids := range componentsByBlockName {
		scope.Variables[blockName] = vc.buildValue(ids, 1)
	}

	return scope
}

// buildValue recursively converts the set of user components into a single
// value. offset is used to determine which element in the userComponentName
// we're looking at.
func (vc *valueCache) buildValue(from []ComponentID, offset int) interface{} {
	// We can't recurse anymore; return the node directly.
	if len(from) == 1 && offset >= len(from[0]) {
		name := from[0].String()

		// TODO(rfratto): should we allow arguments to be returned so users can
		// reference arguments as well as exports?
		exports, ok := vc.exports[name]
		if !ok {
			exports = make(map[string]interface{})
		}
		return exports
	}

	attrs := make(map[string]interface{})

	// First, partition the components by their label.
	var componentsByLabel = make(map[string][]ComponentID)
	for _, id := range from {
		blockName := id[offset]
		componentsByLabel[blockName] = append(componentsByLabel[blockName], id)
	}

	// Then, convert each partition into a single value.
	for label, ids := range componentsByLabel {
		attrs[label] = vc.buildValue(ids, offset+1)
	}
	return attrs
}
