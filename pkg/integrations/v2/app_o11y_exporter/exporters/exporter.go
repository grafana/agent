package exporters

import (
	"github.com/grafana/agent/pkg/integrations/v2/app_o11y_exporter/models"
	"github.com/prometheus/common/model"
)

type AppReceiverExporter interface {
	Init() error
}

type AppDataExporter interface {
	AppReceiverExporter
	Export(payload models.Payload) error
}

type AppMetricTarget struct {
	MetricsPath string
	Labels      model.LabelSet
}

type AppMetricsExporter interface {
	AppReceiverExporter
	Process(payload models.Payload) error
}
