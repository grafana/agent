package flow

import (
	"fmt"
	"sync"

	"github.com/grafana/agent/component"
)

type moduleRegistry struct {
	mut     sync.RWMutex
	modules map[string]*module
}

func newModuleRegistry() *moduleRegistry {
	return &moduleRegistry{
		modules: make(map[string]*module),
	}
}

// Get retrives a module by ID.
func (reg *moduleRegistry) Get(id string) (*module, bool) {
	reg.mut.RLock()
	defer reg.mut.RUnlock()

	mod, ok := reg.modules[id]
	return mod, ok
}

func (reg *moduleRegistry) GetAllComponents() []*component.Info {
	reg.mut.Lock()
	defer reg.mut.Unlock()

	components := make([]*component.Info, 0)
	for _, v := range reg.modules {
		subComponents, err := v.f.ListComponents("", component.InfoOptions{
			GetHealth:    true,
			GetArguments: true,
			GetExports:   true,
			GetDebugInfo: true,
		})
		if err != nil {
			continue
		}
		components = append(components, subComponents...)
	}
	return components
}

// Register registers a module by ID. It returns an error if that module is
// already registered.
func (reg *moduleRegistry) Register(id string, mod *module) error {
	reg.mut.Lock()
	defer reg.mut.Unlock()

	if _, exist := reg.modules[id]; exist {
		return fmt.Errorf("module %q already exists", id)
	}

	reg.modules[id] = mod
	return nil
}

// Unregister unregisters a module by ID. It is a no-op if the provided ID
// isn't registered.
func (reg *moduleRegistry) Unregister(id string) {
	reg.mut.Lock()
	defer reg.mut.Unlock()

	delete(reg.modules, id)
}
