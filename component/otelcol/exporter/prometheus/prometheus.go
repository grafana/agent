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
	"github.com/grafana/agent/service/labelstore"
	"github.com/prometheus/prometheus/model/exemplar"
	"github.com/prometheus/prometheus/model/histogram"
	"github.com/prometheus/prometheus/model/labels"
	"github.com/prometheus/prometheus/model/metadata"
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
	IncludeTargetInfo             bool                 `river:"include_target_info,attr,optional"`
	IncludeScopeInfo              bool                 `river:"include_scope_info,attr,optional"`
	IncludeScopeLabels            bool                 `river:"include_scope_labels,attr,optional"`
	GCFrequency                   time.Duration        `river:"gc_frequency,attr,optional"`
	ForwardTo                     []storage.Appendable `river:"forward_to,attr"`
	AddMetricSuffixes             bool                 `river:"add_metric_suffixes,attr,optional"`
	ResourceToTelemetryConversion bool                 `river:"resource_to_telemetry_conversion,attr,optional"`
}

// DefaultArguments holds defaults values.
var DefaultArguments = Arguments{
	IncludeTargetInfo:             true,
	IncludeScopeInfo:              false,
	IncludeScopeLabels:            true,
	GCFrequency:                   5 * time.Minute,
	AddMetricSuffixes:             true,
	ResourceToTelemetryConversion: false,
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

	mut                 sync.RWMutex
	cfg                 Arguments
	debugStreamCallback func(func() string)
}

var _ component.Component = (*Component)(nil)

// New creates a new otelcol.exporter.prometheus component.
func New(o component.Options, c Arguments) (*Component, error) {
	service, err := o.GetServiceData(labelstore.ServiceName)
	if err != nil {
		return nil, err
	}
	ls := service.(labelstore.LabelStore)

	res := &Component{
		log:                 o.Logger,
		opts:                o,
		debugStreamCallback: func(func() string) {},
	}

	res.fanout = prometheus.NewFanout(nil, o.ID, o.Registerer, ls)

	interceptor := prometheus.NewInterceptor(res.fanout, ls,
		prometheus.WithAppendHook(func(globalRef storage.SeriesRef, l labels.Labels, t int64, v float64, next storage.Appender) (storage.SeriesRef, error) {
			localID := ls.GetLocalRefID(res.opts.ID, uint64(globalRef))
			_, nextErr := next.Append(storage.SeriesRef(localID), l, t, v)
			res.debugStreamCallback(func() string { return fmt.Sprintf("ts=%d, labels=%s, value=%f", t, l, v) })
			return globalRef, nextErr
		}),
		prometheus.WithHistogramHook(func(globalRef storage.SeriesRef, l labels.Labels, t int64, h *histogram.Histogram, fh *histogram.FloatHistogram, next storage.Appender) (storage.SeriesRef, error) {
			localID := ls.GetLocalRefID(res.opts.ID, uint64(globalRef))
			_, nextErr := next.AppendHistogram(storage.SeriesRef(localID), l, t, h, fh)
			res.debugStreamCallback(func() string {
				if h != nil {
					return fmt.Sprintf("ts=%d, labels=%s, histogram=%s", t, l, h.String())
				} else if fh != nil {
					return fmt.Sprintf("ts=%d, labels=%s, float_histogram=%s", t, l, fh.String())
				}
				return fmt.Sprintf("ts=%d, labels=%s, no_value", t, l)
			})
			return globalRef, nextErr
		}),
		prometheus.WithMetadataHook(func(globalRef storage.SeriesRef, l labels.Labels, m metadata.Metadata, next storage.Appender) (storage.SeriesRef, error) {
			localID := ls.GetLocalRefID(res.opts.ID, uint64(globalRef))
			_, nextErr := next.UpdateMetadata(storage.SeriesRef(localID), l, m)
			res.debugStreamCallback(func() string {
				return fmt.Sprintf("labels=%s, type=%s, unit=%s, help=%s", l, m.Type, m.Unit, m.Help)
			})
			return globalRef, nextErr
		}),
		prometheus.WithExemplarHook(func(globalRef storage.SeriesRef, l labels.Labels, e exemplar.Exemplar, next storage.Appender) (storage.SeriesRef, error) {
			localID := ls.GetLocalRefID(res.opts.ID, uint64(globalRef))
			_, nextErr := next.AppendExemplar(storage.SeriesRef(localID), l, e)
			res.debugStreamCallback(func() string {
				return fmt.Sprintf("ts=%d, labels=%s, exemplar_labels=%s, value=%f", e.Ts, l, e.Labels, e.Value)
			})
			return globalRef, nextErr
		}),
	)

	res.converter = convert.New(o.Logger, interceptor, convertArgumentsToConvertOptions(c))

	if err := res.Update(c); err != nil {
		return nil, err
	}

	// Construct a consumer based on our converter and export it. This will
	// remain the same throughout the component's lifetime, so we do this during
	// component construction.
	export := lazyconsumer.New(context.Background())
	export.SetConsumers(nil, res.converter, nil)
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
	c.converter.UpdateOptions(convertArgumentsToConvertOptions(cfg))

	// If our forward_to argument changed, we need to flush the metadata cache to
	// ensure the new children have all the metadata they need.
	//
	// For now, we always flush whenever we update, but we could do something
	// more intelligent here in the future.
	c.converter.FlushMetadata()
	return nil
}

func convertArgumentsToConvertOptions(args Arguments) convert.Options {
	return convert.Options{
		IncludeTargetInfo:             args.IncludeTargetInfo,
		IncludeScopeInfo:              args.IncludeScopeInfo,
		AddMetricSuffixes:             args.AddMetricSuffixes,
		ResourceToTelemetryConversion: args.ResourceToTelemetryConversion,
	}
}

func (c *Component) HookDebugStream(active bool, debugStreamCallback func(computeDataFunc func() string)) {
	c.debugStreamCallback = debugStreamCallback
}
