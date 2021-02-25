local grafana_agent = import 'grafana-agent/v1/main.libsonnet';

grafana_agent.scrapeKubernetesLogs {
  local pipeline_stages = [
    // k3d uses cri for logging
    { cri: {} },

    // Bad words metrics, used in Agent dashboard
    {
      regex: {
        expression: '(?i)(?P<bad_words>panic:|core_dumped|failure|error|attack| bad |illegal |denied|refused|unauthorized|fatal|failed|Segmentation Fault|Corrupted)',
      },
    },
    {
      metrics: {
        panic_total: {
          type: 'Counter',
          description: 'total count of panic: found in log lines',
          source: 'panic',
          config: {
            action: 'inc',
          },
        },
        bad_words_total: {
          type: 'Counter',
          description: 'total count of bad words found in log lines',
          source: 'bad_words',
          config: {
            action: 'inc',
          },
        },
      },
    },
  ],

  scrape_configs: [
    x { pipeline_stages: pipeline_stages }
    for x
    in super.scrape_configs
  ],
}
