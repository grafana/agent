package exporters

import (
	"errors"
	"time"

	"github.com/grafana/agent/pkg/integrations/v2/app_o11y_exporter/models"
	"github.com/grafana/agent/pkg/logs"
	"github.com/grafana/loki/clients/pkg/promtail/api"
	"github.com/grafana/loki/pkg/logproto"
)

// LokiExceptionExporter hold the information for reporting exceptions to Loki
type LokiExceptionExporter struct {
	li *logs.Instance
}

// NewLokiExceptionExporter creates a new LokiExeptionExporter for a specific
// Loki instance
func NewLokiExceptionExporter(lokiInstance *logs.Instance) (AppReceiverExporter, error) {
	return &LokiExceptionExporter{li: lokiInstance}, nil
}

// Init implements the AppReceiverExporter interface
func (le *LokiExceptionExporter) Init() error {
	return nil
}

// Export implements the AppDataExporter interface
func (le *LokiExceptionExporter) Export(payload models.Payload) error {
	for _, exception := range payload.Exceptions {
		e := api.Entry{
			Labels: exception.LabelSet(),
			Entry: logproto.Entry{
				Timestamp: exception.Timestamp,
				Line:      exception.String(),
			},
		}
		if !le.li.SendEntry(e, time.Duration(1000)) {
			return errors.New("Error while sending log over to Loki")
		}
	}
	return nil
}

// Static typecheck tests
var (
	_ AppReceiverExporter = (*LokiExceptionExporter)(nil)
	_ AppDataExporter     = (*LokiExceptionExporter)(nil)
)
