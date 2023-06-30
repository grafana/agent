package controller

import "github.com/grafana/agent/component"

// ModuleController is a lower-level interface for module controllers which
// allows probing for the list of managed modules.
type ModuleController interface {
	component.ModuleController

	// ModuleIDs returns the list of managed modules in unspecified order.
	ModuleIDs() []string
}
