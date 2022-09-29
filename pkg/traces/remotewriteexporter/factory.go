package remotewriteexporter

import (
	"context"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config"
)

const (
	// TypeStr is the unique identifier for the Prometheus remote write exporter.
	// TODO: Rename to walexporter (?). Remote write makes no sense, it appends to a WAL.
	TypeStr = "remote_write"
)

var _ config.Exporter = (*Config)(nil)

// Config holds the configuration for the Prometheus remote write processor.
type Config struct {
	config.ExporterSettings `mapstructure:",squash"`

	ConstLabels  prometheus.Labels `mapstructure:"const_labels"`
	Namespace    string            `mapstructure:"namespace"`
	PromInstance string            `mapstructure:"metrics_instance"`
	// StaleTime is the duration after which a series is considered stale and will be removed.
	StaleTime time.Duration `mapstructure:"stale_time"`
	// LoopInterval is the duration after which the exporter will be checked for new data.
	// New data is flushed to a WAL.
	LoopInterval time.Duration `mapstructure:"loop_interval"`
}

// NewFactory returns a new factory for the Prometheus remote write processor.
func NewFactory() component.ExporterFactory {
	return component.NewExporterFactory(
		TypeStr,
		createDefaultConfig,
		component.WithMetricsExporter(createMetricsExporter, component.StabilityLevelUndefined),
	)
}

func createDefaultConfig() config.Exporter {
	return &Config{
		ExporterSettings: config.NewExporterSettings(config.NewComponentIDWithName(TypeStr, TypeStr)),
	}
}

func createMetricsExporter(
	_ context.Context,
	_ component.ExporterCreateSettings,
	cfg config.Exporter,
) (component.MetricsExporter, error) {

	eCfg := cfg.(*Config)
	return newRemoteWriteExporter(eCfg)
}
