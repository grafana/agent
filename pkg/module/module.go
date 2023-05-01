package module

import (
	"github.com/grafana/agent/component"
)

type Module struct {
	o *Options
}

func NewModule(o *Options) *Module {
	return &Module{
		o: o,
	}
}

// NewModuleController creates a new, unstarted ModuleDelegate.
func (m *Module) NewModuleController(id string) component.ModuleDelegate {
	return NewController(id, m.o)
}
