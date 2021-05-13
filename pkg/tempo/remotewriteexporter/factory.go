package remotewriteexporter

import (
	"context"

	"github.com/prometheus/prometheus/pkg/labels"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config/configmodels"
	"go.opentelemetry.io/collector/exporter/exporterhelper"
)

const (
	// TypeStr is the unique identifier for the Prometheus remote write exporter.
	TypeStr = "remote_write"
)

// Config holds the configuration for the Prometheus SD processor.
type Config struct {
	configmodels.ProcessorSettings `mapstructure:",squash"`

	ConstLabels labels.Labels `mapstructure:"const_labels"`
	Namespace   string        `mapstructure:"namespace"`
}

// NewFactory returns a new factory for the Attributes processor.
func NewFactory() component.ExporterFactory {
	return exporterhelper.NewFactory(
		TypeStr,
		createDefaultConfig,
		exporterhelper.WithMetrics(createMetricsExporter),
	)
}

func createDefaultConfig() configmodels.Exporter {
	return &Config{
		ProcessorSettings: configmodels.ProcessorSettings{
			TypeVal: TypeStr,
			NameVal: TypeStr,
		},
	}
}

func createMetricsExporter(
	_ context.Context,
	_ component.ExporterCreateParams,
	cfg configmodels.Exporter,
) (component.MetricsExporter, error) {
	eCfg := cfg.(*Config)

	return newRemoteWriteExporter(eCfg)
}
