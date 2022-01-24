package exporters

import (
	"github.com/grafana/agent/pkg/integrations/v2/app_o11y_exporter/models"
)

// AppReceiverExporter is the base interface
type AppReceiverExporter interface {
	Name() string
	Init() error
	Export(payload models.Payload) error
}
