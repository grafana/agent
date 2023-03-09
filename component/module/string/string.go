// Package string defines the module.string component.
package string

import (
	"context"
	"net/http"
	"os"
	"sync"
	"time"

>>>>>>> 9d35748c (module.string: new component)
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

	healthMut sync.RWMutex
	health    component.Health
}

var (
	_ component.Component       = (*Component)(nil)
	_ component.HealthComponent = (*Component)(nil)
	_ component.HTTPComponent   = (*Component)(nil)
)

// New creates a new module.string component.
func New(o component.Options, args Arguments) (*Component, error) {
	// TODO(rfratto): replace these with a logger/tracer/registry which properly
	// propagates data back to the parent.
	flowLogger, _ := logging.New(os.Stderr, logging.Options{
		Level:  logging.LevelDebug,
		Format: logging.FormatLogfmt,
	})
	flowTracer, _ := tracing.New(tracing.DefaultOptions)
	flowRegistry := prometheus.NewRegistry()

	c := &Component{
		opts: o,
		log:  o.Logger,

		ctrl: flow.New(flow.Options{
			ControllerID: o.ID,
			Logger:       flowLogger,
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

func (c *Component) updateHealth(err error) {
	c.healthMut.Lock()
	defer c.healthMut.Unlock()

	if err == nil {
		c.health = component.Health{
			Health:     component.HealthTypeHealthy,
			Message:    "module updated",
			UpdateTime: time.Now(),
		}
	} else {
		c.health = component.Health{
			Health:     component.HealthTypeUnhealthy,
			Message:    err.Error(),
			UpdateTime: time.Now(),
		}
	}
}

// Update implements component.Component.
func (c *Component) Update(args component.Arguments) error {
	newArgs := args.(Arguments)

	f, err := flow.ReadFile(c.opts.ID, []byte(newArgs.Content.Value))
	if err != nil {
		return err
	}

	err = c.ctrl.LoadFile(f, newArgs.Arguments)
	c.updateHealth(err)

	return nil
}

// CurrentHealth implements component.HealthComponent.
func (c *Component) CurrentHealth() component.Health {
	c.healthMut.RLock()
	defer c.healthMut.RUnlock()

	return c.health
}

// Handler implements component.HTTPComponent.
func (c *Component) Handler() http.Handler {
	r := mux.NewRouter()

	fa := api.NewFlowAPI(c.ctrl, r)
	fa.RegisterRoutes("/", r)

	r.PathPrefix("/{id}/").Handler(c.ctrl.ComponentHandler())
	return r
}
