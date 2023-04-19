package module

import (
	"github.com/grafana/agent/component"
)

type Module struct {
	o *Options
}

func NewModule(o *Options) *Module {

}

// NewModuleController creates a new, unstarted ModuleController.
func (m *Module) NewModuleController() (_ component.ModuleController, _ error) {

}
