local config = import 'config.libsonnet';
local _config = config._config;

{
  prometheusAlerts+:: {
    groups+: [
      {
        name: 'grafana-agent-tracing',
        rules: [
          {
            alert: 'AgentTracingReceiverErrors',
            // TODO(@mapno): add recording rule for total spans
            expr: |||
              100 * sum(rate(traces_receiver_refused_spans{receiver!="otlp/lb"}[1m])) by (%(group_by_cluster)s, receiver)
                /
              (sum(rate(traces_receiver_refused_spans{receiver!="otlp/lb"}[1m])) by (%(group_by_cluster)s, receiver) + sum(rate(traces_receiver_accepted_spans{receiver!="otlp/lb"}[1m])) by (%(group_by_cluster)s, receiver))
                > 10
            ||| % _config,
            'for': '15m',
            labels: {
              severity: 'warning',
            },
            annotations: {
              message: |||
                Receiver {{ $labels.receiver }} is experiencing {{ printf "%.2f" $value }}% errors.
              |||,
            },
          },
          {
            alert: 'AgentTracingExporterErrors',
            // TODO(@mapno): add recording rule for total spans
            expr: |||
              100 * sum(rate(traces_exporter_send_failed_spans{exporter!="otlp"}[1m])) by (%(group_by_cluster)s, exporter)
                /
              (sum(rate(traces_exporter_send_failed_spans{exporter!="otlp"}[1m])) by (%(group_by_cluster)s, exporter) + sum(rate(traces_exporter_sent_spans{exporter!="otlp"}[1m])) by (%(group_by_cluster)s, exporter))
                > 10
            ||| % _config,
            'for': '15m',
            labels: {
              severity: 'warning',
            },
            annotations: {
              message: |||
                Exporter {{ $labels.exporter }} is experiencing {{ printf "%.2f" $value }}% errors.
              |||,
            },
          },
          {
            alert: 'AgentTracingLoadBalancingErrors',
            expr: |||
              100 * sum(rate(traces_loadbalancer_backend_outcome{success="false"}[1m])) by (%(group_by_cluster)s)
                /
              sum(rate(traces_loadbalancer_backend_outcome{success="true"}[1m])) by (%(group_by_cluster)s)
                > 10
            ||| % _config,
            'for': '15m',
            labels: {
              severity: 'warning',
            },
            annotations: {
              message: |||
                Load balacing is experiencing {{ printf "%.2f" $value }}% errors.
              |||,
            },
          },
        ],
      },
    ],
  },
}
