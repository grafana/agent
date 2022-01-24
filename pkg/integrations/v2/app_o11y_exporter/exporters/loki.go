package exporters

import (
	"fmt"
	"time"

	kitlog "github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/go-logfmt/logfmt"
	"github.com/grafana/agent/pkg/integrations/v2/app_o11y_exporter/models"
	"github.com/grafana/agent/pkg/integrations/v2/app_o11y_exporter/utils"
	loki "github.com/grafana/agent/pkg/logs"
	"github.com/grafana/loki/clients/pkg/promtail/api"
	"github.com/grafana/loki/pkg/logproto"
	prommodel "github.com/prometheus/common/model"
)

// LokiExporterConfig holds the configuration of the loki exporter
type LokiExporterConfig struct {
	SendEntryTimeout int
	LokiInstance     *loki.Instance
	ExtraLabels      map[string]string
}

// LokiExporter is the struct of the loki exporter
type LokiExporter struct {
	li          *loki.Instance
	seTimeout   time.Duration
	logger      kitlog.Logger
	extraLabels prommodel.LabelSet
}

// Init implements the AppReceiverExporter interface
func (le *LokiExporter) Init() error {
	return nil
}

// NewLokiExporter creates a new Loki loki exporter with the given
// configuration
func NewLokiExporter(logger kitlog.Logger, conf LokiExporterConfig) AppReceiverExporter {
	el := make(prommodel.LabelSet)
	for k, v := range conf.ExtraLabels {
		el[prommodel.LabelName(k)] = prommodel.LabelValue(v)
	}

	return &LokiExporter{
		logger:      logger,
		li:          conf.LokiInstance,
		seTimeout:   time.Duration(conf.SendEntryTimeout),
		extraLabels: el,
	}
}

// Name of the exporter, for logging purposes
func (le *LokiExporter) Name() string {
	return "loki exporter"
}

// Export implements the AppDataExporter interface
func (le *LokiExporter) Export(payload models.Payload) error {
	meta := payload.Meta.KeyVal()

	// log events
	for _, logItem := range payload.Logs {
		kv := logItem.KeyVal()
		utils.MergeKeyVal(kv, meta)
		le.sendKeyValsTolokiPipeline(kv)
	}
	// exceptions
	for _, exception := range payload.Exceptions {
		kv := exception.KeyVal()
		utils.MergeKeyVal(kv, meta)
		le.sendKeyValsTolokiPipeline(kv)
	}

	// measurements
	for _, measurement := range payload.Measurements {
		kv := measurement.KeyVal()
		utils.MergeKeyVal(kv, meta)
		le.sendKeyValsTolokiPipeline(kv)
	}

	return nil
}

func (le *LokiExporter) sendKeyValsTolokiPipeline(kv *utils.KeyVal) {
	line, err := logfmt.MarshalKeyvals(utils.KeyValToInterfaceSlice(kv)...)
	if err != nil {
		level.Error(le.logger).Log("msg", "failed to logfmt a frontend log event", "err", err)
		return
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
	}
}

func (le *LokiExporter) labelSet(kv *utils.KeyVal) prommodel.LabelSet {
	set := make(prommodel.LabelSet)

	for k, v := range le.extraLabels {
		if len(v) > 0 {
			set[k] = v
		} else {
			if val, ok := kv.Get(k); ok {
				set[k] = prommodel.LabelValue(fmt.Sprint(val))
			}
		}
	}

	return set
}

// Static typecheck tests
var (
	_ AppReceiverExporter = (*LokiExporter)(nil)
)
