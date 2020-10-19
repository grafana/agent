local agent = import 'grafana-agent/v1/main.libsonnet';

{
  agent:
    agent.new('grafana-agent-traces', 'default') +
    agent.withConfigHash(false) +
    agent.withImages({
      agent: (import 'version.libsonnet'),
    }) +
    agent.withTempoConfig({
      receivers: {
        jaeger: {
          protocols: { thrift_compact: null, grpc: null },
        },
      },
    }) +
    agent.withTempoPushConfig({
      endpoint: '${TEMPO_ENDPOINT}',
      basic_auth: {
        username: '${TEMPO_USERNAME}',
        password: '${TEMPO_PASSWORD}',
      },
      batch: {
        timeout: '5s',
        send_batch_size: 1000,
      },
      queue: {
        retry_on_failure: false,
      },
    }) +
    agent.withTempoSamplingStrategies({
      default_strategy: {
        type: 'probabilistic',
        param: 0.001,
      },
    }) +
    agent.withTempoScrapeConfigs(agent.tempoScrapeKubernetes),
}
