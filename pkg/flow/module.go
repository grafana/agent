package flow

import (
	"github.com/grafana/agent/component"
)

type module struct {
	o *ModuleOptions
}

// newModuleSystem is the entrypoint into creating module delegates.
func newModuleSystem(o *ModuleOptions) component.ModuleSystem {
	return &module{
		o: o,
	}
}

// NewModuleDelegate creates a new, unstarted ModuleDelegate.
func (m *module) NewModuleDelegate(id string) component.ModuleDelegate {
	return newDelegate(id, m.o)
}
