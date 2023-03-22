// Package string defines the module.file component.
package file

import (
	"context"
	"net/http"
	"sync"

	"github.com/go-kit/log"
	"github.com/gorilla/mux"
	"github.com/grafana/agent/component"
	"github.com/grafana/agent/component/local/file"
	"github.com/grafana/agent/pkg/flow"
	"github.com/grafana/agent/pkg/flow/logging"
	"github.com/grafana/agent/pkg/flow/rivertypes"
	"github.com/grafana/agent/pkg/flow/tracing"
	"github.com/grafana/agent/pkg/river"
	"github.com/grafana/agent/web/api"
	"github.com/prometheus/client_golang/prometheus"
)

func init() {
	component.Register(component.Registration{
		Name:    "module.file",
		Args:    Arguments{},
		Exports: Exports{},

		Build: func(opts component.Options, args component.Arguments) (component.Component, error) {
			return New(opts, args.(Arguments))
		},
	})
}

// Arguments holds values which are used to configure the module.file component.
type Arguments struct {
	LocalFileArguments file.Arguments `river:",squash"`

	// Arguments to pass into the module.
	Arguments map[string]any `river:"arguments,attr,optional"`
}

var _ river.Unmarshaler = (*Arguments)(nil)

// UnmarshalRiver implements river.Unmarshaler.
func (a *Arguments) UnmarshalRiver(f func(interface{}) error) error {
	a.LocalFileArguments = file.DefaultArguments

	type arguments Arguments
	err := f((*arguments)(a))
	if err != nil {
		return err
	}

	return nil
}

// Exports holds values which are exported from the run module.
type Exports struct {
	// Exports exported from the running module.
	Exports map[string]any `river:"exports,attr"`
}

// Component implements the module.file component.
type Component struct {
	opts component.Options
	log  log.Logger
	ctrl *flow.Flow

	mut     sync.RWMutex
	args    Arguments
	content rivertypes.OptionalSecret

	managedLocalFile *file.Component
}

var (
	_ component.Component     = (*Component)(nil)
	_ component.HTTPComponent = (*Component)(nil)
)

// New creates a new module.file component.
func New(o component.Options, args Arguments) (*Component, error) {
	// TODO(rfratto): replace these with a tracer/registry which properly
	// propagates data back to the parent.
	flowTracer, _ := tracing.New(tracing.DefaultOptions)
	flowRegistry := prometheus.NewRegistry()

	c := &Component{
		opts: o,
		log:  o.Logger,

		ctrl: flow.New(flow.Options{
			ControllerID: o.ID,
			LogSink:      logging.LoggerSink(o.Logger),
			Tracer:       flowTracer,
			Reg:          flowRegistry,

			DataPath:       o.DataPath,
			HTTPPathPrefix: o.HTTPPath,
			HTTPListenAddr: o.HTTPListenAddr,

			OnExportsChange: func(exports map[string]any) {
				o.OnStateChange(Exports{Exports: exports})
			},
		}),
	}

	localFile, err := c.NewManagedLocalFileComponent(o, args)
	if err != nil {
		return nil, err
	}

	c.managedLocalFile = localFile

	if err := c.Update(args); err != nil {
		return nil, err
	}
	return c, nil
}

// Run implements component.Component.
func (c *Component) Run(ctx context.Context) error {
	go c.managedLocalFile.Run(ctx)
	c.ctrl.Run(ctx)
	return nil
}

// Update implements component.Component.
func (c *Component) Update(args component.Arguments) error {
	newArgs := args.(Arguments)

	c.mut.Lock()
	c.args = newArgs
	c.mut.Unlock()

	c.managedLocalFile.Update(newArgs.LocalFileArguments)

	return c.LoadFlowContent()
}

// NewManagedLocalFileComponent creates the new local.file managed component.
func (c *Component) NewManagedLocalFileComponent(o component.Options, args Arguments) (*file.Component, error) {
	localFileOpts := o
	localFileOpts.OnStateChange = func(e component.Exports) {
		c.mut.Lock()
		c.content = e.(file.Exports).Content
		c.mut.Unlock()
		c.LoadFlowContent()
	}

	return file.New(localFileOpts, args.LocalFileArguments)
}

// LoadFlowContent loads the flow controller with the current component content.
func (c *Component) LoadFlowContent() error {
	c.mut.RLock()
	defer c.mut.RUnlock()
	f, err := flow.ReadFile(c.opts.ID, []byte(c.content.Value))
	if err != nil {
		return err
	}

	return c.ctrl.LoadFile(f, c.args.Arguments)
}

// Handler implements component.HTTPComponent.
func (c *Component) Handler() http.Handler {
	r := mux.NewRouter()

	fa := api.NewFlowAPI(c.ctrl, r)
	fa.RegisterRoutes("/", r)

	r.PathPrefix("/{id}/").Handler(c.ctrl.ComponentHandler())
	return r
}
