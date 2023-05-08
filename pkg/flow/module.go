package flow

import (
	"github.com/grafana/agent/component"
)

type module struct {
	o *moduleOptions
}

var (
	_ component.ModuleSystem = (*module)(nil)
)

// newModuleSystem is the entrypoint into creating module delegates.
/*func newModuleSystem(o *moduleOptions) component.ModuleSystem {
	return &module{
		o: o,
	}
}*/

// NewModuleDelegate creates a new, unstarted ModuleDelegate.
func (m *module) NewModuleDelegate(id string) component.ModuleDelegate {
	return newDelegate(id, m.o)
}
