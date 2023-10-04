package receiver

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/go-logfmt/logfmt"
	"github.com/grafana/agent/component/common/loki"
	"github.com/grafana/agent/component/faro/receiver/internal/payload"
	"github.com/grafana/agent/component/otelcol"
	"github.com/grafana/loki/pkg/logproto"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/model"
)

type exporter interface {
	Name() string
	Export(ctx context.Context, payload payload.Payload) error
}

//
// Metrics
//

type metricsExporter struct {
	totalLogs         prometheus.Counter
	totalMeasurements prometheus.Counter
	totalExceptions   prometheus.Counter
	totalEvents       prometheus.Counter
}

var _ exporter = (*metricsExporter)(nil)

func newMetricsExporter(reg prometheus.Registerer) *metricsExporter {
	exp := &metricsExporter{
		totalLogs: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "faro_receiver_logs_total",
			Help: "Total number of ingested logs",
		}),
		totalMeasurements: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "faro_receiver_measurements_total",
			Help: "Total number of ingested measurements",
		}),
		totalExceptions: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "faro_receiver_exceptions_total",
			Help: "Total number of ingested exceptions",
		}),
		totalEvents: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "faro_receiver_events_total",
			Help: "Total number of ingested events",
		}),
	}

	reg.MustRegister(exp.totalLogs, exp.totalExceptions, exp.totalMeasurements, exp.totalEvents)

	return exp
}

func (exp *metricsExporter) Name() string { return "receiver metrics exporter" }

func (exp *metricsExporter) Export(ctx context.Context, p payload.Payload) error {
	exp.totalExceptions.Add(float64(len(p.Exceptions)))
	exp.totalLogs.Add(float64(len(p.Logs)))
	exp.totalMeasurements.Add(float64(len(p.Measurements)))
	exp.totalEvents.Add(float64(len(p.Events)))
	return nil
}

//
// Logs
//

type logsExporter struct {
	log        log.Logger
	sourceMaps sourceMapsStore

	receiversMut sync.RWMutex
	receivers    []loki.LogsReceiver

	labelsMut sync.RWMutex
	labels    model.LabelSet
}

var _ exporter = (*logsExporter)(nil)

func newLogsExporter(log log.Logger, sourceMaps sourceMapsStore) *logsExporter {
	return &logsExporter{
		log:        log,
		sourceMaps: sourceMaps,
	}
}

// SetReceivers updates the set of logs receivers which will receive logs
// emitted by the exporter.
func (exp *logsExporter) SetReceivers(receivers []loki.LogsReceiver) {
	exp.receiversMut.Lock()
	defer exp.receiversMut.Unlock()

	exp.receivers = receivers
}

func (exp *logsExporter) Name() string { return "logs exporter" }

func (exp *logsExporter) Export(ctx context.Context, p payload.Payload) error {
	meta := p.Meta.KeyVal()

	var errs []error

	// log events
	for _, logItem := range p.Logs {
		kv := logItem.KeyVal()
		payload.MergeKeyVal(kv, meta)
		errs = append(errs, exp.sendKeyValsToLogsPipeline(ctx, kv))
	}

	// exceptions
	for _, exception := range p.Exceptions {
		transformedException := transformException(exp.log, exp.sourceMaps, &exception, p.Meta.App.Release)
		kv := transformedException.KeyVal()
		payload.MergeKeyVal(kv, meta)
		errs = append(errs, exp.sendKeyValsToLogsPipeline(ctx, kv))
	}

	// measurements
	for _, measurement := range p.Measurements {
		kv := measurement.KeyVal()
		payload.MergeKeyVal(kv, meta)
		errs = append(errs, exp.sendKeyValsToLogsPipeline(ctx, kv))
	}

	// events
	for _, event := range p.Events {
		kv := event.KeyVal()
		payload.MergeKeyVal(kv, meta)
		errs = append(errs, exp.sendKeyValsToLogsPipeline(ctx, kv))
	}

	return errors.Join(errs...)
}

func (exp *logsExporter) sendKeyValsToLogsPipeline(ctx context.Context, kv *payload.KeyVal) error {
	// Grab the current value of exp.receivers so sendKeyValsToLogsPipeline
	// doesn't block updating receivers.
	exp.receiversMut.RLock()
	var (
		receivers = exp.receivers
	)
	exp.receiversMut.RUnlock()

	line, err := logfmt.MarshalKeyvals(payload.KeyValToInterfaceSlice(kv)...)
	if err != nil {
		level.Error(exp.log).Log("msg", "failed to logfmt a frontend log event", "err", err)
		return err
	}

	ent := loki.Entry{
		Labels: exp.labelSet(),
		Entry: logproto.Entry{
			Timestamp: time.Now(),
			Line:      string(line),
		},
	}

	ctx, cancel := context.WithTimeout(ctx, 2*time.Second) // TODO(rfratto): potentially make this configurable
	defer cancel()

	for _, receiver := range receivers {
		select {
		case <-ctx.Done():
			return err
		case receiver.Chan() <- ent:
			continue
		}
	}

	return nil
}

func (exp *logsExporter) labelSet() model.LabelSet {
	exp.labelsMut.RLock()
	defer exp.labelsMut.RUnlock()
	return exp.labels
}

func (exp *logsExporter) SetLabels(newLabels map[string]string) {
	exp.labelsMut.Lock()
	defer exp.labelsMut.Unlock()

	ls := make(model.LabelSet, len(newLabels))
	for k, v := range newLabels {
		ls[model.LabelName(k)] = model.LabelValue(v)
	}
	exp.labels = ls
}

//
// Traces
//

type tracesExporter struct {
	log log.Logger

	mut       sync.RWMutex
	consumers []otelcol.Consumer
}

var _ exporter = (*tracesExporter)(nil)

func newTracesExporter(log log.Logger) *tracesExporter {
	return &tracesExporter{
		log: log,
	}
}

// SetConsumers updates the set of OTLP consumers which will receive traces
// emitted by the exporter.
func (exp *tracesExporter) SetConsumers(consumers []otelcol.Consumer) {
	exp.mut.Lock()
	defer exp.mut.Unlock()

	exp.consumers = consumers
}

func (exp *tracesExporter) Name() string { return "traces exporter" }

func (exp *tracesExporter) Export(ctx context.Context, p payload.Payload) error {
	if p.Traces == nil {
		return nil
	}

	var errs []error
	for _, consumer := range exp.getTracesConsumers() {
		errs = append(errs, consumer.ConsumeTraces(ctx, p.Traces.Traces))
	}
	return errors.Join(errs...)
}

func (exp *tracesExporter) getTracesConsumers() []otelcol.Consumer {
	exp.mut.RLock()
	defer exp.mut.RUnlock()

	return exp.consumers
}
