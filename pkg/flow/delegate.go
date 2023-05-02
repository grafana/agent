package flow

import (
	"context"
	"net/http"
	"path"

	"github.com/gorilla/mux"
	"github.com/grafana/agent/component"
	"github.com/grafana/agent/pkg/traces"
	"github.com/grafana/agent/web/api"
)

type delegate struct {
	f        *Flow
	o        *ModuleOptions
	httppath string
	id       string
}

// newDelegate creates a module delegate for a specific component.
func newDelegate(id string, o *ModuleOptions) *delegate {
	return &delegate{
		o:  o,
		id: id,
	}
}

// LoadConfig parses River config and loads it.
func (c *delegate) LoadConfig(config []byte, o component.Options, args map[string]any, onExport component.Export) error {
	if c.f == nil {
		c.httppath = o.HTTPPath
		f := New(Options{
			ControllerID: c.id,
			Logger:       c.o.Logger,
			Tracer:       traces.WrapTracer(c.o.Tracer, c.id),
			Clusterer:    c.o.Clusterer,
			Reg:          c.o.Reg,
			// Use the component options here since they
			// are tied to this specific component.
			DataPath:       o.DataPath,
			HTTPPathPrefix: o.HTTPPath,
			HTTPListenAddr: o.HTTPListenAddr,
			OnExportsChange: func(exports map[string]any) {
				onExport(exports)
			},
		})
		c.f = f
	}

	ff, err := ReadFile(c.id, config)
	if err != nil {
		return err
	}
	return c.f.LoadFile(ff, args)
}

// Run starts the ModuleDelegate. No components within the ModuleDelegate
// will be run until Run is called.
//
// Run blocks until the provided context is canceled.
func (c *delegate) Run(ctx context.Context) {
	c.f.Run(ctx)
}

// ComponentHandler returns an HTTP handler which exposes endpoints of
// components managed by the underlying flow system.
func (c *delegate) ComponentHandler() (_ http.Handler) {
	r := mux.NewRouter()

	fa := api.NewFlowAPI(c.f)
	fa.RegisterRoutes("/", r)

	r.PathPrefix("/{id}/").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Re-add the full path to ensure that nested controllers propagate
		// requests properly.
		r.URL.Path = path.Join(c.httppath, r.URL.Path)

		c.f.ComponentHandler().ServeHTTP(w, r)
	})

	return r
}
