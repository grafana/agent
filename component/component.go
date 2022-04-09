package component

import "context"

// Component is a flow component. Flow components run in the background and
// optionally emit state.
type Component[Config any] interface {
	// Run starts the component, blocking until the provided context is canceled
	// or an error occurs.
	//
	// Components which have an output state may invoke onStateChange any time to
	// queue re-processing the state of the component.
	Run(ctx context.Context, onStateChange func()) error

	// CurrentState returns the current state of the component. Components may
	// return nil if there is no output state.
	//
	// CurrentState may be called at any time and must be safe to call
	// concurrently while the component updates its internal state.
	CurrentState() interface{}

	// Config returns the loaded Config of the component.
	Config() Config
}

// UpdatableComponent is an optional extention interface that Components may
// implement. Components that do not implement UpdatableComponent are updated
// by being shut down and replaced with a new instance constructed with the
// newest config.
type UpdatableComponent[Config any] interface {
	Component[Config]

	// Update provides a new Config to the component. An error may be returned if
	// the provided config object is invalid.
	Update(c Config) error
}
