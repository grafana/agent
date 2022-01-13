package exporters

import (
	"errors"
	"time"

	"github.com/go-sourcemap/sourcemap"
	"github.com/grafana/agent/pkg/integrations/v2/app_o11y_exporter/config"
	"github.com/grafana/agent/pkg/integrations/v2/app_o11y_exporter/models"
	"github.com/grafana/agent/pkg/integrations/v2/app_o11y_exporter/tools/sourcemaps"
	"github.com/grafana/agent/pkg/logs"
	"github.com/grafana/loki/clients/pkg/promtail/api"
	"github.com/grafana/loki/pkg/logproto"
)

// LokiExceptionExporter hold the information for reporting exceptions to Loki
// as well as tranforming stacktraces with sourcemaps
type LokiExceptionExporter struct {
	li  *logs.Instance
	smc *sourcemap.Consumer
}

// NewLokiExceptionExporter creates a new LokiExeptionExporter for a specific
// Loki instance
func NewLokiExceptionExporter(lokiInstance *logs.Instance, conf config.SourceMapConfig) (AppReceiverExporter, error) {
	var (
		smc *sourcemap.Consumer
		err error
	)
	if conf.Enabled {
		smc, err = createSourceMapConsumer(conf)
		if err != nil {
			return nil, err
		}
	}
	return &LokiExceptionExporter{li: lokiInstance, smc: smc}, nil
}

func createSourceMapConsumer(conf config.SourceMapConfig) (scm *sourcemap.Consumer, err error) {
	loader, err := sourcemaps.NewMapLoader(conf)

	if err != nil {
		return nil, err
	}

	scm, err = loader.Load(conf)
	if err != nil {
		return nil, err
	}

	return scm, nil
}

// Init implements the AppReceiverExporter interface
func (le *LokiExceptionExporter) Init() error {
	return nil
}

// Export implements the AppDataExporter interface
func (le *LokiExceptionExporter) Export(payload models.Payload) error {
	for _, exception := range payload.Exceptions {
		if le.smc != nil {
			mappedStacktrace := exception.Stacktrace.MapFrames(le.smc)
			exception.Stacktrace = &mappedStacktrace
		}
		e := api.Entry{
			Labels: exception.LabelSet(),
			Entry: logproto.Entry{
				Timestamp: exception.Timestamp,
				Line:      exception.String(),
			},
		}
		if le.li.SendEntry(e, time.Duration(1000)) {
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
