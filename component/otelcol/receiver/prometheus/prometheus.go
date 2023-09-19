// Package prometheus provides an otelcol.receiver.prometheus component.
package prometheus

import (
	"context"
	"os"
	"regexp"
	"sync"
	"time"

	"github.com/go-kit/log"
	"github.com/grafana/agent/component"
	"github.com/grafana/agent/component/otelcol"
	"github.com/grafana/agent/component/otelcol/internal/fanoutconsumer"
	"github.com/grafana/agent/component/otelcol/receiver/prometheus/internal"
	"github.com/grafana/agent/pkg/build"
	"github.com/grafana/agent/pkg/util/zapadapter"
	otel_service "github.com/grafana/agent/service/otel"
	"github.com/prometheus/prometheus/model/labels"
	"github.com/prometheus/prometheus/storage"
	otelcomponent "go.opentelemetry.io/collector/component"
	otelreceiver "go.opentelemetry.io/collector/receiver"
	"go.opentelemetry.io/otel/metric/noop"
	"go.opentelemetry.io/otel/trace"
)

func init() {
	component.Register(component.Registration{
		Name:          "otelcol.receiver.prometheus",
		Args:          Arguments{},
		Exports:       Exports{},
		NeedsServices: []string{otel_service.ServiceName},

		Build: func(o component.Options, a component.Arguments) (component.Component, error) {
			return New(o, a.(Arguments))
		},
	})
}

// Arguments configures the otelcol.receiver.prometheus component.
type Arguments struct {
	// Output configures where to send received data. Required.
	Output *otelcol.ConsumerArguments `river:"output,block"`
}

// Exports are the set of fields exposed by the otelcol.receiver.prometheus
// component.
type Exports struct {
	Receiver storage.Appendable `river:"receiver,attr"`
}

// Component is the otelcol.receiver.prometheus component.
type Component struct {
	log  log.Logger
	opts component.Options

	mut        sync.RWMutex
	cfg        Arguments
	appendable storage.Appendable
}

var _ component.Component = (*Component)(nil)

// New creates a new otelcol.receiver.prometheus component.
func New(o component.Options, c Arguments) (*Component, error) {
	res := &Component{
		log:  o.Logger,
		opts: o,
	}

	if err := res.Update(c); err != nil {
		return nil, err
	}
	return res, nil
}

// Run implements Component.
func (c *Component) Run(ctx context.Context) error {
	<-ctx.Done()
	return nil
}

// Update implements Component.
func (c *Component) Update(newConfig component.Arguments) error {
	c.mut.Lock()
	defer c.mut.Unlock()

	cfg := newConfig.(Arguments)
	c.cfg = cfg

	// useStartTimeMetric is used to configure the 'metrics adjuster' in the
	// prometheusreceiver which is applied whenever a Commit is called.
	// If set to true, the receiver will utilize a `startTimeMetricAdjuster`
	// to adjust metric start times based on a start time metric. The start
	// time metric defaults to `process_start_time_seconds`, but can be
	// overridden by using this regex.
	//
	// gcInterval should be at least as long as the longest scrape interval
	// used by the upstream scrape configs, plus a delta to avoid race
	// conditions so that jobs are not getting GC'ed between scrapes.
	var (
		useStartTimeMetric   = false
		startTimeMetricRegex *regexp.Regexp

		gcInterval = 5 * time.Minute
	)
	settings := otelreceiver.CreateSettings{
		TelemetrySettings: otelcomponent.TelemetrySettings{
			Logger: zapadapter.New(c.opts.Logger),

			// TODO(tpaschalis): expose tracing and logging statistics.
			TracerProvider: trace.NewNoopTracerProvider(),
			MeterProvider:  noop.NewMeterProvider(),
		},

		BuildInfo: otelcomponent.BuildInfo{
			Command:     os.Args[0],
			Description: "Grafana Agent",
			Version:     build.Version,
		},
	}
	metricsSink := fanoutconsumer.Metrics(cfg.Output.Metrics)

	appendable, err := internal.NewAppendable(
		metricsSink,
		settings,
		gcInterval,
		useStartTimeMetric,
		startTimeMetricRegex,
		otelcomponent.NewID(otelcomponent.Type(c.opts.ID)),
		labels.Labels{},
	)
	if err != nil {
		return err
	}
	c.appendable = appendable

	// Export the receiver.
	c.opts.OnStateChange(Exports{Receiver: c.appendable})

	return nil
}
