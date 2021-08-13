local agent_prometheus = import 'grafana-agent/v1/lib/prometheus.libsonnet';

{
  config: {
    global: {
      scrape_interval: '15s',
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
    ],
  },
}
