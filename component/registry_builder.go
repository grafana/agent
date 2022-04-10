package component

import (
	"context"
	"fmt"
	"sync"

	"github.com/go-kit/log"
	"github.com/hashicorp/hcl/v2"
	"github.com/rfratto/gohcl"
	"github.com/zclconf/go-cty/cty"
)

type BuildContext struct {
	Log         log.Logger
	EvalContext *hcl.EvalContext
}

// BuildHCL builds the named raw component from HCL.
func BuildHCL(name string, bctx *BuildContext, b *hcl.Block) (HCL, error) {
	ent, exists := registered[name]
	if !exists {
		return nil, fmt.Errorf("component %q does not exist", name)
	}
	return ent.builder.BuildComponent(bctx, b)
}

// rawBuilder implements the builder interface and is used to abstract away
// generics.
type rawBuilder[Config any] struct {
	r Registration[Config]
}

func newRawBuilder[Config any](r Registration[Config]) builder {
	return &rawBuilder[Config]{r: r}
}

func (rb *rawBuilder[Config]) BuildComponent(bctx *BuildContext, b *hcl.Block) (HCL, error) {
	var cfg Config
	diags := gohcl.DecodeBody(b.Body, bctx.EvalContext, &cfg)
	if diags.HasErrors() {
		return nil, diags
	}

	c, err := rb.r.BuildComponent(bctx.Log, cfg)
	if err != nil {
		return nil, err
	}

	return newHCLAdapter(bctx.Log, rb.r, c), nil
}

// HCL is an HCL component; registered Components can be converted into HCL
// components by calling BuildHCL.
type HCL interface {
	// Run runs a component until ctx is canceled or an error occurs.
	Run(ctx context.Context, onStateChange func()) error

	// Update updates a component.
	Update(ectx *hcl.EvalContext, block *hcl.Block) error

	// CurrentState returns the currente state of a component.
	CurrentState() cty.Value
}

// hclAdapter wraps a flow component into an HCL component.
type hclAdapter[Config any] struct {
	r Registration[Config]

	mut       sync.Mutex
	cur       Component[Config]
	cancelCur context.CancelFunc
	log       log.Logger

	running     sync.WaitGroup
	newRunnable chan struct{}
}

func newHCLAdapter[Config any](l log.Logger, r Registration[Config], init Component[Config]) *hclAdapter[Config] {
	return &hclAdapter[Config]{
		r:   r,
		cur: init,
		log: l,

		newRunnable: make(chan struct{}, 1),
	}
}

func (rc *hclAdapter[Config]) Run(ctx context.Context, onStateChange func()) error {
	errCh := make(chan error, 1)

	// Not all commponents support being dynamically updated, so we have to
	// conceptually handle updating something which is static. This is done by
	// interally having multiple components over time.
	for {
		// Drain the error channel from the previous run (which will always be
		// written to)
		select {
		case <-errCh:
		default:
		}

		curCtx, curCancel := context.WithCancel(ctx)

		rc.mut.Lock()
		rc.cancelCur = curCancel
		rc.mut.Unlock()

		rc.running.Add(1)
		go func() {
			defer rc.running.Done()

			rc.mut.Lock()
			cur := rc.cur
			rc.mut.Unlock()

			errCh <- cur.Run(curCtx, onStateChange)
		}()

		select {
		case <-ctx.Done():
			// Our parent context is done; Make sure our runnable exits and return.
			curCancel()
			rc.running.Wait()
			return nil
		case err := <-errCh:
			// The goroutine exited; return from our run.
			rc.running.Wait()
			return err
		case <-rc.newRunnable:
			// There's a new runnable component. Make sure the current one is stopped
			// and re-start the loop.
			curCancel()
			rc.running.Wait()
		}
	}
}

func (rc *hclAdapter[Config]) Update(ectx *hcl.EvalContext, block *hcl.Block) error {
	var cfg Config
	diags := gohcl.DecodeBody(block.Body, ectx, &cfg)
	if diags.HasErrors() {
		return diags
	}

	uc, ok := rc.cur.(UpdatableComponent[Config])
	if ok {
		return uc.Update(cfg)
	}

	// The component we're running has no concept of updating. That means we need
	// to:
	//
	// 1. Stop the current component, waiting for it to exit.
	// 2. Attempt to create a new component.
	// 3. Set the component as the next runnable.

	rc.mut.Lock()
	rc.cancelCur()
	rc.mut.Unlock()

	rc.running.Wait()

	c, err := rc.r.BuildComponent(rc.log, cfg)
	if err != nil {
		return err
	}

	rc.mut.Lock()
	rc.cur = c
	rc.mut.Unlock()

	select {
	case rc.newRunnable <- struct{}{}:
	default:
	}

	return nil
}

func (rc *hclAdapter[Config]) CurrentState() cty.Value {
	rc.mut.Lock()
	defer rc.mut.Unlock()

	val := rc.cur.CurrentState()
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
