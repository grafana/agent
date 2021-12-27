package exporters

import (
	"errors"
	"time"

	"github.com/grafana/agent/pkg/integrations/v2/app_o11y_exporter/models"
	"github.com/grafana/agent/pkg/logs"
	"github.com/grafana/loki/clients/pkg/promtail/api"
	"github.com/grafana/loki/pkg/logproto"
)

type LokiExporterConfig struct {
	SendEntryTimeout int
}

type LokiExporter struct {
	li   *logs.Instance
	conf LokiExporterConfig
}

func (le *LokiExporter) Init() error {
	return nil
}

func NewLokiExporter(lokiInstance *logs.Instance, conf LokiExporterConfig) AppReceiverExporter {
	return &LokiExporter{li: lokiInstance, conf: conf}
}

func (le *LokiExporter) Export(payload models.Payload) error {
	for _, log := range payload.Logs {
		e := api.Entry{
			Labels: log.LabelSet(),
			Entry: logproto.Entry{
				Timestamp: log.Timestamp,
				Line:      log.Message,
			},
		}
		if !le.li.SendEntry(e, time.Duration(le.conf.SendEntryTimeout)) {
			return errors.New("Error while sending log over to Loki")
		}
	}
	return nil
}

// Static typecheck tests
var (
	_ AppReceiverExporter = (*LokiExporter)(nil)
	_ AppDataExporter     = (*LokiExporter)(nil)
)
