package flow

import (
	"context"
	"fmt"
	"path"
	"sync"

	"github.com/grafana/agent/component"
	"github.com/grafana/agent/pkg/cluster"
	"github.com/grafana/agent/pkg/flow/internal/controller"
	"github.com/grafana/agent/pkg/flow/logging"
	"github.com/grafana/agent/pkg/flow/tracing"
	"github.com/grafana/agent/pkg/river/scanner"
	"github.com/grafana/agent/pkg/river/token"
	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/exp/maps"
)

type moduleController struct {
	mut     sync.RWMutex
	o       *moduleControllerOptions
	modules map[string]struct{}
}

var (
	_ component.ModuleController = (*moduleController)(nil)
)

// newModuleController is the entrypoint into creating module instances.
func newModuleController(o *moduleControllerOptions) controller.ModuleController {
	return &moduleController{
		o:       o,
		modules: map[string]struct{}{},
	}
}

// NewModule creates a new, unstarted Module.
func (m *moduleController) NewModule(id string, export component.ExportFunc) (component.Module, error) {
	if id != "" && !isValidIdentifier(id) {
		return nil, fmt.Errorf("module ID %q is not a valid River identifier", id)
	}

	m.mut.Lock()
	defer m.mut.Unlock()
	fullPath := m.o.ID
	if id != "" {
		fullPath = path.Join(fullPath, id)
	}
	if _, found := m.modules[fullPath]; found {
		return nil, fmt.Errorf("id %s already exists", id)
	}

	mod := newModule(&moduleOptions{
		ID:                      fullPath,
		export:                  export,
		moduleControllerOptions: m.o,
		parent:                  m,
	})

	if err := m.o.ModuleRegistry.Register(fullPath, mod); err != nil {
		return nil, err
	}

	m.modules[fullPath] = struct{}{}
	return mod, nil
}

func isValidIdentifier(in string) bool {
	s := scanner.New(nil, []byte(in), nil, 0)
	_, tok, lit := s.Scan()
	return tok == token.IDENT && lit == in
}

func (m *moduleController) removeID(id string) {
	m.mut.Lock()
	defer m.mut.Unlock()

	delete(m.modules, id)
	m.o.ModuleRegistry.Unregister(id)
}

// ModuleIDs implements [controller.ModuleController].
func (m *moduleController) ModuleIDs() []string {
	m.mut.RLock()
	defer m.mut.RUnlock()
	return maps.Keys(m.modules)
}

type module struct {
	f *Flow
	o *moduleOptions
}

type moduleOptions struct {
	ID     string
	export component.ExportFunc
	parent *moduleController
	*moduleControllerOptions
}

var (
	_ component.Module = (*module)(nil)
)

// newModule creates a module instance for a specific component.
func newModule(o *moduleOptions) *module {
	return &module{
		o: o,
		f: newController(controllerOptions{
			IsModule:          true,
			ModuleRegistry:    o.ModuleRegistry,
			ComponentRegistry: o.ComponentRegistry,
			Options: Options{
				ControllerID:   o.ID,
				Tracer:         o.Tracer,
				Clusterer:      o.Clusterer,
				Reg:            o.Reg,
				Logger:         o.Logger,
				DataPath:       o.DataPath,
				HTTPPathPrefix: o.HTTPPath,
				HTTPListenAddr: o.HTTPListenAddr,
				OnExportsChange: func(exports map[string]any) {
					if o.export != nil {
						o.export(exports)
					}
				},
				DialFunc: o.DialFunc,
				Services: o.ServiceMap.List(),
			},
		}),
	}
}

// LoadConfig parses River config and loads it.
func (c *module) LoadConfig(config []byte, args map[string]any) error {
	ff, err := ReadFile(c.o.ID, config)
	if err != nil {
		return err
	}
	return c.f.LoadFile(ff, args)
}

// Run starts the Module. No components within the Module
// will be run until Run is called.
//
// Run blocks until the provided context is canceled.
func (c *module) Run(ctx context.Context) {
	defer c.o.parent.removeID(c.o.ID)
	c.f.Run(ctx)
}

// moduleControllerOptions holds static options for module controller.
type moduleControllerOptions struct {
	// Logger to use for controller logs and components. A no-op logger will be
	// created if this is nil.
	Logger *logging.Logger

	// Tracer for components to use. A no-op tracer will be created if this is
	// nil.
	Tracer *tracing.Tracer

	// Clusterer for implementing distributed behavior among components running
	// on different nodes.
	Clusterer *cluster.Clusterer

	// Reg is the prometheus register to use
	Reg prometheus.Registerer

	// A path to a directory with this component may use for storage. The path is
	// guaranteed to be unique across all running components.
	//
	// The directory may not exist when the component is created; components
	// should create the directory if needed.
	DataPath string

	// HTTPListenAddr is the address the server is configured to listen on.
	HTTPListenAddr string

	// HTTPPath is the base path that requests need in order to route to this
	// component. Requests received by a component handler will have this already
	// trimmed off.
	HTTPPath string

	// DialFunc is a function for components to use to properly communicate to
	// HTTPListenAddr. If set, components which send HTTP requests to
	// HTTPListenAddr must use this function to establish connections.
	controller.DialFunc

	// ID is the attached components full ID.
	ID string

	// ComponentRegistry is where controllers can look up components.
	ComponentRegistry controller.ComponentRegistry

	// ModuleRegistry is a shared registry of running modules from the same root
	// controller.
	ModuleRegistry *moduleRegistry

	// ServiceMap is a map of services which can be used in the module
	// controller.
	ServiceMap controller.ServiceMap
}
