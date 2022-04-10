package component

import (
	"context"
	"fmt"
	"sync"

	"github.com/go-kit/log"
	"github.com/hashicorp/hcl/v2"
	"github.com/rfratto/gohcl"
)

// BuildContext is a set of options provided when building a new component.
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

func (b *rawBuilder[Config]) BuildComponent(bctx *BuildContext, block *hcl.Block) (HCL, error) {
	var cfg Config
	diags := gohcl.DecodeBody(block.Body, bctx.EvalContext, &cfg)
	if diags.HasErrors() {
		return nil, diags
	}

	c, err := b.r.BuildComponent(bctx.Log, cfg)
	if err != nil {
		return nil, err
	}

	return newHCLAdapter(bctx.Log, b.r, c), nil
}

// HCL is an HCL component; registered Components can be converted into HCL
// components by calling BuildHCL.
type HCL interface {
	// Run runs a component until ctx is canceled or an error occurs.
	Run(ctx context.Context, onStateChange func()) error

	// Update updates a component.
	Update(ectx *hcl.EvalContext, block *hcl.Block) error

	// Config returns the current config of a component.
	Config() interface{}

	// CurrentState returns the current state of a component.
	CurrentState() interface{}
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

func (a *hclAdapter[Config]) Run(ctx context.Context, onStateChange func()) error {
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

		a.mut.Lock()
		a.cancelCur = curCancel
		a.mut.Unlock()

		a.running.Add(1)
		go func() {
			defer a.running.Done()

			a.mut.Lock()
			cur := a.cur
			a.mut.Unlock()

			errCh <- cur.Run(curCtx, onStateChange)
		}()

		select {
		case <-ctx.Done():
			// Our parent context is done; Make sure our runnable exits and return.
			curCancel()
			a.running.Wait()
			return nil
		case err := <-errCh:
			// The goroutine exited; return from our run.
			a.running.Wait()
			return err
		case <-a.newRunnable:
			// There's a new runnable component. Make sure the current one is stopped
			// and re-start the loop.
			curCancel()
			a.running.Wait()
		}
	}
}

func (a *hclAdapter[Config]) Update(ectx *hcl.EvalContext, block *hcl.Block) error {
	var cfg Config
	diags := gohcl.DecodeBody(block.Body, ectx, &cfg)
	if diags.HasErrors() {
		return diags
	}

	uc, ok := a.cur.(UpdatableComponent[Config])
	if ok {
		return uc.Update(cfg)
	}

	// The component we're running has no concept of updating. That means we need
	// to:
	//
	// 1. Stop the current component, waiting for it to exit.
	// 2. Attempt to create a new component.
	// 3. Set the component as the next runnable.

	a.mut.Lock()
	a.cancelCur()
	a.mut.Unlock()

	a.running.Wait()

	c, err := a.r.BuildComponent(a.log, cfg)
	if err != nil {
		return err
	}

	a.mut.Lock()
	a.cur = c
	a.mut.Unlock()

	select {
	case a.newRunnable <- struct{}{}:
	default:
	}

	return nil
}

func (a *hclAdapter[Config]) Config() interface{} {
	a.mut.Lock()
	defer a.mut.Unlock()
	return a.cur.Config()
}

func (a *hclAdapter[Config]) CurrentState() interface{} {
	a.mut.Lock()
	defer a.mut.Unlock()
	return a.cur.CurrentState()
}
