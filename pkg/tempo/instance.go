package tempo

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/grafana/agent/pkg/build"
	"github.com/grafana/agent/pkg/util"
	"github.com/prometheus/client_golang/prometheus"
	"go.opencensus.io/stats/view"
	"go.uber.org/zap"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config/configmodels"
	"go.opentelemetry.io/collector/service/builder"
)

// Instance wraps the OpenTelemetry collector to enable tracing pipelines
type Instance struct {
	mut         sync.Mutex
	cfg         InstanceConfig
	logger      *zap.Logger
	metricViews []*view.View

	exporter  builder.Exporters
	pipelines builder.BuiltPipelines
	receivers builder.Receivers
}

// NewInstance creates and starts an instance of tracing pipelines.
func NewInstance(reg prometheus.Registerer, cfg InstanceConfig, logger *zap.Logger) (*Instance, error) {
	var err error

	instance := &Instance{}
	instance.logger = logger
	instance.metricViews, err = newMetricViews(reg)
	if err != nil {
		return nil, fmt.Errorf("failed to create metric views: %w", err)
	}

	if err := instance.ApplyConfig(cfg); err != nil {
		return nil, err
	}
	return instance, nil
}

// ApplyConfig updates the configuration of the Instance.
func (i *Instance) ApplyConfig(cfg InstanceConfig) error {
	i.mut.Lock()
	defer i.mut.Unlock()

	if util.CompareYAML(cfg, i.cfg) {
		// No config change
		return nil
	}
	i.cfg = cfg

	// Shut down any existing pipeline
	i.stop()

	createCtx := context.Background()
	err := i.buildAndStartPipeline(createCtx, cfg)
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

	dependencies := []struct {
		name     string
		shutdown func() error
	}{
		{
			name: "receiver",
			shutdown: func() error {
				if i.receivers == nil {
					return nil
				}
				return i.receivers.ShutdownAll(shutdownCtx)
			},
		},
		{
			name: "processors",
			shutdown: func() error {
				if i.pipelines == nil {
					return nil
				}
				return i.pipelines.ShutdownProcessors(shutdownCtx)
			},
		},
		{
			name: "exporters",
			shutdown: func() error {
				if i.exporter == nil {
					return nil
				}
				return i.exporter.ShutdownAll(shutdownCtx)
			},
		},
	}

	for _, dep := range dependencies {
		i.logger.Info(fmt.Sprintf("shutting down %s", dep.name))
		if err := dep.shutdown(); err != nil {
			i.logger.Error(fmt.Sprintf("failed to shutdown %s", dep.name), zap.Error(err))
		}
	}

	i.receivers = nil
	i.pipelines = nil
	i.exporter = nil
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
		return fmt.Errorf("failed to create exporters builder: %w", err)
	}

	err = i.exporter.StartAll(ctx, i)
	if err != nil {
		return fmt.Errorf("failed to start exporters: %w", err)
	}

	// start pipelines
	i.pipelines, err = builder.NewPipelinesBuilder(i.logger, appinfo, otelConfig, i.exporter, factories.Processors).Build()
	if err != nil {
		return fmt.Errorf("failed to create pipelines builder: %w", err)
	}

	err = i.pipelines.StartProcessors(ctx, i)
	if err != nil {
		return fmt.Errorf("failed to start processors: %w", err)
	}

	// start receivers
	i.receivers, err = builder.NewReceiversBuilder(i.logger, appinfo, otelConfig, i.pipelines, factories.Receivers).Build()
	if err != nil {
		return fmt.Errorf("failed to create receivers builder: %w", err)
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
