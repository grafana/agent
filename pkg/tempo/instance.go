package tempo

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/grafana/agent/pkg/build"
	"github.com/grafana/agent/pkg/loki"
	"github.com/grafana/agent/pkg/prom/instance"
	"github.com/grafana/agent/pkg/tempo/contextkeys"
	"github.com/grafana/agent/pkg/util"
	"github.com/prometheus/client_golang/prometheus"
	"go.opencensus.io/stats/view"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config/configmodels"
	"go.opentelemetry.io/collector/service"
	"go.uber.org/zap"
)

// Instance wraps the OpenTelemetry collector to enable tracing pipelines
type Instance struct {
	mut         sync.Mutex
	cfg         InstanceConfig
	logger      *zap.Logger
	metricViews []*view.View

	app *service.Application
}

// NewInstance creates and starts an instance of tracing pipelines.
func NewInstance(loki *loki.Loki, reg prometheus.Registerer, cfg InstanceConfig, logger *zap.Logger, promInstanceManager instance.Manager) (*Instance, error) {
	var err error

	instance := &Instance{}
	instance.logger = logger
	instance.metricViews, err = newMetricViews(reg)
	if err != nil {
		return nil, fmt.Errorf("failed to create metric views: %w", err)
	}

	if err := instance.ApplyConfig(loki, promInstanceManager, cfg); err != nil {
		return nil, err
	}
	return instance, nil
}

// ApplyConfig updates the configuration of the Instance.
func (i *Instance) ApplyConfig(loki *loki.Loki, promInstanceManager instance.Manager, cfg InstanceConfig) error {
	i.mut.Lock()
	defer i.mut.Unlock()

	if util.CompareYAML(cfg, i.cfg) {
		// No config change
		return nil
	}
	i.cfg = cfg

	// Shut down any existing pipeline
	i.stop()

	createCtx := context.WithValue(context.Background(), contextkeys.Loki, loki)
	err := i.buildAndStartPipeline(createCtx, cfg, promInstanceManager)
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
	if i.app == nil {
		return
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	i.app.Shutdown()
	for {
		select {
		case state := <-i.app.GetStateChannel():
			if state == service.Closed {
				i.app = nil
				return
			}
		case <-shutdownCtx.Done():
			return
		}
	}
}

func (i *Instance) buildAndStartPipeline(ctx context.Context, cfg InstanceConfig, promManager instance.Manager) error {
	if cfg.PushConfig.Endpoint != "" {
		i.logger.Warn("Configuring exporter with deprecated push_config. Use remote_write and batch instead")
	}

	if cfg.SpanMetrics != nil && len(cfg.SpanMetrics.PromInstance) != 0 {
		ctx = context.WithValue(ctx, contextkeys.Prometheus, promManager)
	}

	factories, err := tracingFactories()
	if err != nil {
		return fmt.Errorf("failed to load tracing factories: %w", err)
	}

	startInfo := component.ApplicationStartInfo{
		ExeName:  "agent",
		GitHash:  build.Revision,
		LongName: "agent",
		Version:  build.Version,
	}

	params := service.Parameters{
		Factories:            factories,
		ApplicationStartInfo: startInfo,
		ConfigFactory:        cfg.configFactory,
		LoggingOptions:       []zap.Option{zap.Development()},
	}

	app, err := service.New(params)
	if err != nil {
		return fmt.Errorf("failed creating tracing application: %s", err)
	}
	i.app = app

	return i.start(ctx)
}

func (i *Instance) start(ctx context.Context) error {
	errChan := make(chan error)
	go func() {
		cmd := i.app.Command()
		cmd.SetArgs([]string{"--metrics-level=none"})
		cmd.SilenceUsage = true

		err := cmd.ExecuteContext(ctx)
		if err != nil {
			errChan <- err
		}
	}()

	for {
		select {
		case s := <-i.app.GetStateChannel():
			if s == service.Running {
				return nil
			}
		case err := <-errChan:
			return fmt.Errorf("failed to start tracing application: %s", err)
		case <-ctx.Done():
			return fmt.Errorf("failed to start tracing application: timeout")
		}
	}
}

// ReportFatalError implements component.Host
func (i *Instance) ReportFatalError(err error) {
	i.logger.Error("fatal error reported", zap.Error(err))
}

// GetFactory implements component.Host
func (i *Instance) GetFactory(_ component.Kind, _ configmodels.Type) component.Factory {
	return nil
}

// GetExtensions implements component.Host
func (i *Instance) GetExtensions() map[configmodels.Extension]component.Extension {
	return nil
}

// GetExporters implements component.Host
func (i *Instance) GetExporters() map[configmodels.DataType]map[configmodels.NamedEntity]component.Exporter {
	// SpanMetricsProcessor needs to get the configured exporters.
	return i.app.GetExporters()
}
