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

// NewModuleDelegate creates a new, unstarted ModuleDelegate.
func (m *Module) NewModuleDelegate(id string) component.ModuleDelegate {
	return newDelegate(id, m.o)
}
