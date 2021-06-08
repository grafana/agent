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

type label struct {
	Name  string `mapstructure:"name"`
	Value string `mapstructure:"value"`
}

type labelsHelper []label

func (lh labelsHelper) AsLabels() labels.Labels {
	ls := make(labels.Labels, 0, len(lh))
	for _, l := range lh {
		ls = append(ls, labels.Label{
			Name:  l.Name,
			Value: l.Value,
		})
	}
	return ls
}

// Config holds the configuration for the Prometheus SD processor.
type Config struct {
	config.ProcessorSettings `mapstructure:",squash"`

	ConstLabels  labelsHelper `mapstructure:"const_labels,omitempty"`
	Namespace    string       `mapstructure:"namespace"`
	PromInstance string       `mapstructure:"prom_instance"`
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
	processorSettings := config.NewProcessorSettings(config.NewIDWithName(TypeStr, TypeStr))
	return &processorSettings
}

func createMetricsExporter(
	_ context.Context,
	_ component.ExporterCreateParams,
	cfg config.Exporter,
) (component.MetricsExporter, error) {
	eCfg := cfg.(*Config)

	return newRemoteWriteExporter(eCfg)
}
