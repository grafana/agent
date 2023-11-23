package controller

import (
	"fmt"
	"reflect"
	"strings"
	"sync"

	"github.com/grafana/agent/component"
	"github.com/grafana/river/vm"
)

// valueCache caches component arguments and exports to expose as variables for
// River expressions.
//
// The current state of valueCache can then be built into a *vm.Scope for other
// components to be evaluated.
type valueCache struct {
	mut                sync.RWMutex
	components         map[string]ComponentID    // NodeID -> ComponentID
	args               map[string]interface{}    // NodeID -> component arguments value
	exports            map[string]interface{}    // NodeID -> component exports value
	moduleArguments    map[string]any            // key -> module arguments value
	moduleExports      map[string]any            // name -> value for the value of module exports
	declareValues      map[string]map[string]any // DeclareComponentNodeID -> arguments
	declareExports     map[string]any            // NodeID of an Export node associated with a DeclareComponentNode -> value
	namespaces         map[string]string         // NodeID -> Namespace
	moduleChangedIndex int                       // Everytime a change occurs this is incremented
}

// newValueCache creates a new ValueCache.
func newValueCache() *valueCache {
	return &valueCache{
		components:      make(map[string]ComponentID),
		args:            make(map[string]interface{}),
		exports:         make(map[string]interface{}),
		moduleArguments: make(map[string]any),
		declareValues:   make(map[string]map[string]any),
		declareExports:  make(map[string]any),
		moduleExports:   make(map[string]any),
		namespaces:      make(map[string]string),
	}
}

func (vc *valueCache) CacheDeclare(nodeID string, arguments component.Arguments) error {
	vc.mut.Lock()
	defer vc.mut.Unlock()

	exportMap, ok := arguments.(map[string]any)
	if !ok {
		return fmt.Errorf("error retrieving arguments of %s", nodeID)
	}

	vc.declareValues[nodeID] = make(map[string]any)
	for key, value := range exportMap {
		vc.declareValues[nodeID][key] = value
	}
	return nil
}

func (vc *valueCache) CacheNamespace(nodeID string, namespace string) {
	vc.mut.Lock()
	defer vc.mut.Unlock()

	vc.namespaces[nodeID] = namespace
}

