package tempo

import (
	"fmt"
	"os"
	"sync"
	"time"

	"contrib.go.opencensus.io/exporter/prometheus"
	"github.com/grafana/agent/pkg/loki"
	"github.com/grafana/agent/pkg/prom/instance"
	zaplogfmt "github.com/jsternberg/zap-logfmt"
	prom_client "github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
	"go.opencensus.io/stats/view"
	"go.opentelemetry.io/collector/external/obsreportconfig"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"go.opentelemetry.io/collector/config/configtelemetry"
)

// Tempo wraps the OpenTelemetry collector to enable tracing pipelines
type Tempo struct {
	mut       sync.Mutex
	instances map[string]*Instance

	leveller *logLeveller
	logger   *zap.Logger
	reg      prom_client.Registerer

	promInstanceManager instance.Manager
}

// New creates and starts Loki log collection.
func New(loki *loki.Loki, promInstanceManager instance.Manager, reg prom_client.Registerer, cfg Config, level logrus.Level) (*Tempo, error) {
	var leveller logLeveller

	tempo := &Tempo{
		instances:           make(map[string]*Instance),
		leveller:            &leveller,
		logger:              newLogger(&leveller),
		reg:                 reg,
		promInstanceManager: promInstanceManager,
	}
	if err := tempo.ApplyConfig(loki, promInstanceManager, cfg, level); err != nil {
		return nil, err
	}
	return tempo, nil
}

// ApplyConfig updates Tempo with a new Config.
func (t *Tempo) ApplyConfig(loki *loki.Loki, promInstanceManager instance.Manager, cfg Config, level logrus.Level) error {
	t.mut.Lock()
	defer t.mut.Unlock()

	// Update the log level, if it has changed.
	t.leveller.SetLevel(level)

	newInstances := make(map[string]*Instance, len(cfg.Configs))

	for _, c := range cfg.Configs {
		// If an old instance exists, update it and move it to the new map.
		if old, ok := t.instances[c.Name]; ok {
			err := old.ApplyConfig(loki, promInstanceManager, c)
			if err != nil {
				return err
			}

			newInstances[c.Name] = old
			continue
		}

		var (
			instLogger = t.logger.With(zap.String("tempo_config", c.Name))
			instReg    = prom_client.WrapRegistererWith(prom_client.Labels{"tempo_config": c.Name}, t.reg)
		)

		inst, err := NewInstance(loki, instReg, c, instLogger, t.promInstanceManager)
		if err != nil {
			return fmt.Errorf("failed to create tempo instance %s: %w", c.Name, err)
		}
		newInstances[c.Name] = inst
	}

	// Any instance in l.instances that isn't in newInstances has been removed
	// from the config. Stop them before replacing the map.
	for key, i := range t.instances {
		if _, exist := newInstances[key]; exist {
			continue
		}
		i.Stop()
	}
	t.instances = newInstances

	return nil
}

// Stop stops the OpenTelemetry collector subsystem
func (t *Tempo) Stop() {
	t.mut.Lock()
	defer t.mut.Unlock()

	for _, i := range t.instances {
		i.Stop()
	}
}

func newLogger(zapLevel zapcore.LevelEnabler) *zap.Logger {
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

// logLeveller implements the zapcore.LevelEnabler interface and allows for
// switching out log levels at runtime.
type logLeveller struct {
	mut   sync.RWMutex
	inner zapcore.Level
}

func (l *logLeveller) SetLevel(level logrus.Level) {
	l.mut.Lock()
	defer l.mut.Unlock()

	zapLevel := zapcore.InfoLevel

	switch level {
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

	l.inner = zapLevel
}

func (l *logLeveller) Enabled(target zapcore.Level) bool {
	l.mut.RLock()
	defer l.mut.RUnlock()
	return l.inner.Enabled(target)
}

func newMetricViews(reg prom_client.Registerer) ([]*view.View, error) {
	obsMetrics := obsreportconfig.Configure(configtelemetry.LevelBasic)
	err := view.Register(obsMetrics.Views...)
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

	return obsMetrics.Views, nil
}
