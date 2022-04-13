package exporters

import (
	"context"
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

// LogsInstance is an interface with capability to send log entries
type LogsInstance interface {
	SendEntry(entry api.Entry, dur time.Duration) bool
}

// LogsInstanceGetter is a function that returns a LogsInstance to send log entries to
type LogsInstanceGetter func() LogsInstance

// LogsExporterConfig holds the configuration of the logs exporter
type LogsExporterConfig struct {
	SendEntryTimeout int
	GetLogsInstance  LogsInstanceGetter
	Labels           map[string]string
}

// LogsExporter will send logs & errors to loki
type LogsExporter struct {
	getLogsInstance LogsInstanceGetter
	seTimeout       time.Duration
	logger          kitlog.Logger
	labels          map[string]string
	sourceMapStore  sourcemaps.SourceMapStore
}

// NewLogsExporter creates a new logs exporter with the given
// configuration
func NewLogsExporter(logger kitlog.Logger, conf LogsExporterConfig, sourceMapStore sourcemaps.SourceMapStore) AppO11yReceiverExporter {
	return &LogsExporter{
		logger:          logger,
		getLogsInstance: conf.GetLogsInstance,
		seTimeout:       time.Duration(conf.SendEntryTimeout),
		labels:          conf.Labels,
		sourceMapStore:  sourceMapStore,
	}
}

// Name of the exporter, for logging purposes
func (le *LogsExporter) Name() string {
	return "logs exporter"
}

// Export implements the AppDataExporter interface
func (le *LogsExporter) Export(ctx context.Context, payload models.Payload) error {
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

func (le *LogsExporter) sendKeyValsToLogsPipeline(kv *utils.KeyVal) error {
	line, err := logfmt.MarshalKeyvals(utils.KeyValToInterfaceSlice(kv)...)
	if err != nil {
		level.Error(le.logger).Log("msg", "failed to logfmt a frontend log event", "err", err)
		return err
	}
	instance := le.getLogsInstance()
	if instance == nil {
		return fmt.Errorf("failed to get logs instance")
	}
	sent := instance.SendEntry(api.Entry{
		Labels: le.labelSet(kv),
		Entry: logproto.Entry{
			Timestamp: time.Now(),
			Line:      string(line),
		},
	}, le.seTimeout)
	if !sent {
		level.Warn(le.logger).Log("msg", "failed to log frontend log event to logs pipeline")
		return fmt.Errorf("failed to send app o11y event to logs pipeline")
	}
	return nil
}

func (le *LogsExporter) labelSet(kv *utils.KeyVal) prommodel.LabelSet {
	set := make(prommodel.LabelSet, len(le.labels))

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
	_ AppO11yReceiverExporter = (*LogsExporter)(nil)
	_ LogsInstance            = (*logs.Instance)(nil)
)
