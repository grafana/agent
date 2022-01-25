package exporters

import (
	"github.com/grafana/agent/pkg/integrations/v2/app_o11y_receiver/models"
)

// AppO11yReceiverExporter is the base interface
type AppO11yReceiverExporter interface {
	Name() string
	Init() error
	Export(payload models.Payload) error
}
