package flow

import (
	"fmt"
	"reflect"
	"sync"

	"github.com/grafana/agent/component"
	"github.com/hashicorp/hcl/v2"
	"github.com/rfratto/gohcl"
	"github.com/zclconf/go-cty/cty"
)

// componentNode is a lazily-constructed component.
type componentNode struct {
	ref reference

	mut   sync.RWMutex
	block *hcl.Block
	raw   component.Component
	cfg   any
}

// newComponentNode constructs a componentNode from a block.
func newComponentNode(block *hcl.Block) *componentNode {
	return &componentNode{
		ref:   referenceForBlock(block),
		block: block,
	}
}

func referenceForBlock(block *hcl.Block) reference {
	ref := make(reference, 0, 1+len(block.Labels))
	ref = append(ref, block.Type)
	ref = append(ref, block.Labels...)
	return ref
}

// Reference returns the component reference.
func (cn *componentNode) Reference() reference {
	return cn.ref
}

// Name implements dag.Node, returning the reference as a string.
func (cn *componentNode) Name() string {
	return cn.ref.String()
}

// Config returns the current input config of the component.
func (cn *componentNode) Config() any {
	cn.mut.RLock()
	defer cn.mut.RUnlock()
	return cn.cfg
}

// ConfigValue returns the current config of the component as a cty.Value.
func (cn *componentNode) ConfigValue() cty.Value {
	cn.mut.RLock()
	defer cn.mut.RUnlock()

	if cn.cfg == nil {
		return cty.EmptyObjectVal
	}

	ty, err := gohcl.ImpliedType(cn.cfg)
	if err != nil {
		panic(err)
	}
	cv, err := gohcl.ToCtyValue(cn.cfg, ty)
	if err != nil {
		panic(err)
	}
	return cv
}

// State returns the current output state of the component.
func (cn *componentNode) State() interface{} {
	cn.mut.RLock()
	defer cn.mut.RUnlock()

	sc, _ := cn.raw.(component.StatefulComponent)
	if sc == nil {
		return nil
	}
	return sc.CurrentState()
}

// StateValue returns the current output state of the component as a cty.Value.
func (cn *componentNode) StateValue() cty.Value {
	cn.mut.RLock()
	defer cn.mut.RUnlock()

	sc, _ := cn.raw.(component.StatefulComponent)
	if sc == nil {
		return cty.EmptyObjectVal
	}

	val := sc.CurrentState()
	if val == nil {
		return cty.EmptyObjectVal
	}

	ty, err := gohcl.ImpliedType(val)
	if err != nil {
		panic(err)
	}
	cv, err := gohcl.ToCtyValue(val, ty)
	if err != nil {
		panic(err)
	}
	return cv
}

// Get retrieves the underlying component, if one is set.
func (cn *componentNode) Get() component.Component {
	cn.mut.RLock()
	defer cn.mut.RUnlock()
	return cn.raw
}

// Build performs the initial construction of the component. Fails if the
// component was already built.
func (cn *componentNode) Build(opts component.Options, ectx *hcl.EvalContext) error {
	cn.mut.Lock()
	defer cn.mut.Unlock()

	if cn.raw != nil {
		return fmt.Errorf("component %s already built", cn.Name())
	}

	componentID := cn.ref[:len(cn.ref)-1]
	reg, ok := component.Get(componentID.String())
	if !ok {
		// This should never happen: it's not possible for us to get this far if
		// the component doesn't exist since the block has already been validated
		// with the schema of registered components.
		panic("Could not find registration for component " + cn.Name())
	}

	cfg := reg.CloneConfig()
	diags := gohcl.DecodeBody(cn.block.Body, ectx, cfg)
	if diags.HasErrors() {
		return diags
	}

	cfgCopy := reflect.ValueOf(cfg).Elem().Interface()
	raw, err := reg.BuildComponent(opts, cfgCopy)
	if err != nil {
		return err
	}

	cn.raw = raw
	cn.cfg = cfgCopy
	return nil
}

// Update re-evaluates the component.
func (cn *componentNode) Update(ectx *hcl.EvalContext) error {
	cn.mut.Lock()
	defer cn.mut.Unlock()

	if cn.raw == nil {
		return fmt.Errorf("component not built")
	}

	componentID := cn.ref[:len(cn.ref)-1]
	reg, ok := component.Get(componentID.String())
	if !ok {
		// This should never happen: it's not possible for us to get this far if
		// the component doesn't exist since the block has already been validated
		// with the schema of registered components.
		panic("Could not find registration for component " + cn.Name())
	}
	cfg := reg.CloneConfig()
	diags := gohcl.DecodeBody(cn.block.Body, ectx, &cfg)
	if diags.HasErrors() {
		return diags
	}

	cfgCopy := reflect.ValueOf(cfg).Elem().Interface()
	if err := cn.raw.Update(cfgCopy); err != nil {
		return err
	}

	cn.cfg = cfgCopy
	return nil
}
