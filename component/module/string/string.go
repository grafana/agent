// Package string defines the module.string component.
package string

import (
	"context"
	"net/http"
	"path"
	"sync"

	"github.com/go-kit/log"
	"github.com/gorilla/mux"
	"github.com/grafana/agent/component"
	"github.com/grafana/agent/pkg/flow"
	"github.com/grafana/agent/pkg/flow/logging"
	"github.com/grafana/agent/pkg/flow/rivertypes"
	"github.com/grafana/agent/pkg/flow/tracing"
	"github.com/grafana/agent/web/api"
	"github.com/prometheus/client_golang/prometheus"
)

func init() {
	component.Register(component.Registration{
		Name:    "module.string",
		Args:    Arguments{},
		Exports: Exports{},

		Build: func(opts component.Options, args component.Arguments) (component.Component, error) {
			return New(opts, args.(Arguments))
		},
	})
}

// Arguments holds values which are used to configure the module.string
// component.
type Arguments struct {
	// Content to load for the module.
	Content rivertypes.OptionalSecret `river:"content,attr"`

	// Arguments to pass into the module.
	Arguments map[string]any `river:"arguments,attr,optional"`
}

// Exports holds values which are exported from the run module.
type Exports struct {
	// Exports exported from the running module.
	Exports map[string]any `river:"exports,attr"`
}

// Component implements the module.string component.
type Component struct {
	opts component.Options
	log  log.Logger
	ctrl *flow.Flow

	exportsMut sync.Mutex
	exports    map[string]any
}

var (
	_ component.Component     = (*Component)(nil)
	_ component.HTTPComponent = (*Component)(nil)
)

// New creates a new module.string component.
func New(o component.Options, args Arguments) (*Component, error) {
	// TODO(rfratto): replace these with a tracer/registry which properly
	// propagates data back to the parent.
	flowTracer, _ := tracing.New(tracing.DefaultOptions)
	flowRegistry := prometheus.NewRegistry()

	c := &Component{
		opts:    o,
		log:     o.Logger,
		exports: make(map[string]any),
	}
	c.ctrl = flow.New(flow.Options{
		ControllerID: o.ID,
		LogSink:      logging.LoggerSink(o.Logger),
		Tracer:       flowTracer,
		Reg:          flowRegistry,

		DataPath:       o.DataPath,
		HTTPPathPrefix: o.HTTPPath,
		HTTPListenAddr: o.HTTPListenAddr,

		OnExportsChange: func(exports map[string]any) {
			c.exportsMut.Lock()
			defer c.exportsMut.Unlock()

			// Update our primary export map with all the values.
			for k, v := range exports {
				c.exports[k] = v
			}
			// This is to prevent any access to our primary export map. So we create a copy for other usages.
			exportable := make(map[string]any)
			for k, v := range c.exports {
				exportable[k] = v
			}
			o.OnStateChange(Exports{Exports: exportable})
		},
	})

	if err := c.Update(args); err != nil {
		return nil, err
	}
	return c, nil
}

// Run implements component.Component.
func (c *Component) Run(ctx context.Context) error {
	c.ctrl.Run(ctx)
	return nil
}

// Update implements component.Component.
func (c *Component) Update(args component.Arguments) error {
	newArgs := args.(Arguments)

	f, err := flow.ReadFile(c.opts.ID, []byte(newArgs.Content.Value))
	if err != nil {
		return err
	}
	// Unsure of our exports so we should refresh those.
	c.exportsMut.Lock()
	c.exports = make(map[string]any)
	c.exportsMut.Unlock()

	return c.ctrl.LoadFile(f, newArgs.Arguments)
}

// Handler implements component.HTTPComponent.
func (c *Component) Handler() http.Handler {
	r := mux.NewRouter()

	fa := api.NewFlowAPI(c.ctrl, r)
	fa.RegisterRoutes("/", r)

	r.PathPrefix("/{id}/").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Re-add the full path to ensure that nested controllers propagate
		// requests properly.
		r.URL.Path = path.Join(c.opts.HTTPPath, r.URL.Path)

		c.ctrl.ComponentHandler().ServeHTTP(w, r)
	})

	return r
}
