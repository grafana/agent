local agent_prometheus = import 'grafana-agent/v1/lib/prometheus.libsonnet';

{
  config: {
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
    ],
  },
}
