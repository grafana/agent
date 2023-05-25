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
        name: 'GrafanaAgentSmokeChecks',
        rules: [
          {
            alert: 'GrafanaAgentDown',
            expr: |||
              up{
                namespace="agent-smoke-test",
                pod=~"grafana-agent-smoke-test-(0|cluster-0|cluster-1|cluster-2)",
              } == 0
            |||,
            'for': '5m',
            annotations: {
              summary: '{{ $labels.job }} is down',
            },
          },
          {
            alert: 'GrafanaAgentFlapping',
            expr: |||
              avg_over_time(up{
                namespace="agent-smoke-test",
                pod=~"grafana-agent-smoke-test-(0|cluster-0|cluster-1|cluster-2)",
              }[5m]) < 1
            |||,
            'for': '15m',
            annotations: {
              summary: '{{ $labels.job }} is flapping',
            },
          },

          // Checks that the CPU usage doesn't go too high. This was generated from internal usage where
          // every 1,000 active series used roughly 0.0013441% of CPU. This alert only fires if there is a
          // minimum load threshold of at least 1000 active series.
          {
            alert: 'GrafanaAgentCPUHigh',
            expr: |||
              (sum by (pod) (rate(container_cpu_usage_seconds_total{cluster=~".+", namespace=~"agent-smoke-test", container=~".+", pod="grafana-agent-smoke-test-cluster-2"}[5m]))
              /
              (sum by (pod) (agent_wal_storage_active_series{cluster=~".+", namespace=~"agent-smoke-test", container=~".+", pod="grafana-agent-smoke-test-cluster-2"}) / 1000)
              > 0.0013441)
              and
              sum by (pod) (agent_wal_storage_active_series{cluster=~".+", namespace=~"agent-smoke-test", container=~".+", pod="grafana-agent-smoke-test-cluster-2"}) > 1000
            |||,
            'for': '1h',
            annotations: {
              summary: '{{ $labels.pod }} is using more than 0.0013441 CPU per 1000 series over the last 5 minutes',
            },
          },

          // We assume roughly ~8KB per series. Check that each deployment
          // doesn't go too far above this.
          //
          // We aggregate the memory of the scraping service together since an individual
          // node with a really small number of active series will throw this metric off.
          {
            alert: 'GrafanaAgentMemHigh',
            expr: |||
              sum without (pod, instance) (go_memstats_heap_inuse_bytes{job=~"agent-smoke-test/grafana-agent-smoke-test.*"}) /
              sum without (pod, instance, instance_group_name) (agent_wal_storage_active_series{job=~"agent-smoke-test/grafana-agent-smoke-test.*"}) / 1e3 > 10
            |||,
            'for': '1h',
            annotations: {
              summary: '{{ $labels.job }} has used more than 10KB per series for more than 5 minutes',
            },
          },
          {
            alert: 'GrafanaAgentContainerRestarts',
            expr: |||
              sum by (pod) (rate(kube_pod_container_status_restarts_total{namespace="agent-smoke-test"}[10m])) > 0
            |||,
            annotations: {
              summary: '{{ $labels.pod }} has a high rate of container restarts',
            },
          },
        ],
      },
      {
        name: 'GrafanaAgentCrowChecks',
        rules: [
          {
            alert: 'CrowDown',
            expr: |||
              up{job=~"agent-smoke-test/crow-.*"} == 0
            |||,
            'for': '5m',
            annotations: {
              summary: 'Crow {{ $labels.job }} is down.',
            },
          },
          {
            alert: 'CrowFlapping',
            expr: |||
              avg_over_time(up{job=~"agent-smoke-test/crow-.*"}[5m]) < 1
            |||,
            'for': '15m',
            annotations: {
              summary: 'Crow {{ $labels.job }} is flapping.',
            },
          },
          {
            alert: 'CrowNotScraped',
            expr: |||
              rate(crow_test_samples_total[5m]) == 0
            |||,
            'for': '15m',
            annotations: {
              summary: 'Crow {{ $labels.job }} is not being scraped.',
            },
          },
          {
            alert: 'CrowFailures',
            expr: |||
              (
                  rate(crow_test_sample_results_total{result="success"}[5m])
                  /
                  ignoring(result) sum without (result) (rate(crow_test_sample_results_total[5m]))
              )
              < 1
            |||,
            'for': '15m',
            annotations: {
              summary: 'Crow {{ $labels.job }} has had failures for at least 5m',
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
        name: 'GrafanaAgentManagement',
        rules: [
          {
            alert: 'AgentManagementBadAPIRequests',
            expr: |||
              100 * sum(rate(agent_remote_config_fetches_total{status_code="(4|5).."}[10m])) by (%(group_by_cluster)s)
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
            alert: 'AgentManagementBadAPIRequests',
            expr: |||
              100 * sum(rate(agent_remote_config_fetches_total{status_code="(4|5).."}[10m])) by (%(group_by_cluster)s)
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
            alert: 'AgentManagementRequestFailures',
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
            alert: 'AgentManagementRequestFailures',
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
            alert: 'AgentManagementInvalidAPIResponses',
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
            alert: 'AgentManagementInvalidAPIResponses',
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
            alert: 'AgentManagementFailureToReload',
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
            alert: 'AgentManagementFailureToReload',
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
                > 0
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
