package remotewriteexporter

import (
	"context"

	"github.com/prometheus/prometheus/pkg/labels"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config"
	"go.opentelemetry.io/collector/exporter/exporterhelper"
)

const (
	// TypeStr is the unique identifier for the Prometheus remote write exporter.
	TypeStr = "remote_write"
)

// Config holds the configuration for the Prometheus SD processor.
type Config struct {
	config.ProcessorSettings `mapstructure:",squash"`

	ConstLabels  labels.Labels `mapstructure:"const_labels"`
	Namespace    string        `mapstructure:"namespace"`
	PromInstance string        `mapstructure:"prom_instance"`
}

// NewFactory returns a new factory for the Attributes processor.
func NewFactory() component.ExporterFactory {
	return exporterhelper.NewFactory(
		TypeStr,
		createDefaultConfig,
		exporterhelper.WithMetrics(createMetricsExporter),
	)
}

func createDefaultConfig() config.Exporter {
	return &Config{
		ProcessorSettings: config.NewProcessorSettings(config.NewIDWithName(TypeStr, TypeStr)),
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
