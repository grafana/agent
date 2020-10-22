local agent = import 'grafana-agent/v1/main.libsonnet';

local k = import 'ksonnet-util/kausal.libsonnet';
local containerPort = k.core.v1.containerPort;

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
          protocols: {
            thrift_http: null,
            thrift_binary: null,
            thrift_compact: null,
            grpc: null,
          },
        },
        zipkin: null,
        otlp: {
          protocols: {
            http: null,
            grpc: null,
          },
        },
        opencensus: null,
      },
    }) +
    agent.withPortsMixin([
      // Jaeger receiver
      containerPort.new('tempo-jaeger-thrift-compact', 6831) + containerPort.withProtocol('UDP'),
      containerPort.new('tempo-jaeger-thrift-binary', 6832) + containerPort.withProtocol('UDP'),
      containerPort.new('tempo-jaeger-thrift-http', 14268) + containerPort.withProtocol('TCP'),
      containerPort.new('tempo-jaeger-grpc', 14250) + containerPort.withProtocol('TCP'),

      // Zipkin
      containerPort.new('tempo-zipkin', 9411) + containerPort.withProtocol('TCP'),

      // OTLP
      containerPort.new('tempo-otlp', 55680) + containerPort.withProtocol('TCP'),

      // Opencensus
      containerPort.new('tempo-opencensus', 55678) + containerPort.withProtocol('TCP'),
    ]) +
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
