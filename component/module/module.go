package module

import (
	"context"
	"fmt"
	"net/http"
	"path"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/grafana/agent/component"
	"github.com/grafana/agent/pkg/flow"
	"github.com/grafana/agent/pkg/flow/logging"
	"github.com/grafana/agent/pkg/flow/tracing"
	"github.com/grafana/agent/web/api"
	"github.com/prometheus/client_golang/prometheus"
)

// ModuleComponent implements the module.file component.
type ModuleComponent struct {
	opts component.Options
	ctrl *flow.Flow

	mut    sync.RWMutex
	health component.Health
}

// Exports holds values which are exported from the run module.
type Exports struct {
	// Exports exported from the running module.
	Exports map[string]any `river:"exports,attr"`
}

func NewComponent(o component.Options) ModuleComponent {
	// TODO(rfratto): replace these with a tracer/registry which properly
	// propagates data back to the parent.
	flowTracer, _ := tracing.New(tracing.DefaultOptions)
	flowRegistry := prometheus.NewRegistry()

	return ModuleComponent{
		opts: o,
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
}

// LoadFlowContent loads the flow controller with the current component content.
func (c *ModuleComponent) LoadFlowContent(arguments map[string]any, contentValue []byte) error {
	c.mut.RLock()
	f, err := flow.ReadFile(c.opts.ID, []byte(contentValue))
	c.mut.RUnlock()
	if err != nil {
		c.SetHealth(component.Health{
			Health:     component.HealthTypeUnhealthy,
			Message:    fmt.Sprintf("failed to parse module content: %s", err),
			UpdateTime: time.Now(),
		})
	} else {
		c.mut.RLock()
		err = c.ctrl.LoadFile(f, arguments)
		c.mut.RUnlock()
		if err != nil {
			c.SetHealth(component.Health{
				Health:     component.HealthTypeUnhealthy,
				Message:    fmt.Sprintf("failed to load module content: %s", err),
				UpdateTime: time.Now(),
			})
		} else {
			c.SetHealth(component.Health{
				Health:     component.HealthTypeHealthy,
				Message:    "module content loaded",
				UpdateTime: time.Now(),
			})
		}

		return err
	}

	return nil
}

func (c *ModuleComponent) RunFlowController(ctx context.Context) {
	c.ctrl.Run(ctx)
}

// CurrentHealth implements component.HealthComponent.
func (c *ModuleComponent) CurrentHealth() component.Health {
	c.mut.RLock()
	defer c.mut.RUnlock()
	return c.health
}

func (c *ModuleComponent) SetHealth(h component.Health) {
	c.mut.Lock()
	defer c.mut.Unlock()
	c.health = h
}

// Handler implements component.HTTPComponent.
func (c *ModuleComponent) Handler() http.Handler {
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
