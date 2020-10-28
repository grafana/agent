package tempo

import (
	"context"
	"fmt"
	"os"
	"time"

	"contrib.go.opencensus.io/exporter/prometheus"
	"github.com/grafana/agent/pkg/build"
	zaplogfmt "github.com/jsternberg/zap-logfmt"
	prom_client "github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
	"github.com/weaveworks/common/logging"
	"go.opencensus.io/stats/view"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config/configmodels"
	"go.opentelemetry.io/collector/obsreport"
	"go.opentelemetry.io/collector/service/builder"
)

// Tempo wraps the OpenTelemetry collector to enablet tracing pipelines
type Tempo struct {
	logger      *zap.Logger
	metricViews []*view.View

	exporter  builder.Exporters
	pipelines builder.BuiltPipelines
	receivers builder.Receivers
}

// New creates and starts Loki log collection.
func New(cfg Config, level logging.Level) (*Tempo, error) {
	var err error

	tempo := &Tempo{}
	tempo.logger = newLogger(level)
	tempo.metricViews, err = newMetricViews()
	if err != nil {
		return nil, fmt.Errorf("failed to create metric views: %w", err)
	}

	createCtx := context.Background()
	err = tempo.buildAndStartPipeline(createCtx, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create exporter: %w", err)
	}

	return tempo, nil
}

// Stop stops the OpenTelemetry collector subsystem
func (t *Tempo) Stop() {
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := t.receivers.ShutdownAll(shutdownCtx); err != nil {
		t.logger.Error("failed to shutdown receiver", zap.Error(err))
	}

	if err := t.pipelines.ShutdownProcessors(shutdownCtx); err != nil {
		t.logger.Error("failed to shutdown processors", zap.Error(err))
	}

	if err := t.receivers.ShutdownAll(shutdownCtx); err != nil {
		t.logger.Error("failed to shutdown receivers", zap.Error(err))
	}

	view.Unregister(t.metricViews...)
}

func (t *Tempo) buildAndStartPipeline(ctx context.Context, cfg Config) error {
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
	t.exporter, err = builder.NewExportersBuilder(t.logger, appinfo, otelConfig, factories.Exporters).Build()
	if err != nil {
		return fmt.Errorf("failed to build exporters: %w", err)
	}

	err = t.exporter.StartAll(ctx, t)
	if err != nil {
		return fmt.Errorf("failed to start exporters: %w", err)
	}

	// start pipelines
	t.pipelines, err = builder.NewPipelinesBuilder(t.logger, appinfo, otelConfig, t.exporter, factories.Processors).Build()
	if err != nil {
		return fmt.Errorf("failed to build exporters: %w", err)
	}

	err = t.pipelines.StartProcessors(ctx, t)
	if err != nil {
		return fmt.Errorf("failed to start processors: %w", err)
	}

	// start receivers
	t.receivers, err = builder.NewReceiversBuilder(t.logger, appinfo, otelConfig, t.pipelines, factories.Receivers).Build()
	if err != nil {
		return fmt.Errorf("failed to start receivers: %w", err)
	}

	err = t.receivers.StartAll(ctx, t)
	if err != nil {
		return fmt.Errorf("failed to start receivers: %w", err)
	}

	return nil
}

// ReportFatalError implements component.Host
func (t *Tempo) ReportFatalError(err error) {
	t.logger.Error("fatal error reported", zap.Error(err))
}

// GetFactory implements component.Host
func (t *Tempo) GetFactory(kind component.Kind, componentType configmodels.Type) component.Factory {
	return nil
}

// GetExtensions implements component.Host
func (t *Tempo) GetExtensions() map[configmodels.Extension]component.ServiceExtension {
	return nil
}

// GetExporters implements component.Host
func (t *Tempo) GetExporters() map[configmodels.DataType]map[configmodels.Exporter]component.Exporter {
	return nil
}

func newLogger(level logging.Level) *zap.Logger {
	zapLevel := zapcore.InfoLevel

	switch level.Logrus {
	case logrus.PanicLevel:
		zapLevel = zapcore.PanicLevel
	case logrus.FatalLevel:
		zapLevel = zapcore.FatalLevel
	case logrus.ErrorLevel:
		zapLevel = zapcore.ErrorLevel
	case logrus.WarnLevel:
		zapLevel = zapcore.WarnLevel
	case logrus.InfoLevel:
		zapLevel = zapcore.InfoLevel
	case logrus.DebugLevel:
	case logrus.TraceLevel:
		zapLevel = zapcore.DebugLevel
	}

	config := zap.NewProductionEncoderConfig()
	config.EncodeTime = func(ts time.Time, encoder zapcore.PrimitiveArrayEncoder) {
		encoder.AppendString(ts.UTC().Format(time.RFC3339))
	}
	logger := zap.New(zapcore.NewCore(
		zaplogfmt.NewEncoder(config),
		os.Stdout,
		zapLevel,
	))
	logger = logger.With(zap.String("component", "tempo"))
	logger.Info("Tempo Logger Initialized")

	return logger
}

func newMetricViews() ([]*view.View, error) {
	views := obsreport.Configure(false, true)
	err := view.Register(views...)
	if err != nil {
		return nil, fmt.Errorf("failed to register views: %w", err)
	}

	pe, err := prometheus.NewExporter(prometheus.Options{
		Namespace:  "tempo",
		Registerer: prom_client.DefaultRegisterer,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create prometheus exporter: %w", err)
	}

	view.RegisterExporter(pe)

	return views, nil
}
