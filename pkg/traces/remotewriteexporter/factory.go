package remotewriteexporter

import (
	"context"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config"
	"go.opentelemetry.io/collector/exporter/exporterhelper"
)

const (
	// TypeStr is the unique identifier for the Prometheus remote write exporter.
	TypeStr = "remote_write"
)

type label struct {
	Name  string `mapstructure:"name"`
	Value string `mapstructure:"name"`
}

var _ config.Exporter = (*Config)(nil)

// Config holds the configuration for the Prometheus remote write processor.
type Config struct {
	config.ExporterSettings `mapstructure:",squash"`

	ConstLabels  []label `mapstructure:"const_labels"`
	Namespace    string  `mapstructure:"namespace"`
	PromInstance string  `mapstructure:"metrics_instance"`
}

// NewFactory returns a new factory for the Prometheus remote write processor.
func NewFactory() component.ExporterFactory {
	return exporterhelper.NewFactory(
		TypeStr,
		createDefaultConfig,
		exporterhelper.WithMetrics(createMetricsExporter),
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
