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

// ModuleComponent holds the common properties for module components.
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

// NewModuleComponent initializes a new ModuleComponent.
func NewModuleComponent(o component.Options) *ModuleComponent {
	// TODO(rfratto): replace these with a tracer/registry which properly
	// propagates data back to the parent.
	flowTracer, _ := tracing.New(tracing.DefaultOptions)
	flowRegistry := prometheus.NewRegistry()

	return &ModuleComponent{
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
			ModuleDepth:    o.ModuleDepth + 1,
			MaxModuleDepth: o.MaxModuleDepth,
		}),
	}
}

// LoadFlowContent loads the flow controller with the current component content. It
// will set the component health in addition to return the error so that the consumer
// can rely on either or both.
func (c *ModuleComponent) LoadFlowContent(arguments map[string]any, contentValue string) error {
	f, err := flow.ReadFile(c.opts.ID, []byte(contentValue))
	if err != nil {
		c.setHealth(component.Health{
			Health:     component.HealthTypeUnhealthy,
			Message:    fmt.Sprintf("failed to parse module content: %s", err),
			UpdateTime: time.Now(),
		})

		return err
	}

	err = c.ctrl.LoadFile(f, arguments)
	if err != nil {
		c.setHealth(component.Health{
			Health:     component.HealthTypeUnhealthy,
			Message:    fmt.Sprintf("failed to load module content: %s", err),
			UpdateTime: time.Now(),
		})

		return err
	}

	c.setHealth(component.Health{
		Health:     component.HealthTypeHealthy,
		Message:    "module content loaded",
		UpdateTime: time.Now(),
	})
	return nil
}

// RunFlowController runs the flow controller that all module components start.
func (c *ModuleComponent) RunFlowController(ctx context.Context) {
	c.ctrl.Run(ctx)
}

// CurrentHealth contains the implementation details for CurrentHealth in a module component.
func (c *ModuleComponent) CurrentHealth() component.Health {
	c.mut.RLock()
	defer c.mut.RUnlock()
	return c.health
}

// SetHealth contains the implementation details for setHealth in a module component.
func (c *ModuleComponent) setHealth(h component.Health) {
	c.mut.Lock()
	defer c.mut.Unlock()
	c.health = h
}

// Handler contains the implementation details for Handler in a module component.
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
