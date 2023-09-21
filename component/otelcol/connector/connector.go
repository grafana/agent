// Package connector exposes utilities to create a Flow component from
// OpenTelemetry Collector connectors.
package connector

import (
	"context"
	"errors"
	"os"

	"github.com/grafana/agent/component"
	"github.com/grafana/agent/component/otelcol"
	"github.com/grafana/agent/component/otelcol/internal/fanoutconsumer"
	"github.com/grafana/agent/component/otelcol/internal/lazycollector"
	"github.com/grafana/agent/component/otelcol/internal/lazyconsumer"
	"github.com/grafana/agent/component/otelcol/internal/scheduler"
	"github.com/grafana/agent/pkg/build"
	"github.com/grafana/agent/pkg/util/zapadapter"
	"github.com/prometheus/client_golang/prometheus"
	otelcomponent "go.opentelemetry.io/collector/component"
	otelconnector "go.opentelemetry.io/collector/connector"
	otelextension "go.opentelemetry.io/collector/extension"
	sdkprometheus "go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/sdk/metric"
)

const (
	ConnectorTracesToTraces = iota
	ConnectorTracesToMetrics
	ConnectorTracesToLogs
	ConnectorMetricsToTraces
	ConnectorMetricsToMetrics
	ConnectorMetricsToLogs
	ConnectorLogsToTraces
	ConnectorLogsToMetrics
	ConnectorLogsToLogs
)

// Arguments is an extension of component.Arguments which contains necessary
// settings for OpenTelemetry Collector connectors.
type Arguments interface {
	component.Arguments

	// Convert converts the Arguments into an OpenTelemetry Collector connector
	// configuration.
	Convert() (otelcomponent.Config, error)

	// Extensions returns the set of extensions that the configured component is
	// allowed to use.
	Extensions() map[otelcomponent.ID]otelextension.Extension

	// Exporters returns the set of exporters that are exposed to the configured
	// component.
	Exporters() map[otelcomponent.DataType]map[otelcomponent.ID]otelcomponent.Component

	// NextConsumers returns the set of consumers to send data to.
	NextConsumers() *otelcol.ConsumerArguments

	ConnectorType() int
}

// Connector is a Flow component shim which manages an OpenTelemetry Collector
// connector component.
type Connector struct {
	ctx    context.Context
	cancel context.CancelFunc

	opts     component.Options
	factory  otelconnector.Factory
	consumer *lazyconsumer.Consumer

	sched     *scheduler.Scheduler
	collector *lazycollector.Collector
}

var (
	_ component.Component       = (*Connector)(nil)
	_ component.HealthComponent = (*Connector)(nil)
)

// New creates a new Flow component which encapsulates an OpenTelemetry
// Collector connector. args must hold a value of the argument type registered
// with the Flow component.
//
// The registered component must be registered to export the
// otelcol.ConsumerExports type, otherwise New will panic.
func New(opts component.Options, f otelconnector.Factory, args Arguments) (*Connector, error) {
	ctx, cancel := context.WithCancel(context.Background())

	consumer := lazyconsumer.New(ctx)

	// Create a lazy collector where metrics from the upstream component will be
	// forwarded.
	collector := lazycollector.New()
	opts.Registerer.MustRegister(collector)

	// Immediately set our state with our consumer. The exports will never change
	// throughout the lifetime of our component.
	//
	// This will panic if the wrapping component is not registered to export
	// otelcol.ConsumerExports.
	opts.OnStateChange(otelcol.ConsumerExports{Input: consumer})

	p := &Connector{
		ctx:    ctx,
		cancel: cancel,

		opts:     opts,
		factory:  f,
		consumer: consumer,

		sched:     scheduler.New(opts.Logger),
		collector: collector,
	}
	if err := p.Update(args); err != nil {
		return nil, err
	}
	return p, nil
}

// Run starts the Connector component.
func (p *Connector) Run(ctx context.Context) error {
	defer p.cancel()
	return p.sched.Run(ctx)
}

// Update implements component.Component. It will convert the Arguments into
// configuration for OpenTelemetry Collector connector configuration and manage
// the underlying OpenTelemetry Collector connector.
func (p *Connector) Update(args component.Arguments) error {
	pargs := args.(Arguments)

	host := scheduler.NewHost(
		p.opts.Logger,
		scheduler.WithHostExtensions(pargs.Extensions()),
		scheduler.WithHostExporters(pargs.Exporters()),
	)

	reg := prometheus.NewRegistry()
	p.collector.Set(reg)

	promExporter, err := sdkprometheus.New(sdkprometheus.WithRegisterer(reg), sdkprometheus.WithoutTargetInfo())
	if err != nil {
		return err
	}

	settings := otelconnector.CreateSettings{
		TelemetrySettings: otelcomponent.TelemetrySettings{
			Logger: zapadapter.New(p.opts.Logger),

			TracerProvider: p.opts.Tracer,
			MeterProvider:  metric.NewMeterProvider(metric.WithReader(promExporter)),
		},

		BuildInfo: otelcomponent.BuildInfo{
			Command:     os.Args[0],
			Description: "Grafana Agent",
			Version:     build.Version,
		},
	}

	connectorConfig, err := pargs.Convert()
	if err != nil {
		return err
	}

	next := pargs.NextConsumers()

	// Create instances of the connector from our factory for each of our
	// supported telemetry signals.
	var components []otelcomponent.Component

	var tracesConnector otelconnector.Traces
	var metricsConnector otelconnector.Metrics
	var logsConnector otelconnector.Logs

	switch pargs.ConnectorType() {
	case ConnectorTracesToMetrics:
		if len(next.Traces) > 0 || len(next.Logs) > 0 {
			return errors.New("this connector can only output metrics")
		}

		if len(next.Metrics) > 0 {
			nextMetrics := fanoutconsumer.Metrics(next.Metrics)
			tracesConnector, err = p.factory.CreateTracesToMetrics(p.ctx, settings, connectorConfig, nextMetrics)
			if err != nil && !errors.Is(err, otelcomponent.ErrDataTypeIsNotSupported) {
				return err
			} else if tracesConnector != nil {
				components = append(components, tracesConnector)
			}
		}
	default:
		return errors.New("unsupported connector type")
	}

	// Schedule the components to run once our component is running.
	p.sched.Schedule(host, components...)
	p.consumer.SetConsumers(tracesConnector, metricsConnector, logsConnector)
	return nil
}

// CurrentHealth implements component.HealthComponent.
func (p *Connector) CurrentHealth() component.Health {
	return p.sched.CurrentHealth()
}
