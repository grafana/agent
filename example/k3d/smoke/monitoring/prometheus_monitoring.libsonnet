local agent_prometheus = import 'grafana-agent/v1/lib/metrics.libsonnet';

{
  config: {
    global: {
      scrape_interval: '1m',
    },
    scrape_configs: agent_prometheus.scrapeInstanceKubernetes.scrape_configs,
  },

  rules: {
    groups: [
      {
        name: 'GrafanaAgentChecks',
        rules: [
          // Basic sanity checks: ensure that Agents exist, are up,
          // and haven't been flapping.
          {
            alert: 'GrafanaAgentMissing',
            expr: |||
              absent(up{ namespace="smoke", pod="grafana-agent-0" })         == 1 or
              absent(up{ namespace="smoke", pod="grafana-agent-cluster-0" }) == 1 or
              absent(up{ namespace="smoke", pod="grafana-agent-cluster-1" }) == 1 or
              absent(up{ namespace="smoke", pod="grafana-agent-cluster-2" }) == 1
            |||,
            'for': '5m',
            annotations: {
              summary: '{{ $labels.pod }} is not running.',
            },
          },
          {
            alert: 'GrafanaAgentDown',
            expr: |||
              up{
                namespace="smoke",
                pod=~"grafana-agent-(0|cluster-0|cluster-1|cluster-2)",
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
                namespace="smoke",
                pod=~"grafana-agent-(0|cluster-0|cluster-1|cluster-2)",
              }[5m]) < 1
            |||,
            'for': '15m',
            annotations: {
              summary: '{{ $labels.job }} is flapping',
            },
          },

          // Checks that the CPU usage doesn't go too high. This was generated
          // from main where the CPU usage hovered around 2-3% per pod.
          //
          // TODO: something less guessworky here.
          {
            alert: 'GrafanaAgentCPUHigh',
            expr: |||
              rate(container_cpu_usage_seconds_total{namespace="smoke", pod=~"grafana-agent-.*"}[1m]) > 0.05
            |||,
            'for': '5m',
            annotations: {
              summary: '{{ $labels.pod }} is using more than 5% CPU over the last 5 minutes',
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
              sum without (pod, instance) (go_memstats_heap_inuse_bytes{job=~"smoke/grafana-agent.*"}) /
              sum without (pod, instance, instance_group_name) (agent_wal_storage_active_series{job=~"smoke/grafana-agent.*"}) / 1e3 > 10
            |||,
            'for': '5m',
            annotations: {
              summary: '{{ $labels.job }} has used more than 10KB per series for more than 5 minutes',
            },
          },
        ],
      },
      {
        name: 'CrowChecks',
        rules: [
          {
            alert: 'CrowMissing',
            expr: |||
              absent(up{container="crow-single"})  == 1 or
              absent(up{container="crow-cluster"}) == 1
            |||,
            'for': '5m',
            annotations: {
              summary: '{{ $labels.container }} is not running.',
            },
          },
          {
            alert: 'CrowDown',
            expr: |||
              up{job=~"smoke/crow-.*"} == 0
            |||,
            'for': '5m',
            annotations: {
              summary: 'Crow {{ $labels.job }} is down.',
            },
          },
          {
            alert: 'CrowFlapping',
            expr: |||
              avg_over_time(up{job=~"smoke/crow-.*"}[5m]) < 1
            |||,
            'for': '15m',
            annotations: {
              summary: 'Crow {{ $labels.job }} is flapping.',
            },
          },
          {
            alert: 'CrowNotScraped',
            expr: |||
              rate(crow_test_samples_total[1m]) == 0
            |||,
            'for': '5m',
            annotations: {
              summary: 'Crow {{ $labels.job }} is not being scraped.',
            },
          },
          {
            alert: 'CrowFailures',
            expr: |||
              (
                rate(crow_test_sample_results_total{result="success"}[1m])
                / ignoring(result) rate(crow_test_samples_total[1m])
              ) < 1
            |||,
            'for': '5m',
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
            alert: 'VultureMissing',
            expr: |||
              absent(up{container="vulture"})  == 1
            |||,
            'for': '5m',
            annotations: {
              summary: '{{ $labels.container }} is not running.',
            },
          },
          {
            alert: 'VultureDown',
            expr: |||
              up{job=~"smoke/vulture"} == 0
            |||,
            'for': '5m',
            annotations: {
              summary: 'Vulture {{ $labels.job }} is down.',
            },
          },
          {
            alert: 'VultureFlapping',
            expr: |||
              avg_over_time(up{job=~"smoke/vulture"}[5m]) < 1
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
        name: 'CanaryChecks',
        rules: [
          {
            alert: 'CanaryMissing',
            expr: |||
              absent(up{container="loki-canary"})  == 1
            |||,
            'for': '5m',
            annotations: {
              summary: '{{ $labels.container }} is not running.',
            },
          },
          {
            alert: 'CanaryDown',
            expr: |||
              up{job=~"smoke/loki-canary"} == 0
            |||,
            'for': '5m',
            annotations: {
              summary: ' Canary is down.',
            },
          },
          {
            alert: 'CanaryNotScraped',
            expr: |||
              rate(loki_canary_entries_total[1m]) == 0
            |||,
            'for': '5m',
            annotations: {
              summary: 'Canary is not being scraped.',
            },
          },
          {
            alert: 'CanaryMissingEntries',
            expr: |||
              (rate(loki_canary_missing_entries_total[2m])) > 0 
            |||,
            'for': '2m',
            annotations: {
              summary: 'Canary has had missing entries for at least 2m',
            },
          },
          {
            alert: 'CanarySpotChecksMissingEntries',
            expr: |||
              (rate(loki_canary_spot_check_missing_entries_total[2m])) > 0 
            |||,
            'for': '2m',
            annotations: {
              summary: 'Canary has had missing spot check entries for at least 2m',
            },
          },
          {
            alert: 'CanaryWebsocketMissingEntries',
            expr: |||
              (rate(loki_canary_websocket_missing_entries_total[2m])) > 0 
            |||,
            'for': '2m',
            annotations: {
              summary: 'Canary has had missing websocket entries for at least 2m',
            },
          },
          {
            alert: 'CanaryUnexpectedEntries',
            expr: |||
              (rate(loki_canary_unexpected_entries_total[2m])) > 0 
            |||,
            'for': '2m',
            annotations: {
              summary: 'Canary has had unexpected entries for at least 2m',
            },
          },
        ],
      },
    ],
  },
}
