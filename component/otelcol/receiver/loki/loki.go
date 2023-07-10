// Package loki provides an otelcol.receiver.loki component.
package loki

import (
	"context"
	"path"
	"strings"
	"sync"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/grafana/agent/component"
	"github.com/grafana/agent/component/common/loki"
	"github.com/grafana/agent/component/otelcol"
	"github.com/grafana/agent/component/otelcol/internal/fanoutconsumer"
	loki_translator "github.com/open-telemetry/opentelemetry-collector-contrib/pkg/translator/loki"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/pdata/plog"
)

func init() {
	component.Register(component.Registration{
		Name:    "otelcol.receiver.loki",
		Args:    Arguments{},
		Exports: Exports{},

		Build: func(o component.Options, a component.Arguments) (component.Component, error) {
			return New(o, a.(Arguments))
		},
	})
}

var hintAttributes = "loki.attribute.labels"

// Arguments configures the otelcol.receiver.loki component.
type Arguments struct {
	// Output configures where to send received data. Required.
	Output *otelcol.ConsumerArguments `river:"output,block"`
}

// Exports holds the receiver that is used to send log entries to the
// loki.write component.
type Exports struct {
	Receiver loki.LogsReceiver `river:"receiver,attr"`
}

// Component is the otelcol.receiver.loki component.
type Component struct {
	log  log.Logger
	opts component.Options

	mut      sync.RWMutex
	receiver loki.LogsReceiver
	logsSink consumer.Logs
}

var _ component.Component = (*Component)(nil)

// New creates a new otelcol.receiver.loki component.
func New(o component.Options, c Arguments) (*Component, error) {
	// TODO(@tpaschalis) Create a metrics struct to count
	// total/successful/errored log entries?
	res := &Component{
		log:  o.Logger,
		opts: o,
	}

	// Create and immediately export the receiver which remains the same for
	// the component's lifetime.
	res.receiver = loki.NewLogsReceiver()
	o.OnStateChange(Exports{Receiver: res.receiver})

	if err := res.Update(c); err != nil {
		return nil, err
	}
	return res, nil
}

// Run implements Component.
func (c *Component) Run(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		case entry := <-c.receiver.Chan():

			logs := convertLokiEntryToPlog(entry)

			// TODO(@tpaschalis) Is there any more handling to be done here?
			err := c.logsSink.ConsumeLogs(ctx, logs)
			if err != nil {
				level.Error(c.opts.Logger).Log("msg", "failed to consume log entries", "err", err)
			}
		}
	}
}

// Update implements Component.
func (c *Component) Update(newConfig component.Arguments) error {
	c.mut.Lock()
	defer c.mut.Unlock()

	cfg := newConfig.(Arguments)
	c.logsSink = fanoutconsumer.Logs(cfg.Output.Logs)

	return nil
}

// Create a new Otlp Logs entry from a Promtail entry
func convertLokiEntryToPlog(lokiEntry loki.Entry) plog.Logs {
	logs := plog.NewLogs()

	lr := logs.ResourceLogs().AppendEmpty().ScopeLogs().AppendEmpty().LogRecords().AppendEmpty()

	if filename, exists := lokiEntry.Labels["filename"]; exists {
		filenameStr := string(filename)
		// The `promtailreceiver` from the opentelemetry-collector-contrib
		// repo adds these two labels based on these "semantic conventions
		// for log media".
		// https://opentelemetry.io/docs/reference/specification/logs/semantic_conventions/media/
		// We're keeping them as well, but we're also adding the `filename`
		// attribute so that it can be used from the
		// `loki.attribute.labels` hint for when the opposite OTel -> Loki
		// transformation happens.
		lr.Attributes().PutStr("log.file.path", filenameStr)
		lr.Attributes().PutStr("log.file.name", path.Base(filenameStr))
		// TODO(@tpaschalis) Remove the addition of "log.file.path" and "log.file.name",
		// because the Collector doesn't do it and we would be more in line with it.
	}

	var lbls []string
	for key := range lokiEntry.Labels {
		keyStr := string(key)
		lbls = append(lbls, keyStr)
	}

	if len(lbls) > 0 {
		// This hint is defined in the pkg/translator/loki package and the
		// opentelemetry-collector-contrib repo, but is not exported so we
		// re-define it.
		// It is used to detect which attributes should be promoted to labels
		// when transforming back from OTel -> Loki.
		lr.Attributes().PutStr(hintAttributes, strings.Join(lbls, ","))
	}

	loki_translator.ConvertEntryToLogRecord(&lokiEntry.Entry, &lr, lokiEntry.Labels, true)

	return logs
}
