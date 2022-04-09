package rawcomponent

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/zclconf/go-cty/cty"
	"golang.org/x/net/context"
)

type Component interface {
	// Run runs a component until ctx is canceled or an error occurs.
	Run(ctx context.Context, onStateChange func()) error

	// Update updates a component.
	Update(ectx *hcl.EvalContext, block *hcl.Block) error

	// CurrentState returns the currente state of a component.
	CurrentState() cty.Value
}
