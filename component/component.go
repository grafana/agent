// Package component describes the interfaces which Flow components implement.
//
// A Flow component is a distinct piece of business logic that accepts inputs
// (Arguments) for its configuration and can optionally export a set of outputs
// (Exports).
//
// Arguments and Exports do not need to be static for the lifetime of a
// component. A component will be given a new Config if the runtime
// configuration changes. A component may also update its Exports throughout
// its lifetime, such as a component which outputs the current day of the week.
//
// Components are built by users with HCL, where they can use HCL expressions
// to refer to any input or exported field from other components. This allows
// users to connect components together to declaratively form a pipeline.
//
// Defining Arguments and Exports structs
//
// Arguments and Exports implemented by new components must be able to be
// encoded to and from HCL. "hcl" struct field tags are used for encoding;
// refer to the package documentation of github.com/rfratto/gohcl for a
// description of how to write these tags.
//
// The set of HCL element names of a given component's Arguments and Exports
// types must not overlap. Additionally, the following HCL element names are
// reserved for use by the Flow controller:
//
//     * for_each
//     * enabled
//     * health
//     * debug
//
// Default values for Arguments may be provided by implementing gohcl.Decoder.
//
// Mapping HCL strings to custom types
//
// Custom encoding and decoding of fields is available by implementing
// encoding.TextMarshaler and encoding.TextUnmarshaler. Types implementing
// these interfaces will be represented as strings in the HCL.
//
// Exposing advanced Go structs to HCL
//
// Go structs which contain interfaces, channels, or pointers can be encoded to
// and from HCL by calling RegisterGoStruct. This allows components to pass
// around arbitrary values for binding complex logic, such as a data stream.
//
// Component registration
//
// Components are registered globally by calling Register. These components are
// then made available by including them in the import path. The "all" child
// package imports all known component packages and should be updated when
// creating a new one.
package component

import "context"

// The Arguments contains the input fields for a specific component, which is
// unmarshaled from HCL.
//
// Refer to the package documentation for details around how to build proper
// Arguments implementations.
type Arguments interface{}

// Exports contains the current set of outputs for a specific component, which
// is then marshaled to HCL.
//
// Refer to the package documentation for details around how to build proper
// Exports implementations.
type Exports interface{}

// Component is the base interface for a Flow component. Components may
// implement extension interfaces (named <Extension>Component) to implement
// extra known behavior.
type Component interface {
	// Run starts the component, blocking until ctx is canceled or the component
	// suffers a fatal error. Run is guaranteed to be called exactly once per
	// Component.
	//
	// Implementations of Component should perform any necessary cleanup before
	// returning from Run.
	Run(ctx context.Context) error

	// Update provides a new Config to the component. The type of newConfig will
	// always match the struct type which the component registers.
	//
	// Update will be called concurrently with Run. The component must be able to
	// gracefully handle updating its config will still running.
	//
	// An error may be returned if the provided config is invalid.
	Update(args Arguments) error
}

// DebugComponent is an extension interface for components which can report
// debugging information upon request.
type DebugComponent interface {
	Component

	// DebugInfo returns the current debug information of the component. May
	// return nil if there is no debug info to currently report. The result of
	// DebugInfo must be encodable to HCL like Arguments and Exports.
	//
	// Values from DebugInfo are not exposed to other components for use in
	// expressions.
	//
	// DebugInfo must be safe for calling concurrently.
	DebugInfo() interface{}
}
