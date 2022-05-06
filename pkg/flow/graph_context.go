package flow

import (
	"fmt"
	"sync"

	"github.com/hashicorp/hcl/v2"
	"github.com/rfratto/gohcl"
	"github.com/zclconf/go-cty/cty"
)

// graphContext caches component values across the graph, which can then be
// converted into an *hcl.EvalContext.
type graphContext struct {
	// The parent EvalContext to build a child for.
	parent *hcl.EvalContext

	mut        sync.RWMutex
	components map[string]*userComponent // Component ID -> userComponent
	configs    map[string]cty.Value      // Component ID -> cty.Value for Config
	exports    map[string]cty.Value      // Component ID -> cty.Value for Exports
}

// newGraphContext creates a new graphContext.
func newGraphContext(parent *hcl.EvalContext) *graphContext {
	return &graphContext{
		parent: parent,

		components: make(map[string]*userComponent),
		configs:    make(map[string]cty.Value),
		exports:    make(map[string]cty.Value),
	}
}

// StoreComponent will update the graphContext with a component. updateConfig
// and updateExports can optionally be set to true to cache that component's
// current Config and Exports respectively.
func (gc *graphContext) StoreComponent(uc *userComponent, updateConfig, updateExports bool) {
	gc.mut.Lock()
	defer gc.mut.Unlock()

	componentID := uc.Name().String()

	// Update the components map in case this is the first time we're storing
	// this component.
	gc.components[componentID] = uc

	if updateConfig {
		configVal := cty.EmptyObjectVal

		if cfg := uc.CurrentConfig(); cfg != nil {
			ty, err := gohcl.ImpliedType(cfg)
			if err != nil {
				panic(err)
			}
			cv, err := gohcl.ToCtyValue(cfg, ty)
			if err != nil {
				panic(err)
			}
			configVal = cv
		}

		gc.configs[componentID] = configVal
	}

	if updateExports {
		exportsVal := cty.EmptyObjectVal

		if cfg := uc.CurrentExports(); cfg != nil {
			ty, err := gohcl.ImpliedType(cfg)
			if err != nil {
				panic(err)
			}
			cv, err := gohcl.ToCtyValue(cfg, ty)
			if err != nil {
				panic(err)
			}
			exportsVal = cv
		}

		gc.exports[componentID] = exportsVal
	}
}

// RemoveComponent will remove a stored component and its cached values from
// gc.
func (gc *graphContext) RemoveComponent(uc *userComponent) {
	gc.mut.Lock()
	defer gc.mut.Unlock()

	componentID := uc.Name().String()
	delete(gc.components, componentID)
	delete(gc.configs, componentID)
	delete(gc.exports, componentID)
}

// RemoveStaleComponents will remove cached values for any component not in
// expect.
func (gc *graphContext) RemoveStaleComponents(expect []*userComponent) {
	expectMap := make(map[string]*userComponent, len(expect))
	for _, uc := range expect {
		expectMap[uc.NodeID()] = uc
	}

	for id := range gc.components {
		if _, keep := expectMap[id]; keep {
			continue
		}
		delete(gc.components, id)
		delete(gc.configs, id)
		delete(gc.exports, id)
	}
}

// Build builds an hcl.EvalContext based on the current cached values in the
// graphContext.
func (gc *graphContext) Build() *hcl.EvalContext {
	ectx := gc.parent.NewChild()

	// Variables is used to build the mapping of referenceable values.
	//
	// For the following HCL config file:
	//
	//     foo {
	//       something = true
	//     }
	//
	//     bar "label_a" {
	//       number = 12
	//     }
	//
	//     bar "label_b" {
	//       number = 34
	//     }
	//
	// Variables will be populated to be equivalent to the following JSON object:
	//
	//     {
	//       "foo": {
	//         "something": true
	//       },
	//       "bar": {
	//         "label_a": {
	//           "number": 12
	//         },
	//         "label_b": {
	//           "number": 34
	//         }
	//       }
	//     }
	ectx.Variables = make(map[string]cty.Value)

	// First, partition components by HCL block name.
	var componentsByBlockName = make(map[string][]*userComponent)
	for _, uc := range gc.components {
		blockName := uc.Name()[0]
		componentsByBlockName[blockName] = append(componentsByBlockName[blockName], uc)
	}

	// Then, convert each partition into a single value.
	for blockName, ucs := range componentsByBlockName {
		ectx.Variables[blockName] = gc.buildValue(ucs, 1)
	}

	return ectx
}

// buildValue recursively converts the set of user components into a single
// cty.Value. offset is used to determine which element in the
// userComponentName we're looking at.
func (gc *graphContext) buildValue(from []*userComponent, offset int) cty.Value {
	// We can't recurse anymore; return the node directly.
	if len(from) == 1 && offset >= len(from[0].Name()) {
		name := from[0].Name().String()

		cfg, ok := gc.configs[name]
		if !ok {
			cfg = cty.EmptyObjectVal
		}
		exports, ok := gc.exports[name]
		if !ok {
			exports = cty.EmptyObjectVal
		}

		return mergeComponentValues(cfg, exports)
	}

	attrs := make(map[string]cty.Value)

	// First, partition the components by their label.
	var componentsByLabel = make(map[string][]*userComponent)
	for _, uc := range from {
		blockName := uc.Name()[offset]
		componentsByLabel[blockName] = append(componentsByLabel[blockName], uc)
	}

	// Then, convert each partition into a single value.
	for label, ucs := range componentsByLabel {
		attrs[label] = gc.buildValue(ucs, offset+1)
	}

	return cty.ObjectVal(attrs)
}

// mergeComponentValues merges a component's config and exports. mergeState
// panics if a key exits in both inputs and store or if neither argument is an
// object.
func mergeComponentValues(config, exports cty.Value) cty.Value {
	if !config.Type().IsObjectType() {
		panic("component config must be object type")
	}
	if !exports.Type().IsObjectType() {
		panic("component exports must be object type")
	}

	var (
		inputMap = config.AsValueMap()
		stateMap = exports.AsValueMap()
	)

	mergedMap := make(map[string]cty.Value, len(inputMap)+len(stateMap))
	for key, value := range inputMap {
		mergedMap[key] = value
	}
	for key, value := range stateMap {
		if _, exist := mergedMap[key]; exist {
			panic(fmt.Sprintf("component exports overrides config key %s", key))
		}
		mergedMap[key] = value
	}

	return cty.ObjectVal(mergedMap)
}
