package flow

import (
	"github.com/grafana/agent/component"
)

type moduleController struct {
	o *moduleControllerOptions
}

var (
	_ component.ModuleController = (*moduleController)(nil)
)

// newModuleController is the entrypoint into creating module instances.
/*func newModuleController(o *moduleControllerOptions) component.ModuleController {
	return &module{
		o: o,
	}
}*/

// NewModule creates a new, unstarted Module.
func (m *moduleController) NewModule(id string, export component.ExportFunc) component.Module {
	return newModule(&moduleOptions{
		ID:                      id,
		export:                  export,
		moduleControllerOptions: m.o,
	})
}
