package controller

import (
	"fmt"
	"sync"

	"github.com/grafana/agent/component"
	"github.com/hashicorp/hcl/v2"
	"github.com/rfratto/gohcl"
	"github.com/zclconf/go-cty/cty"
)

// valueCache caches component arguments and exports evaluated as cty.Values.
//
// The current state of valueCache can then be built into a *hcl.EvalContext
// for other components to be evaluated.
type valueCache struct {
	mut        sync.RWMutex
	components map[string]ComponentID // NodeID -> ComponentID
	args       map[string]cty.Value   // NodeID -> cty.Value of component arguments
	exports    map[string]cty.Value   // NodeID -> cty.Value of component exports
}

// newValueCache cretes a new ValueCache.
func newValueCache() *valueCache {
	return &valueCache{
		components: make(map[string]ComponentID),
		args:       make(map[string]cty.Value),
		exports:    make(map[string]cty.Value),
	}
}

// CacheArguments will cache the provided arguments by the given id. args may
// be nil to store an empty object.
func (vc *valueCache) CacheArguments(id ComponentID, args component.Arguments) {
	vc.mut.Lock()
	defer vc.mut.Unlock()

	nodeID := id.String()
	vc.components[nodeID] = id

	argsVal := cty.EmptyObjectVal
	if args != nil {
		ty, err := gohcl.ImpliedType(args)
		if err != nil {
			panic(err)
		}
		cv, err := gohcl.ToCtyValue(args, ty)
		if err != nil {
			panic(err)
		}
		argsVal = cv
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

	exportsVal := cty.EmptyObjectVal
	if exports != nil {
		ty, err := gohcl.ImpliedType(exports)
		if err != nil {
			panic(err)
		}
		cv, err := gohcl.ToCtyValue(exports, ty)
		if err != nil {
			panic(err)
		}
		exportsVal = cv
	}

	vc.exports[nodeID] = exportsVal
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

// BuildContext builds an hcl.EvalContext based on the current set of cached
// values. The arguments and exports for the same ID are merged into one
// object.
func (vc *valueCache) BuildContext(parent *hcl.EvalContext) *hcl.EvalContext {
	var ectx *hcl.EvalContext
	if parent == nil {
		ectx = parent.NewChild()
	} else {
		ectx = &hcl.EvalContext{}
	}

	// Variables is used to build the mapping of referenceable values. See
	// value_cache_test.go for examples of what the expected output is.
	ectx.Variables = make(map[string]cty.Value)

	// First, partition components by HCL block name.
	var componentsByBlockName = make(map[string][]ComponentID)
	for _, id := range vc.components {
		blockName := id[0]
		componentsByBlockName[blockName] = append(componentsByBlockName[blockName], id)
	}

	// Then, convert each partition into a single value.
	for blockName, ids := range componentsByBlockName {
		ectx.Variables[blockName] = vc.buildValue(ids, 1)
	}

	return ectx
}

// buildValue recursively converts the set of user components into a single
// cty.Value. offset is used to determine which element in the
// userComponentName we're looking at.
func (vc *valueCache) buildValue(from []ComponentID, offset int) cty.Value {
	// We can't recurse anymore; return the node directly.
	if len(from) == 1 && offset >= len(from[0]) {
		name := from[0].String()

		cfg, ok := vc.args[name]
		if !ok {
			cfg = cty.EmptyObjectVal
		}
		exports, ok := vc.exports[name]
		if !ok {
			exports = cty.EmptyObjectVal
		}

		return mergeComponentValues(cfg, exports)
	}

	attrs := make(map[string]cty.Value)

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

	return cty.ObjectVal(attrs)
}

// mergeComponentValues merges a component's config and exports. mergeState
// panics if a key exits in both inputs and store or if neither argument is an
// object.
func mergeComponentValues(args, exports cty.Value) cty.Value {
	if !args.Type().IsObjectType() {
		panic("component arguments must be object type")
	}
	if !exports.Type().IsObjectType() {
		panic("component exports must be object type")
	}

	var (
		inputMap = args.AsValueMap()
		stateMap = exports.AsValueMap()
	)

	mergedMap := make(map[string]cty.Value, len(inputMap)+len(stateMap))
	for key, value := range inputMap {
		mergedMap[key] = value
	}
	for key, value := range stateMap {
		if _, exist := mergedMap[key]; exist {
			panic(fmt.Sprintf("component exports overrides arguments key %s", key))
		}
		mergedMap[key] = value
	}

	return cty.ObjectVal(mergedMap)
}
