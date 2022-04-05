package exporters

import (
	"fmt"
	"time"

	kitlog "github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/go-logfmt/logfmt"
	"github.com/grafana/agent/pkg/integrations/v2/app_o11y_receiver/models"
	"github.com/grafana/agent/pkg/integrations/v2/app_o11y_receiver/sourcemaps"
	"github.com/grafana/agent/pkg/integrations/v2/app_o11y_receiver/utils"
	"github.com/grafana/agent/pkg/logs"
	"github.com/grafana/loki/clients/pkg/promtail/api"
	"github.com/grafana/loki/pkg/logproto"
	prommodel "github.com/prometheus/common/model"
)

type lokiInstance interface {
	SendEntry(entry api.Entry, dur time.Duration) bool
}

// LokiExporterConfig holds the configuration of the loki exporter
type LokiExporterConfig struct {
	SendEntryTimeout int
	LokiInstance     lokiInstance
	Labels           map[string]string
}

// LokiExporter is the struct of the loki exporter
type LokiExporter struct {
	li             lokiInstance
	seTimeout      time.Duration
	logger         kitlog.Logger
	labels         map[string]string
	sourceMapStore sourcemaps.SourceMapStore
}

// NewLokiExporter creates a new Loki loki exporter with the given
// configuration
func NewLokiExporter(logger kitlog.Logger, conf LokiExporterConfig, sourceMapStore sourcemaps.SourceMapStore) AppO11yReceiverExporter {
	return &LokiExporter{
		logger:         logger,
		li:             conf.LokiInstance,
		seTimeout:      time.Duration(conf.SendEntryTimeout),
		labels:         conf.Labels,
		sourceMapStore: sourceMapStore,
	}
}

// Name of the exporter, for logging purposes
func (le *LokiExporter) Name() string {
	return "loki exporter"
}

// Export implements the AppDataExporter interface
func (le *LokiExporter) Export(payload models.Payload) error {
	meta := payload.Meta.KeyVal()

	var err error

	// log events
	for _, logItem := range payload.Logs {
		kv := logItem.KeyVal()
		utils.MergeKeyVal(kv, meta)
		err = le.sendKeyValsToLogsPipeline(kv)
	}

	// exceptions
	for _, exception := range payload.Exceptions {
		transformedException := le.sourceMapStore.TransformException(&exception, payload.Meta.App.Release)
		kv := transformedException.KeyVal()
		utils.MergeKeyVal(kv, meta)
		err = le.sendKeyValsToLogsPipeline(kv)
	}

	// measurements
	for _, measurement := range payload.Measurements {
		kv := measurement.KeyVal()
		utils.MergeKeyVal(kv, meta)
		err = le.sendKeyValsToLogsPipeline(kv)
	}

	return err
}

func (le *LokiExporter) sendKeyValsToLogsPipeline(kv *utils.KeyVal) error {
	line, err := logfmt.MarshalKeyvals(utils.KeyValToInterfaceSlice(kv)...)
	if err != nil {
		level.Error(le.logger).Log("msg", "failed to logfmt a frontend log event", "err", err)
		return err
	}
	sent := le.li.SendEntry(api.Entry{
		Labels: le.labelSet(kv),
		Entry: logproto.Entry{
			Timestamp: time.Now(),
			Line:      string(line),
		},
	}, le.seTimeout)
	if !sent {
		level.Warn(le.logger).Log("msg", "failed to log frontend log event to loki pipeline")
		return fmt.Errorf("Failed to send app o11y event to logs pipeline")
	}
	return nil
}

func (le *LokiExporter) labelSet(kv *utils.KeyVal) prommodel.LabelSet {
	set := make(prommodel.LabelSet)

	for k, v := range le.labels {
		if len(v) > 0 {
			set[prommodel.LabelName(k)] = prommodel.LabelValue(v)
		} else {
			if val, ok := kv.Get(k); ok {
				set[prommodel.LabelName(k)] = prommodel.LabelValue(fmt.Sprint(val))
			}
		}
	}

	return set
}

// Static typecheck tests
var (
	_ AppO11yReceiverExporter = (*LokiExporter)(nil)
	_ lokiInstance            = (*logs.Instance)(nil)
)
