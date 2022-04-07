package exporters

import (
	"context"

	"github.com/grafana/agent/pkg/integrations/v2/app_o11y_receiver/models"
)

// AppO11yReceiverExporter is the base interface
type AppO11yReceiverExporter interface {
	Name() string
	Export(ctx context.Context, payload models.Payload) error
}
