package tempo

import (
	"fmt"
	"os"
	"time"

	"contrib.go.opencensus.io/exporter/prometheus"
	zaplogfmt "github.com/jsternberg/zap-logfmt"
	prom_client "github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
	"github.com/weaveworks/common/logging"
	"go.opencensus.io/stats/view"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"go.opentelemetry.io/collector/config/configtelemetry"
	"go.opentelemetry.io/collector/obsreport"
)

// Tempo wraps the OpenTelemetry collector to enable tracing pipelines
type Tempo struct {
	instances []*Instance
}

// New creates and starts Loki log collection.
func New(reg prom_client.Registerer, cfg Config, level logging.Level) (*Tempo, error) {
	var (
		tempo  Tempo
		logger = newLogger(level)
	)
	for _, c := range cfg.Configs {
		var (
			instLogger = logger.With(zap.String("tempo_config", c.Name))
			instReg    = prom_client.WrapRegistererWith(prom_client.Labels{"tempo_config": c.Name}, reg)
		)

		inst, err := NewInstance(instReg, c, instLogger)
		if err != nil {
			return nil, fmt.Errorf("failed to create tempo instance %s: %w", c.Name, err)
		}
		tempo.instances = append(tempo.instances, inst)
	}

	return &tempo, nil
}

// Stop stops the OpenTelemetry collector subsystem
func (t *Tempo) Stop() {
	for _, i := range t.instances {
		i.Stop()
	}
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

func newMetricViews(reg prom_client.Registerer) ([]*view.View, error) {
	views := obsreport.Configure(configtelemetry.LevelBasic)
	err := view.Register(views...)
	if err != nil {
		return nil, fmt.Errorf("failed to register views: %w", err)
	}

	pe, err := prometheus.NewExporter(prometheus.Options{
		Namespace:  "tempo",
		Registerer: reg,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create prometheus exporter: %w", err)
	}

	view.RegisterExporter(pe)

	return views, nil
}
