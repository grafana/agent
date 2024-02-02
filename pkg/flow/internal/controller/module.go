package controller

import (
	"context"

	"github.com/grafana/agent/component"
	"github.com/grafana/river/ast"
)

// ModuleController is a lower-level interface for module controllers which
// allows probing for the list of managed modules.
type ModuleController interface {
	component.ModuleController

	// ModuleIDs returns the list of managed modules in unspecified order.
	ModuleIDs() []string

	NewModuleV2(id string, export component.ExportFunc) (Module, error)
}

type LoaderConfigOptions struct {
	CustomComponentRegistry *CustomComponentRegistry
}

// Module is a controller for running components within a Module.
type Module interface {
	// LoadBody loads a River AST body into the Module. LoadBody can be called
	// multiple times, and called prior to [Module.Run].
	LoadBody(body ast.Body, args map[string]any, options LoaderConfigOptions) error

	// Run starts the Module. No components within the Module
	// will be run until Run is called.
	//
	// Run blocks until the provided context is canceled. The ID of a module as defined in
	// ModuleController.NewModule will not be released until Run returns.
	Run(context.Context) error
}
