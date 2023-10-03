package app_agent_receiver

import (
	"context"

	"github.com/go-kit/log"
	"github.com/grafana/agent/component"
	internal "github.com/grafana/agent/pkg/integrations/v2/app_agent_receiver"
)

func init() {
	component.Register(component.Registration{
		Name:    "integrations.v2.app_agent_receiver",
		Args:    Arguments{},
		Exports: Exports{},

		Build: func(opts component.Options, args component.Arguments) (component.Component, error) {
			return New(opts, args.(Arguments))
		},
	})
}

type Exports struct {
	Config internal.Config `river:"self,attr"`
}

type Component struct {
	prefixedRegisterer *prefixedRegistry

	metrics *metricsExporter
	logs    *logsExporter
	traces  *tracesExporter

	exporters []exporter
}

func New(o component.Options, args Arguments) (*Component, error) {
	// NOTE(rfratto): by default, all metrics are prefixed with faro_receiver.
	// For backwards compatibility with users of the static mode integration, the
	// prefix can be changed to app_agent_receiver.
	prefixedRegistry := newPrefixedRegistry("faro_receiver")
	o.Registerer.MustRegister(prefixedRegistry)

	var (
		metrics = newMetricsExporter(prefixedRegistry)
		logs    = newLogsExporter(log.With(o.Logger, "exporter", "logs"), nil) // TODO(rfratto): lazy sourcemaps
		traces  = newTracesExporter(log.With(o.Logger, "exporter", "traces"))
	)

	c := &Component{
		prefixedRegisterer: prefixedRegistry,

		metrics: metrics,
		logs:    logs,
		traces:  traces,

		exporters: []exporter{metrics, logs, traces},
	}

	if err := c.Update(args); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *Component) Run(ctx context.Context) error {
	// TODO(rfratto):
	//
	// * Start HTTP server for collecting telemetry from Faro clients.
	// * Metrics for HTTP server:
	//   * app_agent_receiver_request_duration_seconds
	//   * app_agent_receiver_request_message_bytes
	//   * app_agent_receiver_response_message_bytes
	//   * app_agent_receiver_inflight_requests

	return nil
}

func (c *Component) Update(args component.Arguments) error {
	newArgs := args.(Arguments)

	c.logs.SetReceivers(newArgs.Output.Logs)
	c.traces.SetConsumers(newArgs.Output.Traces)

	// TODO(rfratto):
	//
	// * Ensure server gets restarted with new settings.
	// * Allow updating prefix of prefixCollector.

	return nil
}
