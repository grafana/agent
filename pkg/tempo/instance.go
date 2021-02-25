package tempo

import (
	"context"
	"fmt"
	"time"

	"github.com/grafana/agent/pkg/build"
	"github.com/prometheus/client_golang/prometheus"
	"go.opencensus.io/stats/view"
	"go.uber.org/zap"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config/configmodels"
	"go.opentelemetry.io/collector/service/builder"
)

// Instance wraps the OpenTelemetry collector to enable tracing pipelines
type Instance struct {
	logger      *zap.Logger
	metricViews []*view.View

	exporter  builder.Exporters
	pipelines builder.BuiltPipelines
	receivers builder.Receivers
}

// New creates and starts Loki log collection.
func NewInstance(reg prometheus.Registerer, cfg InstanceConfig, logger *zap.Logger) (*Instance, error) {
	var err error

	instance := &Instance{}
	instance.logger = logger
	instance.metricViews, err = newMetricViews(reg)
	if err != nil {
		return nil, fmt.Errorf("failed to create metric views: %w", err)
	}

	createCtx := context.Background()
	err = instance.buildAndStartPipeline(createCtx, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create exporter: %w", err)
	}

	return instance, nil
}

// Stop stops the OpenTelemetry collector subsystem
func (i *Instance) Stop() {
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := i.receivers.ShutdownAll(shutdownCtx); err != nil {
		i.logger.Error("failed to shutdown receiver", zap.Error(err))
	}

	if err := i.pipelines.ShutdownProcessors(shutdownCtx); err != nil {
		i.logger.Error("failed to shutdown processors", zap.Error(err))
	}

	if err := i.receivers.ShutdownAll(shutdownCtx); err != nil {
		i.logger.Error("failed to shutdown receivers", zap.Error(err))
	}

	view.Unregister(i.metricViews...)
}

func (i *Instance) buildAndStartPipeline(ctx context.Context, cfg InstanceConfig) error {
	// create component factories
	otelConfig, err := cfg.otelConfig()
	if err != nil {
		return fmt.Errorf("failed to load otelConfig from agent tempo config: %w", err)
	}

	factories, err := tracingFactories()
	if err != nil {
		return fmt.Errorf("failed to load tracing factories: %w", err)
	}

	appinfo := component.ApplicationStartInfo{
		ExeName:  "agent",
		GitHash:  build.Revision,
		LongName: "agent",
		Version:  build.Version,
	}

	// start exporter
	i.exporter, err = builder.NewExportersBuilder(i.logger, appinfo, otelConfig, factories.Exporters).Build()
	if err != nil {
		return fmt.Errorf("failed to build exporters: %w", err)
	}

	err = i.exporter.StartAll(ctx, i)
	if err != nil {
		return fmt.Errorf("failed to start exporters: %w", err)
	}

	// start pipelines
	i.pipelines, err = builder.NewPipelinesBuilder(i.logger, appinfo, otelConfig, i.exporter, factories.Processors).Build()
	if err != nil {
		return fmt.Errorf("failed to build exporters: %w", err)
	}

	err = i.pipelines.StartProcessors(ctx, i)
	if err != nil {
		return fmt.Errorf("failed to start processors: %w", err)
	}

	// start receivers
	i.receivers, err = builder.NewReceiversBuilder(i.logger, appinfo, otelConfig, i.pipelines, factories.Receivers).Build()
	if err != nil {
		return fmt.Errorf("failed to start receivers: %w", err)
	}

	err = i.receivers.StartAll(ctx, i)
	if err != nil {
		return fmt.Errorf("failed to start receivers: %w", err)
	}

	return nil
}

// ReportFatalError implements component.Host
func (i *Instance) ReportFatalError(err error) {
	i.logger.Error("fatal error reported", zap.Error(err))
}

// GetFactory implements component.Host
func (i *Instance) GetFactory(kind component.Kind, componentType configmodels.Type) component.Factory {
	return nil
}

// GetExtensions implements component.Host
func (i *Instance) GetExtensions() map[configmodels.Extension]component.ServiceExtension {
	return nil
}

// GetExporters implements component.Host
func (i *Instance) GetExporters() map[configmodels.DataType]map[configmodels.Exporter]component.Exporter {
	return nil
}
