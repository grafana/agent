package flow

import (
	"github.com/grafana/agent/component"
	"github.com/hashicorp/hcl/v2"
	"github.com/rfratto/gohcl"
	"github.com/zclconf/go-cty/cty"
)

// componentNode is a lazily-constructed component.
type componentNode struct {
	ref   reference
	block *hcl.Block

	raw component.HCL
}

// newComponentNode constructs a componentNode from a block.
func newComponentNode(block *hcl.Block) *componentNode {
	ref := make(reference, 0, 1+len(block.Labels))
	ref = append(ref, block.Type)
	ref = append(ref, block.Labels...)

	return &componentNode{
		ref:   ref,
		block: block,
	}
}

func (cn *componentNode) Reference() reference {
	return cn.ref
}

func (cn *componentNode) Name() string {
	return cn.ref.String()
}

func (cn *componentNode) Config() cty.Value {
	val := cn.raw.Config()
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

func (cn *componentNode) CurrentState() cty.Value {
	val := cn.raw.CurrentState()
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

func (cn *componentNode) Set(rc component.HCL) {
	cn.raw = rc
}
