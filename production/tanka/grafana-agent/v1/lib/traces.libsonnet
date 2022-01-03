{
  // withTracesConfig adds a Traces config to collect traces.
  //
  // For the full list of options, refer to the configuration reference:
  //
  withTracesConfig(config):: {
    assert std.objectHasAll(self, '_mode') : |||
      withTracesConfig must be merged with the result of calling new.
    |||,
    _trace_config:: config,
  },

  // withTracesRemoteWrite configures one or multiple backends to write traces to.
  //
  // Available options can be found in the configuration reference:
  // https://github.com/grafana/agent/blob/main/docs/configuration-reference.md#traces_config
  withTracesRemoteWrite(remote_write):: {
    assert std.objectHasAll(self, '_trace_config') : |||
      withTracesRemoteWrite must be merged with the result of calling
      withTracesConfig.
    |||,
    _trace_config+:: { remote_write: remote_write },
  },

  // withTracesSamplingStrategies accepts an object for trace sampling strategies.
  //
  // Refer to Jaeger's documentation for available fields:
  // https://www.jaegertracing.io/docs/1.17/sampling/#collector-sampling-configuration
  //
  // Creating a file isn't necessary; just provide the object and a ConfigMap
  // will be created for you and added to the tempo config.
  withTracesSamplingStrategies(strategies):: {
    assert std.objectHasAll(self, '_trace_config') : |||
      withTracesPushConfig must be merged with the result of calling
      withTracesConfig.
    |||,

    assert
    std.objectHasAll(self._trace_config, 'receivers') &&
    std.objectHasAll(self._trace_config.receivers, 'jaeger') : |||
        withStrategies can only be used if the traces config is configured for
        receiving Jaeger spans and traces.
      |||,

    // The main library should detect the presence of _traces_sampling_strategies
    // and create a ConfigMap bound to /etc/agent/strategies.json.
    _traces_sampling_strategies:: strategies,
    _trace_config+:: {
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
  withTracesScrapeConfigs(scrape_configs):: {
    assert std.objectHasAll(self, '_trace_config') : |||
      withTracesScrapeConfigs must be merged with the result of calling
      withTracesConfig.
    |||,
    _trace_config+: { scrape_configs: scrape_configs },
  },

  // Provides a default set of scrape_configs to use for discovering labels from
  // Pods. Labels will be attached to any traces sent from the discovered pods.
  tracesScrapeKubernetes:: [
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

  // withTracesTailSamplingConfig tail-based sampling for traces.
  //
  // Available options can be found in the configuration reference:
  // https://github.com/grafana/agent/blob/main/docs/configuration-reference.md#traces_config
  withTracesTailSamplingConfig(tail_sampling):: {
    assert std.objectHasAll(self, '_trace_config') : |||
      withTracesTailSamplingConfig must be merged with the result of calling
      withTracesConfig.
    |||,
    _trace_config+:: { tail_sampling: tail_sampling },
  },

  withTracesLoadBalancingConfig(load_balancing):: {
    assert std.objectHasAll(self, '_trace_config') : |||
      withTracesLoadBalancingConfig must be merged with the result of calling
      withTracesConfig.
    |||,
    _trace_config+:: { load_balancing: load_balancing },
  },
}
