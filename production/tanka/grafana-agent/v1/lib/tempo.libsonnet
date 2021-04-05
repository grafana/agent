{
  // withTempoConfig adds a Tempo config to collect traces.
  //
  // For the full list of options, refer to the configuration reference:
  //
  withTempoConfig(config):: {
    assert std.objectHasAll(self, '_mode') : |||
      withLokiConfig must be merged with the result of calling new.
    |||,
    _tempo_config:: config,
  },

  // Deprecated in favor of withTempoRemoteWrite.
  // withTempoPushConfig configures a location to write traces to.
  //
  // Availabile options can be found in the configuration reference:
  // https://github.com/grafana/agent/blob/main/docs/configuration-reference.md#tempo_config
  withTempoPushConfig(push_config):: {
    assert std.objectHasAll(self, '_tempo_config') : |||
      withTempoPushConfig must be merged with the result of calling
      withTempoConfig.
    |||,
    _tempo_config+:: { push_config: push_config },
  },

  // withTempoRemoteWrite configures one or multiple backends to write traces to.
  //
  // Availabile options can be found in the configuration reference:
  // https://github.com/grafana/agent/blob/main/docs/configuration-reference.md#tempo_config
  withTempoRemoteWrite(remote_write):: {
    assert std.objectHasAll(self, '_tempo_config') : |||
      withTempoRemoteWrite must be merged with the result of calling
      withTempoConfig.
    |||,
    _tempo_config+:: { remote_write: remote_write },
  },

  // withTempoSamplingStrategies accepts an object for trace sampling strategies.
  //
  // Refer to Jaeger's documentation for available fields:
  // https://www.jaegertracing.io/docs/1.17/sampling/#collector-sampling-configuration
  //
  // Creating a file isn't necessary; just provide the object and a ConfigMap
  // will be created for you and added to the tempo config.
  withTempoSamplingStrategies(strategies):: {
    assert std.objectHasAll(self, '_tempo_config') : |||
      withTempoPushConfig must be merged with the result of calling
      withTempoConfig.
    |||,

    assert
    std.objectHasAll(self._tempo_config, 'receivers') &&
    std.objectHasAll(self._tempo_config.receivers, 'jaeger') : |||
        withStrategies can only be used if the tempo config is configured for
        receiving Jaeger spans and traces.
      |||,

    // The main library should detect the presence of _tempo_sampling_strategies
    // and create a ConfigMap bound to /etc/agent/strategies.json.
    _tempo_sampling_strategies:: strategies,
    _tempo_config+:: {
      receivers+: {
        jaeger+: {
          remote_sampling: {
            strategy_file: '/etc/agent/strategies.json',
            insecure: true,
          },
        },
      },
    },
  },

  // Configures scrape_configs for discovering meta labels that will be attached
  // to incoming metrics and spans whose IP matches the __address__ of the
  // target.
  withTempoScrapeConfigs(scrape_configs):: {
    assert std.objectHasAll(self, '_tempo_config') : |||
      withTempoScrapeConfigs must be merged with the result of calling
      withTempoConfig.
    |||,
    _tempo_config+: { scrape_configs: scrape_configs },
  },

  // Provides a default set of scrape_configs to use for discovering labels from
  // Pods. Labels will be attached to any traces sent from the discovered pods.
  tempoScrapeKubernetes:: [
    {
      bearer_token_file: '/var/run/secrets/kubernetes.io/serviceaccount/token',
      job_name: 'kubernetes-pods',
      kubernetes_sd_configs: [{ role: 'pod' }],
      relabel_configs: [
        {
          action: 'replace',
          source_labels: ['__meta_kubernetes_namespace'],
          target_label: 'namespace',
        },
        {
          action: 'replace',
          source_labels: ['__meta_kubernetes_pod_name'],
          target_label: 'pod',
        },
        {
          action: 'replace',
          source_labels: ['__meta_kubernetes_pod_container_name'],
          target_label: 'container',
        },
      ],
      tls_config: {
        ca_file: '/var/run/secrets/kubernetes.io/serviceaccount/ca.crt',
        insecure_skip_verify: false,
      },
    },
  ],
}
