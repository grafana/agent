// Package string defines the module.string component.
package string

import (
	"context"
	"net/http"
	"os"
	"sync"
	"time"

	"go.uber.org/atomic"

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
	Content rivertypes.Secret `river:"content,attr"`

	// Arguments to pass into the module.
	Arguments map[string]any `river:"arguments,attr,optional"`
}

// Exports holds values which are exported from the run module.
type Exports struct {
	// Values exported from the running module.
	Values map[string]any `river:"values,attr"`
}

// Component implements the module.string component.
type Component struct {
	opts component.Options
	log  log.Logger
	ctrl *flow.Flow

	moduleArgs        atomic.Pointer[map[string]any]
	file              atomic.Pointer[flow.File]
	updateContollerCh chan struct{}

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
				o.OnStateChange(Exports{Values: exports})
			},
		}),

		updateContollerCh: make(chan struct{}, 1),
	}
	if err := c.Update(args); err != nil {
		return nil, err
	}
	return c, nil
}

// Run implements component.Component.
func (c *Component) Run(ctx context.Context) error {
Loop:
	for {
		select {
		case <-ctx.Done():
			break Loop
		case <-c.updateContollerCh:
			err := c.ctrl.LoadFile(c.file.Load(), *c.moduleArgs.Load())
			c.updateHealth(err)
		}
	}

	return c.ctrl.Close()
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

	f, err := flow.ReadFile(c.opts.ID, []byte(newArgs.Content))
	if err != nil {
		return err
	}

	// TODO(rfratto): sync exports with current set rather than updating it in
	// full every time.
	var emptyExports = map[string]any{}

	for _, b := range f.ConfigBlocks {
		if b.Name[0] == "export" {
			emptyExports[b.Label] = nil
		}
	}

	// Create an initial exported value which contains all the keys that
	// correspond with the exports from our module so that sibling components can
	// properly reference them by name, albeit with the zero value.
	//
	// Currently, this runs every time Update is called, so there will always be
	// two calls to OnStateChange happening every time: once which sets everything
	// to the zero value, and once where the evaluated concrete values are
	// exposed (in the callback to OnExportsChange).
	//
	// TODO(rfratto): find a way to not override the last evaluated value with
	// null.
	c.opts.OnStateChange(Exports{Values: emptyExports})

	c.file.Store(f)
	c.moduleArgs.Store(&newArgs.Arguments)

	select {
	case c.updateContollerCh <- struct{}{}:
	default: // no-op
	}

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
