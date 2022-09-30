package traces

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"go.opencensus.io/stats/view"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config"
	"go.opentelemetry.io/collector/service/extensions"
	"go.opentelemetry.io/collector/service/external/pipelines"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"

	"github.com/grafana/agent/pkg/build"
	"github.com/grafana/agent/pkg/logs"
	"github.com/grafana/agent/pkg/metrics/instance"
	"github.com/grafana/agent/pkg/traces/automaticloggingprocessor"
	"github.com/grafana/agent/pkg/traces/contextkeys"
	"github.com/grafana/agent/pkg/util"
)

// Instance wraps the OpenTelemetry collector to enable tracing pipelines
type Instance struct {
	mut         sync.Mutex
	cfg         InstanceConfig
	logger      *zap.Logger
	metricViews []*view.View

	extensions *extensions.Extensions
	pipelines  *pipelines.Pipelines
	factories  component.Factories
}

var _ component.Host = (*Instance)(nil)

// NewInstance creates and starts an instance of tracing pipelines.
func NewInstance(logsSubsystem *logs.Logs, reg prometheus.Registerer, cfg InstanceConfig, logger *zap.Logger, promInstanceManager instance.Manager) (*Instance, error) {
	var err error

	instance := &Instance{}
	instance.logger = logger
	instance.metricViews, err = newMetricViews(reg)
	if err != nil {
		return nil, fmt.Errorf("failed to create metric views: %w", err)
	}

	if err := instance.ApplyConfig(logsSubsystem, promInstanceManager, reg, cfg); err != nil {
		return nil, err
	}
	return instance, nil
}

// ApplyConfig updates the configuration of the Instance.
func (i *Instance) ApplyConfig(logsSubsystem *logs.Logs, promInstanceManager instance.Manager, reg prometheus.Registerer, cfg InstanceConfig) error {
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
	view.Unregister(i.metricViews...)
}

func (i *Instance) stop() {
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if i.extensions != nil {
		err := i.extensions.NotifyPipelineNotReady()
		if err != nil {
			i.logger.Error("failed to notify extension of pipeline shutdown", zap.Error(err))
		}
	}

	dependencies := []struct {
		name     string
		shutdown func() error
	}{
		{
			name: "pipelines",
			shutdown: func() error {
				if i.pipelines == nil {
					return nil
				}
				return i.pipelines.ShutdownAll(shutdownCtx)
			},
		},
		{
			name: "extensions",
			shutdown: func() error {
				if i.extensions == nil {
					return nil
				}
				return i.extensions.Shutdown(shutdownCtx)
			},
		},
	}

	for _, dep := range dependencies {
		i.logger.Info(fmt.Sprintf("shutting down %s", dep.name))
		if err := dep.shutdown(); err != nil {
			i.logger.Error(fmt.Sprintf("failed to shutdown %s", dep.name), zap.Error(err))
		}
	}

	i.pipelines = nil
	i.extensions = nil
}

func (i *Instance) buildAndStartPipeline(ctx context.Context, cfg InstanceConfig, logs *logs.Logs, instManager instance.Manager, reg prometheus.Registerer) error {
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

	if cfg.ServiceGraphs != nil {
		ctx = context.WithValue(ctx, contextkeys.PrometheusRegisterer, reg)
	}

	factories, err := tracingFactories()
	if err != nil {
		return fmt.Errorf("failed to load tracing factories: %w", err)
	}
	i.factories = factories

	appinfo := component.BuildInfo{
		Command:     "agent",
		Description: "agent",
		Version:     build.Version,
	}

	settings := component.TelemetrySettings{
		Logger:         i.logger,
		TracerProvider: trace.NewNoopTracerProvider(),
		MeterProvider:  metric.NewNoopMeterProvider(),
	}

	// start extensions
	i.extensions, err = extensions.New(ctx, extensions.Settings{
		Telemetry: settings,
		BuildInfo: appinfo,

		Factories: factories.Extensions,
		Configs:   otelConfig.Extensions,
	}, otelConfig.Service.Extensions)
	if err != nil {
		i.logger.Error(fmt.Sprintf("failed to build extensions: %s", err.Error()))
		return fmt.Errorf("failed to create extensions builder: %w", err)
	}
	err = i.extensions.Start(ctx, i)
	if err != nil {
		i.logger.Error(fmt.Sprintf("failed to start extensions: %s", err.Error()))
		return fmt.Errorf("failed to start extensions: %w", err)
	}

	i.pipelines, err = pipelines.Build(ctx, pipelines.Settings{
		Telemetry: settings,
		BuildInfo: appinfo,

		ReceiverFactories:  factories.Receivers,
		ReceiverConfigs:    otelConfig.Receivers,
		ProcessorFactories: factories.Processors,
		ProcessorConfigs:   otelConfig.Processors,
		ExporterFactories:  factories.Exporters,
		ExporterConfigs:    otelConfig.Exporters,

		PipelineConfigs: otelConfig.Pipelines,
	})
	if err != nil {
		return fmt.Errorf("failed to create pipelines: %w", err)
	}
	if err := i.pipelines.StartAll(ctx, i); err != nil {
		i.logger.Error(fmt.Sprintf("failed to start pipelines: %s", err.Error()))
		return fmt.Errorf("failed to start pipelines: %w", err)
	}

	return i.extensions.NotifyPipelineReady()
}

// ReportFatalError implements component.Host
func (i *Instance) ReportFatalError(err error) {
	i.logger.Error("fatal error reported", zap.Error(err))
}

// GetFactory implements component.Host
func (i *Instance) GetFactory(kind component.Kind, componentType config.Type) component.Factory {
	switch kind {
	case component.KindReceiver:
		return i.factories.Receivers[componentType]
	default:
		return nil
	}
}

// GetExtensions implements component.Host
func (i *Instance) GetExtensions() map[config.ComponentID]component.Extension {
	return i.extensions.GetExtensions()
}

// GetExporters implements component.Host
func (i *Instance) GetExporters() map[config.DataType]map[config.ComponentID]component.Exporter {
	// SpanMetricsProcessor needs to get the configured exporters.
	return i.pipelines.GetExporters()
}
