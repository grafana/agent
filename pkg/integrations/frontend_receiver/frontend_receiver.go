package frontend_receiver //nolint:golint

import (
	"context"
	"log"
	"net/http"
	"time"

	kitlog "github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/go-logfmt/logfmt"
	loki "github.com/grafana/agent/pkg/logs"
	"github.com/grafana/grafana-frontend-telemetry-receiver/pkg/exporters"
	"github.com/grafana/grafana-frontend-telemetry-receiver/pkg/models"
	"github.com/grafana/grafana-frontend-telemetry-receiver/pkg/receiver"
	"github.com/grafana/grafana-frontend-telemetry-receiver/pkg/utils"
	recvutils "github.com/grafana/grafana-frontend-telemetry-receiver/pkg/utils"
	promtailapi "github.com/grafana/loki/clients/pkg/promtail/api"
	"github.com/grafana/loki/pkg/logproto"

	"github.com/grafana/agent/pkg/integrations"
	"github.com/grafana/agent/pkg/integrations/config"
)

// Integration is the frontend_receiver integration. It exposes a frontend telemetry
// collection http endpoint that can receive payloads from grafana frontend agent
// and forward this data to logs, tempo and metrics
type Integration struct {
	receiver     *receiver.FrontendReceiver
	logger       kitlog.Logger
	logsInstance *loki.Instance
	config       *Config
}

func init() {
	integrations.RegisterIntegration(&Config{})
}

// Handlers satisfies Integration.Handlers.
func (i *Integration) Handlers() (map[string]http.Handler, error) {
	logger := log.New(kitlog.NewStdlibAdapter(i.logger), "frontend_receiver:collector", 0)
	return map[string]http.Handler{
		i.config.Endpoint: i.receiver.ReceiverHandler(logger),
	}, nil
}

// ScrapeConfigs satisfies Integration.ScrapeConfigs.
func (i *Integration) ScrapeConfigs() []config.ScrapeConfig {
	return []config.ScrapeConfig{}
}

// Run satisfies Integration.Run.
func (i *Integration) Run(ctx context.Context) error {
	// We don't need to do anything here, so we can just wait for the context to
	// finish.
	<-ctx.Done()
	return ctx.Err()
}

// New creates a new frontend_receiver integration.
func New(logger kitlog.Logger, c *Config, logsInstance *loki.Instance) (integrations.Integration, error) {
	level.Debug(logger).Log("msg", "initializing frontend receiver", "endpoint", c.Endpoint)

	exporters := []exporters.FrontendReceiverExporter{
		logsPipelineExporter(logger, c, logsInstance),
	}

	receiver := receiver.NewFrontendReceiver(c.Receiver, exporters)

	integration := &Integration{
		receiver:     &receiver,
		config:       c,
		logsInstance: logsInstance,
	}

	return integration, nil
}

func sendKeyValsToLogsPipeline(logger kitlog.Logger, kv *utils.KeyVal, c *Config, logsInstance *loki.Instance) {
	line, err := logfmt.MarshalKeyvals(utils.KeyValToInterfaceSlice(kv)...)
	if err != nil {
		level.Error(logger).Log("msg", "failed to logfmt a frontend log event", "err", err)
		return
	}
	sent := logsInstance.SendEntry(promtailapi.Entry{
		Labels: c.labelSet(kv),
		Entry: logproto.Entry{
			Timestamp: time.Now(),
			Line:      string(line),
		},
	}, c.LogsTimeout)
	if !sent {
		level.Warn(logger).Log("msg", "failed to log frontend log event to logs pipeline")
	}
}

func logsPipelineExporter(logger kitlog.Logger, c *Config, logsInstance *loki.Instance) exporters.FrontendReceiverExporter {
	return func(payload models.Payload) error {
		meta := payload.Meta.KeyVal()

		// log events
		for _, logItem := range payload.Logs {
			kv := logItem.KeyVal()
			recvutils.MergeKeyVal(kv, meta)
			sendKeyValsToLogsPipeline(logger, kv, c, logsInstance)
		}
		// exceptions
		for _, exception := range payload.Exceptions {
			kv := exception.KeyVal()
			recvutils.MergeKeyVal(kv, meta)
			sendKeyValsToLogsPipeline(logger, kv, c, logsInstance)
		}
		// measurements
		for _, measurement := range payload.Measurements {
			kv := measurement.KeyVal()
			recvutils.MergeKeyVal(kv, meta)
			sendKeyValsToLogsPipeline(logger, kv, c, logsInstance)
		}
		return nil
	}
}
