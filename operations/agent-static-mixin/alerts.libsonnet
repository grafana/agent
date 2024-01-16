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
                Load balancing is experiencing {{ printf "%.2f" $value }}% errors.
              |||,
            },
          },
        ],
      },
      {
        name: 'VultureChecks',
        rules: [
          {
            alert: 'VultureDown',
            expr: |||
              up{job=~"agent-smoke-test/vulture"} == 0
            |||,
            'for': '5m',
            annotations: {
              summary: 'Vulture {{ $labels.job }} is down.',
            },
          },
          {
            alert: 'VultureFlapping',
            expr: |||
              avg_over_time(up{job=~"agent-smoke-test/vulture"}[5m]) < 1
            |||,
            'for': '15m',
            annotations: {
              summary: 'Vulture {{ $labels.job }} is flapping.',
            },
          },
          {
            alert: 'VultureNotScraped',
            expr: |||
              rate(tempo_vulture_trace_total[1m]) == 0
            |||,
            'for': '5m',
            annotations: {
              summary: 'Vulture {{ $labels.job }} is not being scraped.',
            },
          },
          {
            alert: 'VultureFailures',
            expr: |||
              (rate(tempo_vulture_error_total[5m]) / rate(tempo_vulture_trace_total[5m])) > 0.3
            |||,
            'for': '5m',
            annotations: {
              summary: 'Vulture {{ $labels.job }} has had failures for at least 5m',
            },
          },
        ],
      },
      {
        name: 'GrafanaAgentConfig',
        rules: [
          {
            alert: 'AgentRemoteConfigBadAPIRequests',
            expr: |||
              100 * sum(rate(agent_remote_config_fetches_total{status_code=~"(4|5).."}[10m])) by (%(group_by_cluster)s)
                /
              sum(rate(agent_remote_config_fetches_total[10m])) by (%(group_by_cluster)s)
                > 5
            ||| % _config,
            'for': '10m',
            labels: {
              severity: 'warning',
            },
            annotations: {
              message: |||
                Receiving HTTP {{ $labels.status_code }} errors from API in {{ printf "%.2f" $value }}% of cases.
              |||,
            },
          },
          {
            alert: 'AgentRemoteConfigBadAPIRequests',
            expr: |||
              100 * sum(rate(agent_remote_config_fetches_total{status_code=~"(4|5).."}[10m])) by (%(group_by_cluster)s)
                /
              sum(rate(agent_remote_config_fetches_total[10m])) by (%(group_by_cluster)s)
                > 10
            ||| % _config,
            'for': '10m',
            labels: {
              severity: 'critical',
            },
            annotations: {
              message: |||
                Receiving HTTP {{ $labels.status_code }} errors from API in {{ printf "%.2f" $value }}% of cases.
              |||,
            },
          },
          {
            alert: 'AgentRemoteConfigFetchErrors',
            expr: |||
              100 * sum(rate(agent_remote_config_fetch_errors_total[10m])) by (%(group_by_cluster)s)
                /
              sum(rate(agent_remote_config_fetches_total[10m])) by (%(group_by_cluster)s)
                > 5
            ||| % _config,
            'for': '10m',
            labels: {
              severity: 'warning',
            },
            annotations: {
              message: |||
                Failing to reach Agent Management API.
              |||,
            },
          },
          {
            alert: 'AgentRemoteConfigFetchErrors',
            expr: |||
              100 * sum(rate(agent_remote_config_fetch_errors_total[10m])) by (%(group_by_cluster)s)
                /
              sum(rate(agent_remote_config_fetches_total[10m])) by (%(group_by_cluster)s)
                > 10
            ||| % _config,
            'for': '10m',
            labels: {
              severity: 'critical',
            },
            annotations: {
              message: |||
                Failing to reach Agent Management API.
              |||,
            },
          },
          {
            alert: 'AgentRemoteConfigInvalidAPIResponse',
            expr: |||
              100 * sum(rate(agent_remote_config_invalid_total{reason=~".+"}[10m])) by (%(group_by_cluster)s)
                /
              sum(rate(agent_remote_config_fetches_total[10m])) by (%(group_by_cluster)s)
                > 5
            ||| % _config,
            'for': '10m',
            labels: {
              severity: 'warning',
            },
            annotations: {
              message: |||
                API is responding with {{ $labels.reason }} in {{ printf "%.2f" $value }}% of cases.
              |||,
            },
          },
          {
            alert: 'AgentRemoteConfigInvalidAPIResponse',
            expr: |||
              100 * sum(rate(agent_remote_config_invalid_total{reason=~".+"}[10m])) by (%(group_by_cluster)s)
                /
              sum(rate(agent_remote_config_fetches_total[10m])) by (%(group_by_cluster)s)
                > 10
            ||| % _config,
            'for': '10m',
            labels: {
              severity: 'critical',
            },
            annotations: {
              message: |||
                API is responding with {{ $labels.reason }} in {{ printf "%.2f" $value }}% of cases.
              |||,
            },
          },
          {
            alert: 'AgentFailureToReloadConfig',
            expr: |||
              avg_over_time(agent_config_last_load_successful[10m]) < 0.9
            ||| % _config,
            'for': '10m',
            labels: {
              severity: 'warning',
            },
            annotations: {
              message: |||
                Instance {{ $labels.instance }} failed to successfully reload the config.
              |||,
            },
          },
          {
            alert: 'AgentFailureToReloadConfig',
            expr: |||
              avg_over_time(agent_config_last_load_successful[10m]) < 0.9
            ||| % _config,
            'for': '30m',
            labels: {
              severity: 'critical',
            },
            annotations: {
              message: |||
                Instance {{ $labels.instance }} failed to successfully reload the config.
              |||,
            },
          },
          {
            alert: 'AgentManagementFallbackToEmptyConfig',
            expr: |||
              sum(rate(agent_management_config_fallbacks_total{fallback_to="empty_config"}[10m])) by (%(group_by_cluster)s) > 0
            ||| % _config,
            'for': '10m',
            labels: {
              severity: 'warning',
            },
            annotations: {
              message: |||
                Instance {{ $labels.instance }} fell back to empty configuration.
              |||,
            },
          },
          {
            alert: 'AgentManagementFallbackToEmptyConfig',
            expr: |||
              sum(rate(agent_management_config_fallbacks_total{fallback_to="empty_config"}[10m])) by (%(group_by_cluster)s) > 0
            ||| % _config,
            'for': '30m',
            labels: {
              severity: 'critical',
            },
            annotations: {
              message: |||
                Instance {{ $labels.instance }} fell back to empty configuration.
              |||,
            },
          },
        ],
      },
    ],
  },
}
