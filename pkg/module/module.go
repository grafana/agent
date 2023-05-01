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

// NewModuleController creates a new, unstarted ModuleController.
func (m *Module) NewModuleController(id string) component.ModuleController {
	return NewController(id, m.o)

}
