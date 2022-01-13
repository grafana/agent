package exporters

import (
	"github.com/grafana/agent/pkg/integrations/v2/app_o11y_exporter/models"
)

// AppReceiverExporter is the base interface
type AppReceiverExporter interface {
	Init() error
}

// AppDataExporter  an interface for exporters
// that are forwarding data to a different service
type AppDataExporter interface {
	AppReceiverExporter
	Export(payload models.Payload) error
}

// AppMetricsExporter is an interface for exporters
// that are exporting metrics into prometheus
type AppMetricsExporter interface {
	AppReceiverExporter
	Process(payload models.Payload) error
}
