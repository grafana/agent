package component

import (
	"context"
	"fmt"
	"net/http"
)

// Config is the configuration for a specific component.
type Config interface{}

// Component is a Flow component. Components run in a dedicated Goroutine.
type Component interface {
	// Run should run the component, blocking until ctx is canceled or the
	// component fails. Run is guaranteed to only be called exactly per
	// component, and is expected to clean up resources on defer.
	Run(ctx context.Context) error

	// Update provides a new config to the component. The type of config is
	// guaranteed to be the type of config registered with the component.
	//
	// An error should be returned if the provided config object is invalid.
	Update(newConfig Config) error
}

// StatefulComponent is an optional extension interface that Components which
// expose State to other components may implement.
type StatefulComponent interface {
	// CurrentState returns the current state of the component.
	//
	// CurrentState may be called at any time and must be safe to call
	// concurrently while the component is running.
	CurrentState() any
}

// HTTPComponent is an optional extension interface that Components which wish
// to register HTTP endpoints may implement.
type HTTPComponent interface {
	Component

	// ComponentHandler returns an http.Handler for the current component.
	// ComponentHandler may return nil to avoid registering any handlers.
	// ComponentHandler will only be invoked once per component.
	//
	// Each Component has a unique HTTP path prefix where its handler can be
	// reached. This prefix is trimmed when invoking the http.Handler. Use
	// HTTPPrefix to determine what that prefix is.
	ComponentHandler() (http.Handler, error)
}

// HTTPPrefix returns the URL path prefix assigned to a specific componentID.
// The path returned by HTTPPrefix ends in a trailing slash.
func HTTPPrefix(componentID string) string {
	return fmt.Sprintf("/component/%s/", componentID)
}