func (vc *valueCache) CacheDeclareExport(nodeID string, value any) {
	vc.mut.Lock()
	defer vc.mut.Unlock()

	vc.declareExports[nodeID] = value
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

// CacheModuleArgument will cache the provided exports using the given id.
func (vc *valueCache) CacheModuleArgument(key string, value any) {
	vc.mut.Lock()
	defer vc.mut.Unlock()

	if value == nil {
		vc.moduleArguments[key] = nil
	} else {
		vc.moduleArguments[key] = value
	}
}

// CacheModuleExportValue saves the value to the map
func (vc *valueCache) CacheModuleExportValue(name string, value any) {
	vc.mut.Lock()
	defer vc.mut.Unlock()

	// Need to see if the module exports have changed.
	v, found := vc.moduleExports[name]
	if !found {
		vc.moduleChangedIndex++
	} else if !reflect.DeepEqual(v, value) {
		vc.moduleChangedIndex++
	}

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

// ClearModuleExports empties the map and notifies that the exports have changed.
func (vc *valueCache) ClearModuleExports() {
	vc.mut.Lock()
	defer vc.mut.Unlock()

	vc.moduleChangedIndex++
	vc.moduleExports = make(map[string]any)
}

// ExportChangeIndex return the change index.
func (vc *valueCache) ExportChangeIndex() int {
	vc.mut.RLock()
	defer vc.mut.RUnlock()

	return vc.moduleChangedIndex
}

// SyncIDs will remove any cached values for any Component ID which is not in
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

// SyncModuleArgs will remove any cached values for any args no longer in the map.
func (vc *valueCache) SyncModuleArgs(args map[string]any) {
	vc.mut.Lock()
	defer vc.mut.Unlock()

	for id := range vc.moduleArguments {
		if _, keep := args[id]; keep {
			continue
		}
		delete(vc.moduleArguments, id)
	}
}

func (vc *valueCache) SyncDeclareIDs(ids map[string]struct{}) {
	vc.mut.Lock()
	defer vc.mut.Unlock()

	for id := range vc.declareValues {
		if _, keep := ids[id]; keep {
			continue
		}
		delete(vc.declareValues, id)
	}
}

func (vc *valueCache) SyncNamespaces(ids map[string]struct{}) {
	vc.mut.Lock()
	defer vc.mut.Unlock()

	for id := range vc.namespaces {
		if _, keep := ids[id]; keep {
			continue
		}
		delete(vc.namespaces, id)
	}
}

func (vc *valueCache) SyncDeclareExportIDs(ids map[string]struct{}) {
	vc.mut.Lock()
	defer vc.mut.Unlock()

	for id := range vc.declareExports {
		if _, keep := ids[id]; keep {
			continue
		}
		delete(vc.declareExports, id)
	}
}

// BuildContext builds a vm.Scope based on the current set of cached values.
// The arguments and exports for the same ID are merged into one object.
func (vc *valueCache) BuildContext(n BlockNode) *vm.Scope {
	vc.mut.RLock()
	defer vc.mut.RUnlock()

	scope := &vm.Scope{
		Parent:    nil,
		Variables: make(map[string]interface{}),
	}

	nodeNamespace := n.Namespace()

	// First, partition components by River block name.
	var componentsByBlockName = make(map[string][]ComponentID)
	for nodeID, id := range vc.components {
		// only access components in the same namespace
		if vc.namespaces[nodeID] == nodeNamespace {
			trimmedComponentID := removeSequentialPrefix(nodeNamespace, id)
			blockName := trimmedComponentID[0]
			componentsByBlockName[blockName] = append(componentsByBlockName[blockName], trimmedComponentID)
		}
	}

	// Then, convert each partition into a single value.
	for blockName, ids := range componentsByBlockName {
		scope.Variables[blockName] = vc.buildValue(nodeNamespace, ids, 1)
	}

	// Add module arguments to the scope.
	if len(vc.moduleArguments) > 0 {
		scope.Variables["argument"] = make(map[string]any)
	}
	for key, value := range vc.moduleArguments {
		keyMap := make(map[string]any)
		keyMap["value"] = value

		switch args := scope.Variables["argument"].(type) {
		case map[string]any:
			args[key] = keyMap
		}
	}

	for key, value := range vc.declareExports {
		// we trim the namespace that they have in common here. This is needed for nested declares
		trimmedKey := strings.TrimPrefix(key, n.Namespace()+".")
		convertToNestedMap(trimmedKey, value, scope.Variables)
	}

	// add arguments available in the namespace
	if n.Namespace() != "" {
		if valueMap, exists := vc.declareValues[n.Namespace()]; exists {
			if len(valueMap) > 0 {
				scope.Variables["argument"] = make(map[string]any)
			}
			for key, value := range valueMap {
				keyMap := make(map[string]any)
				keyMap["value"] = value

				switch args := scope.Variables["argument"].(type) {
				case map[string]any:
					args[key] = keyMap
				}
			}
		}
	}

	return scope
}

func removeSequentialPrefix(prefix string, arr []string) []string {
	prefixArr := strings.Split(prefix, ".")

	prefixLen := 0
	for i, part := range prefixArr {
		if i < len(arr) && arr[i] == part {
			prefixLen++
		} else {
			break
		}
	}
	return arr[prefixLen:]
}

func convertToNestedMap(key string, value any, rootMap map[string]any) {
	parts := strings.Split(key, ".")

	currentMap := rootMap
	for i := 0; i < len(parts); i++ {
		part := parts[i]

		// We want to ignore the type export of the node.
		// For example if the key is add.example.export.sum, we want to consider it as add.example.sum
		// We check the length to allow add.example.export.export to be referred as add.example.export
		// This is not a very nice trick.
		if i == len(parts)-2 && part == "export" {
			continue
		}

		if i == len(parts)-1 {
			currentMap[part] = value
		} else {
			if _, exists := currentMap[part]; !exists {
				currentMap[part] = make(map[string]any)
			}
			currentMap = currentMap[part].(map[string]any)
		}
	}
}

// buildValue recursively converts the set of user components into a single
// value. offset is used to determine which element in the userComponentName
// we're looking at.
func (vc *valueCache) buildValue(namespace string, from []ComponentID, offset int) interface{} {
	// We can't recurse anymore; return the node directly.
	if len(from) == 1 && offset >= len(from[0]) {
		name := from[0].String()

		if namespace != "" {
			name = namespace + "." + name
		}

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
		attrs[label] = vc.buildValue(namespace, ids, offset+1)
	}
	return attrs
}
