package traces

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/connector"
	otelexporter "go.opentelemetry.io/collector/exporter"
	"go.opentelemetry.io/collector/extension"
	"go.opentelemetry.io/collector/featuregate"
	"go.opentelemetry.io/collector/otelcol"
	"go.opentelemetry.io/collector/processor"
	"go.opentelemetry.io/collector/receiver"
	"go.opentelemetry.io/collector/service"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/grafana/agent/pkg/build"
	"github.com/grafana/agent/pkg/logs"
	"github.com/grafana/agent/pkg/metrics/instance"
	"github.com/grafana/agent/pkg/traces/automaticloggingprocessor"
	"github.com/grafana/agent/pkg/traces/contextkeys"
	"github.com/grafana/agent/pkg/traces/servicegraphprocessor"
	"github.com/grafana/agent/pkg/util"
	prom_client "github.com/prometheus/client_golang/prometheus"
	otelprom "go.opentelemetry.io/otel/exporters/prometheus"
)

// Instance wraps the OpenTelemetry collector to enable tracing pipelines
type Instance struct {
	mut    sync.Mutex
	cfg    InstanceConfig
	logger *zap.Logger
	// metricViews []*view.View

	factories otelcol.Factories
	service   *service.Service
}

//TODO: Do we need this?
// var _ component.Host = (*Instance)(nil)

// NewInstance creates and starts an instance of tracing pipelines.
func NewInstance(logsSubsystem *logs.Logs, cfg InstanceConfig, logger *zap.Logger, promInstanceManager instance.Manager, reg prom_client.Registerer) (*Instance, error) {
	// var err error

	instance := &Instance{}
	instance.logger = logger
	// instance.metricViews, err = newMetricViews(reg)
	// if err != nil {
	// 	return nil, fmt.Errorf("failed to create metric views: %w", err)
	// }

	if err := instance.ApplyConfig(logsSubsystem, promInstanceManager, cfg, reg); err != nil {
		return nil, err
	}
	return instance, nil
}

// ApplyConfig updates the configuration of the Instance.
func (i *Instance) ApplyConfig(logsSubsystem *logs.Logs, promInstanceManager instance.Manager, cfg InstanceConfig, reg prom_client.Registerer) error {
	i.mut.Lock()
	defer i.mut.Unlock()

	if util.CompareYAML(cfg, i.cfg) {
		// No config change
		return nil
	}
	i.cfg = cfg

	// Shut down any existing pipeline
	i.stop()

	err := i.buildAndStartPipeline(context.Background(), cfg, logsSubsystem, promInstanceManager, reg)
	if err != nil {
		return fmt.Errorf("failed to create pipeline: %w", err)
	}

	return nil
}

// Stop stops the OpenTelemetry collector subsystem
func (i *Instance) Stop() {
	i.mut.Lock()
	defer i.mut.Unlock()

	i.stop()
	// view.Unregister(i.metricViews...)
}

func (i *Instance) stop() {
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if i.service != nil {
		//TODO: Should we discard the error? At least log it?
		_ = i.service.Shutdown(shutdownCtx)
	}
}

