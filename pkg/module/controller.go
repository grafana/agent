package module

import (
	"context"
	"net/http"
	"path"

	"github.com/gorilla/mux"
	"github.com/grafana/agent/pkg/flow"
	"github.com/grafana/agent/web/api"
)

type controller struct {
	f      *flow.Flow
	o      *Options
	id     string
	parent *Module
}

func NewController(id string, o *Options, parent *Module) *controller {
	return &controller{
		o:      o,
		id:     id,
		parent: parent,
	}

}

// LoadConfig parses River config and loads it.
func (c *controller) LoadConfig(config []byte, args map[string]any, onExport func(exports map[string]any)) error {
	if c.f == nil {
		f := flow.New(flow.Options{
			ControllerID:   c.id,
			LogSink:        c.o.LogSink,
			Tracer:         c.o.Tracer,
			Clusterer:      c.o.Clusterer,
			DataPath:       c.o.DataPath,
			Reg:            c.o.Reg,
			HTTPPathPrefix: c.o.HTTPPathPrefix,
			HTTPListenAddr: c.o.HTTPListenAddr,
			OnExportsChange: func(exports map[string]any) {
				onExport(exports)
			},
			Controller: c.parent,
		})
		c.f = f
	}

	ff, err := flow.ReadFile(c.id, config)
	if err != nil {
		return err
	}
	return c.f.LoadFile(ff, args)
}

// Run starts the ModuleController. No components within the ModuleController
// will be run until Run is called.
//
// Run blocks until the provided context is canceled.
func (c *controller) Run(ctx context.Context) {
	c.f.Run(ctx)
}

// ComponentHandler returns an HTTP handler which exposes endpoints of
// components managed by the c *controllerController.
func (c *controller) ComponentHandler() (_ http.Handler) {
	r := mux.NewRouter()

	fa := api.NewFlowAPI(c.f, r)
	fa.RegisterRoutes("/", r)

	r.PathPrefix("/{id}/").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Re-add the full path to ensure that nested controllers propagate
		// requests properly.
		r.URL.Path = path.Join(c.o.HTTPPathPrefix, r.URL.Path)

		c.f.ComponentHandler().ServeHTTP(w, r)
	})

	return r

}
