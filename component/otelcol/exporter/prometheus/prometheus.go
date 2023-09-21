// Package prometheus provides an otelcol.exporter.prometheus component.
package prometheus

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/go-kit/log"
	"github.com/grafana/agent/component"
	"github.com/grafana/agent/component/otelcol"
	"github.com/grafana/agent/component/otelcol/exporter/prometheus/internal/convert"
	"github.com/grafana/agent/component/otelcol/internal/lazyconsumer"
	"github.com/grafana/agent/component/prometheus"
	"github.com/prometheus/prometheus/storage"
)

func init() {
	component.Register(component.Registration{
		Name:    "otelcol.exporter.prometheus",
		Args:    Arguments{},
		Exports: otelcol.ConsumerExports{},

		Build: func(o component.Options, a component.Arguments) (component.Component, error) {
			return New(o, a.(Arguments))
		},
	})
}

// Arguments configures the otelcol.exporter.prometheus component.
type Arguments struct {
	IncludeTargetInfo  bool                 `river:"include_target_info,attr,optional"`
	IncludeScopeInfo   bool                 `river:"include_scope_info,attr,optional"`
	IncludeScopeLabels bool                 `river:"include_scope_labels,attr,optional"`
	GCFrequency        time.Duration        `river:"gc_frequency,attr,optional"`
	ForwardTo          []storage.Appendable `river:"forward_to,attr"`
}

// DefaultArguments holds defaults values.
var DefaultArguments = Arguments{
	IncludeTargetInfo:  true,
	IncludeScopeInfo:   false,
	IncludeScopeLabels: true,
	GCFrequency:        5 * time.Minute,
}

// SetToDefault implements river.Defaulter.
func (args *Arguments) SetToDefault() {
	*args = DefaultArguments
}

// Validate implements river.Validator.
func (args *Arguments) Validate() error {
	if args.GCFrequency == 0 {
		return fmt.Errorf("gc_frequency must be greater than 0")
	}

	return nil
}

// Component is the otelcol.exporter.prometheus component.
type Component struct {
	log  log.Logger
	opts component.Options

	fanout    *prometheus.Fanout
	converter *convert.Converter

	mut sync.RWMutex
	cfg Arguments
}

var _ component.Component = (*Component)(nil)

// New creates a new otelcol.exporter.prometheus component.
func New(o component.Options, c Arguments) (*Component, error) {
	fanout := prometheus.NewFanout(nil, o.ID, o.Registerer)

	converter := convert.New(o.Logger, fanout, convert.Options{
		IncludeTargetInfo: true,
		IncludeScopeInfo:  false,
	})

	res := &Component{
		log:  o.Logger,
		opts: o,

		fanout:    fanout,
		converter: converter,
	}
	if err := res.Update(c); err != nil {
		return nil, err
	}

	// Construct a consumer based on our converter and export it. This will
	// remain the same throughout the component's lifetime, so we do this during
	// component construction.
	export := lazyconsumer.New(context.Background())
	export.SetConsumers(nil, converter, nil)
	o.OnStateChange(otelcol.ConsumerExports{Input: export})

	return res, nil
}

// Run implements Component.
func (c *Component) Run(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-time.After(c.nextGC()):
			// TODO(rfratto): we may want to consider making this an option in the
			// future, but hard-coding to 5 minutes is a reasonable default to start
			// with.
			c.converter.GC(5 * time.Minute)
		}
	}
}

func (c *Component) nextGC() time.Duration {
	c.mut.RLock()
	defer c.mut.RUnlock()
	return c.cfg.GCFrequency
}

// Update implements Component.
func (c *Component) Update(newConfig component.Arguments) error {
	c.mut.Lock()
	defer c.mut.Unlock()

	cfg := newConfig.(Arguments)
	c.cfg = cfg

	c.fanout.UpdateChildren(cfg.ForwardTo)
	c.converter.UpdateOptions(convert.Options{
		IncludeTargetInfo: cfg.IncludeTargetInfo,
		IncludeScopeInfo:  cfg.IncludeScopeInfo,
	})

	// If our forward_to argument changed, we need to flush the metadata cache to
	// ensure the new children have all the metadata they need.
	//
	// For now, we always flush whenever we update, but we could do something
	// more intelligent here in the future.
	c.converter.FlushMetadata()
	return nil
}