func (i *Instance) buildAndStartPipeline(ctx context.Context, cfg InstanceConfig, logs *logs.Logs, instManager instance.Manager, reg prom_client.Registerer) error {
	// create component factories
	otelConfig, err := cfg.otelConfig()
	if err != nil {
		return fmt.Errorf("failed to load otelConfig from agent traces config: %w", err)
	}
	for _, rw := range cfg.RemoteWrite {
		if rw.InsecureSkipVerify {
			i.logger.Warn("Configuring TLS with insecure_skip_verify. Use tls_config.insecure_skip_verify instead")
		}
		if rw.TLSConfig != nil && rw.TLSConfig.ServerName != "" {
			i.logger.Warn("Configuring unsupported tls_config.server_name")
		}
	}

	if cfg.SpanMetrics != nil && len(cfg.SpanMetrics.MetricsInstance) != 0 {
		ctx = context.WithValue(ctx, contextkeys.Metrics, instManager)
	}

	if cfg.LoadBalancing == nil && (cfg.TailSampling != nil || cfg.ServiceGraphs != nil) {
		i.logger.Warn("Configuring tail_sampling and/or service_graphs without load_balance." +
			"Load balancing is required for those features to properly work in multi agent deployments")
	}

	if cfg.AutomaticLogging != nil && cfg.AutomaticLogging.Backend != automaticloggingprocessor.BackendStdout {
		ctx = context.WithValue(ctx, contextkeys.Logs, logs)
	}

	factories, err := tracingFactories()
	if err != nil {
		return fmt.Errorf("failed to load tracing factories: %w", err)
	}
	// i.factories = factories

	componentId := "grafana-agent/" + cfg.Name
	appinfo := component.BuildInfo{
		Command:     componentId,
		Description: componentId,
		Version:     build.Version,
	}

	//TODO: Delete this later

	// settings := component.TelemetrySettings{
	// 	Logger:         i.logger,
	// 	TracerProvider: trace.NewNoopTracerProvider(),
	// 	MeterProvider:  metric.NewNoopMeterProvider(),
	// }

	// start extensions
	// i.extensions, err = extensions.New(ctx, extensions.Settings{
	// 	Telemetry: settings,
	// 	BuildInfo: appinfo,

	// 	Factories: factories.Extensions,
	// 	Configs:   otelConfig,
	// }, otelConfig.Extensions)
	// if err != nil {
	// 	i.logger.Error(fmt.Sprintf("failed to build extensions: %s", err.Error()))
	// 	return fmt.Errorf("failed to create extensions builder: %w", err)
	// }
	// err = i.extensions.Start(ctx, i)
	// if err != nil {
	// 	i.logger.Error(fmt.Sprintf("failed to start extensions: %s", err.Error()))
	// 	return fmt.Errorf("failed to start extensions: %w", err)
	// }

	// i.pipelines, err = pipelines.Build(ctx, pipelines.Settings{
	// 	Telemetry: settings,
	// 	BuildInfo: appinfo,

	// 	ReceiverFactories:  factories.Receivers,
	// 	ReceiverConfigs:    otelConfig.Receivers,
	// 	ProcessorFactories: factories.Processors,
	// 	ProcessorConfigs:   otelConfig.Processors,
	// 	ExporterFactories:  factories.Exporters,
	// 	ExporterConfigs:    otelConfig.Exporters,

	// 	PipelineConfigs: otelConfig.Pipelines,
	// })
	// if err != nil {
	// 	return fmt.Errorf("failed to create pipelines: %w", err)
	// }
	// if err := i.pipelines.StartAll(ctx, i); err != nil {
	// 	i.logger.Error(fmt.Sprintf("failed to start pipelines: %s", err.Error()))
	// 	return fmt.Errorf("failed to start pipelines: %w", err)
	// }

	// return i.extensions.NotifyPipelineReady()

	// var resAttrs []attribute.KeyValue
	// for k, v := range attrs {
	// 	resAttrs = append(resAttrs, attribute.String(k, v))
	// }

	// res, err := resource.New(context.Background(), resource.WithAttributes(resAttrs...))
	// if err != nil {
	// 	return fmt.Errorf("error creating otel resources: %w", err)
	// }

	// ----------------------------
	// wrappedRegisterer := prometheus.WrapRegistererWithPrefix("otelcol_", reg)
	// exporter, err := otelprom.New(
	// 	otelprom.WithRegisterer(wrappedRegisterer),
	// 	otelprom.WithoutUnits(),
	// 	// Disabled for the moment until this becomes stable, and we are ready to break backwards compatibility.
	// 	otelprom.WithoutScopeInfo())
	// if err != nil {
	// 	return fmt.Errorf("error creating otel prometheus exporter: %w", err)
	// }
	// mp = sdkmetric.NewMeterProvider(
	// 	// sdkmetric.WithResource(res),
	// 	sdkmetric.WithReader(exporter),
	// 	// sdkmetric.WithView(batchViews()...),
	// )
	// ----------------------------

	//TODO: Remove this later. How should we set the otel logging level via config?
	otelConfig.Service.Telemetry.Logs.Level = zapcore.DebugLevel

	//TODO: Can this feature gate remain enabled?
	fgReg := featuregate.GlobalRegistry()
	fgReg.Set("telemetry.useOtelForInternalMetrics", true)

	//TODO: Review the arguments passed to otelprom.New
	otelExporter, err := otelprom.New(
		otelprom.WithRegisterer(reg),
		otelprom.WithoutUnits(),
		// Disabled for the moment until this becomes stable, and we are ready to break backwards compatibility.
		otelprom.WithoutScopeInfo())
	if err != nil {
		return fmt.Errorf("error creating otel prometheus exporter: %w", err)
	}

	i.service, err = service.New(ctx, service.Settings{
		BuildInfo:  appinfo,
		Receivers:  receiver.NewBuilder(otelConfig.Receivers, factories.Receivers),
		Processors: processor.NewBuilder(otelConfig.Processors, factories.Processors),
		Exporters:  otelexporter.NewBuilder(otelConfig.Exporters, factories.Exporters),
		Connectors: connector.NewBuilder(otelConfig.Connectors, factories.Connectors),
		Extensions: extension.NewBuilder(otelConfig.Extensions, factories.Extensions),
		//TODO: Maybe we should make this more generic so that we pull views from all processors?
		OtelMetricViews:  servicegraphprocessor.OtelMetricViews(),
		OtelMetricReader: *otelExporter,
		//TODO: Fill these later?
		// AsyncErrorChannel: col.asyncErrorChannel,
		// LoggingOptions:    col.set.LoggingOptions,
	}, otelConfig.Service)
	if err != nil {
		//TODO: Log the error?
		return err
	}
	//TODO: Log the error?
	return i.service.Start(ctx)
}

// ReportFatalError implements component.Host
func (i *Instance) ReportFatalError(err error) {
	i.logger.Error("fatal error reported", zap.Error(err))
}

// GetFactory implements component.Host
func (i *Instance) GetFactory(kind component.Kind, componentType component.Type) component.Factory {
	switch kind {
	case component.KindReceiver:
		return i.factories.Receivers[componentType]
	default:
		return nil
	}
}

//TODO: Do we need this?
// // GetExtensions implements component.Host
// func (i *Instance) GetExtensions() map[component.ID]extension.Extension {
// 	return i.extensions.GetExtensions()
// }

//TODO: Do we need this?
// // GetExporters implements component.Host
// func (i *Instance) GetExporters() map[component.DataType]map[component.ID]component.Component {
// 	// SpanMetricsProcessor needs to get the configured exporters.
// 	return i.pipelines.GetExporters()
// }
