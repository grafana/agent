package flow

import (
	"fmt"
	"sync"

	"github.com/hashicorp/hcl/v2"
	"github.com/rfratto/gohcl"
	"github.com/zclconf/go-cty/cty"
)

// TODO(rfratto): need the ability to remove nodes

// execContext tracks tracks components and caches gocty values. execContext
// can be modified concurrently.
type execContext struct {
	// execContext caches cty.Values to minimize the number of allocs when
	// updating many nodes at once.

	mut        sync.RWMutex
	components map[string]*componentNode
	configs    map[string]cty.Value
	states     map[string]cty.Value
}

func newNameTable() *execContext {
	return &execContext{
		components: make(map[string]*componentNode),
		configs:    make(map[string]cty.Value),
		states:     make(map[string]cty.Value),
	}
}

// AddNode inserts the provided component into nt.
func (nt *execContext) AddNode(cn *componentNode) {
	nt.mut.Lock()
	defer nt.mut.Unlock()
	nt.components[cn.Name()] = cn
}

// CacheConfig caches the input config for a node.
func (nt *execContext) CacheConfig(cn *componentNode) {
	var val cty.Value

	if cfg := cn.Config(); cfg == nil {
		val = cty.EmptyObjectVal
	} else {
		ty, err := gohcl.ImpliedType(cfg)
		if err != nil {
			panic(err)
		}
		cv, err := gohcl.ToCtyValue(cfg, ty)
		if err != nil {
			panic(err)
		}
		val = cv
	}

	nt.mut.Lock()
	defer nt.mut.Unlock()
	nt.configs[cn.Name()] = val
}

// CacheState caches the state for a node.
func (nt *execContext) CacheState(cn *componentNode) {
	var val cty.Value

	if state := cn.State(); state == nil {
		val = cty.EmptyObjectVal
	} else {
		ty, err := gohcl.ImpliedType(state)
		if err != nil {
			panic(err)
		}
		cv, err := gohcl.ToCtyValue(state, ty)
		if err != nil {
			panic(err)
		}
		val = cv
	}

	nt.mut.Lock()
	defer nt.mut.Unlock()
	nt.states[cn.Name()] = val
}

// LookupTraversal tries to find a node from a traversal. The traversal will be
// incrementally searched until the node is found.
func (nt *execContext) LookupTraversal(t hcl.Traversal) (*componentNode, hcl.Diagnostics) {
	nt.mut.RLock()
	defer nt.mut.RUnlock()

	var (
		diags hcl.Diagnostics

		split = t.SimpleSplit()
		ref   = reference{split.RootName()}
		rem   = split.Rel
	)

Lookup:
	for {
		if cn, found := nt.components[ref.String()]; found {
			return cn, nil
		}

		if len(rem) == 0 {
			// Stop: There's no more elements in the traversal; stop.
			break
		}

		// Find the next name in the traversal and append it to our reference.
		switch v := rem[0].(type) {
		case hcl.TraverseAttr:
			ref = append(ref, v.Name)
			// Shift rem forward one
			rem = rem[1:]
		default:
			break Lookup
		}
	}

	diags = diags.Append(&hcl.Diagnostic{
		Severity: hcl.DiagError,
		Summary:  "Invalid reference",
		Detail:   fmt.Sprintf("could not find component %s", ref),
		Subject:  split.Abs.SourceRange().Ptr(),
	})
	return nil, diags
}

// BuildEvalContext builds an hcl.EvalContext based on the nametable.
func (nt *execContext) BuildEvalContext() (*hcl.EvalContext, error) {
	ectx := &hcl.EvalContext{
		Variables: make(map[string]cty.Value),
	}

	// Split by top level-key.
	var nodesByKey = make(map[string][]*componentNode)
	for _, n := range nt.components {
		key := n.ref[0]
		nodesByKey[key] = append(nodesByKey[key], n)
	}

	// Now convert those nodes into a single value.
	for key, nodes := range nodesByKey {
		val, err := nt.buildValue(nodes, 1)
		if err != nil {
			return nil, err
		}
		ectx.Variables[key] = val
	}

	return ectx, nil
}

func (nt *execContext) buildValue(from []*componentNode, offset int) (cty.Value, error) {
	// We can't recurse anymore; return the node directly.
	if len(from) == 1 && offset >= len(from[0].ref) {
		cn := from[0]
		name := cn.Name()

		cfg, ok := nt.configs[name]
		if !ok {
			cfg = cty.EmptyObjectVal
		}
		state, ok := nt.states[name]
		if !ok {
			state = cty.EmptyObjectVal
		}

		return mergeState(cfg, state), nil
	}

	attrs := make(map[string]cty.Value)

	// We move more nodes to parition by.
	var nodesByKey = make(map[string][]*componentNode)
	for _, n := range from {
		key := n.ref[offset]
		nodesByKey[key] = append(nodesByKey[key], n)
	}

	// Now convert those nodes into a single value.
	for key, nodes := range nodesByKey {
		val, err := nt.buildValue(nodes, offset+1)
		if err != nil {
			return cty.DynamicVal, err
		}
		attrs[key] = val
	}

	return cty.ObjectVal(attrs), nil
}

// mergeState merges two the inputs of a component with its current state.
// mergeState panics if a key exits in both inputs and store or if neither
// argument is an object.
func mergeState(inputs, state cty.Value) cty.Value {
	if !inputs.Type().IsObjectType() {
		panic("component input must be object type")
	}
	if !state.Type().IsObjectType() {
		panic("component state must be object type")
	}

	var (
		inputMap = inputs.AsValueMap()
		stateMap = state.AsValueMap()
	)

	mergedMap := make(map[string]cty.Value, len(inputMap)+len(stateMap))
	for key, value := range inputMap {
		mergedMap[key] = value
	}
	for key, value := range stateMap {
		if _, exist := mergedMap[key]; exist {
			panic(fmt.Sprintf("component state overriding key %s", key))
		}
		mergedMap[key] = value
	}

	return cty.ObjectVal(mergedMap)
}
