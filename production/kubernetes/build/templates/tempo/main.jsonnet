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
      containerPort.new('thrift-c', 6831) + containerPort.withProtocol('UDP'),
      containerPort.new('thrift-bin', 6832) + containerPort.withProtocol('UDP'),
      containerPort.new('thrift-http', 14268) + containerPort.withProtocol('TCP'),
      containerPort.new('thrift-grpc', 14250) + containerPort.withProtocol('TCP'),

      // Zipkin
      containerPort.new('zipkin', 9411) + containerPort.withProtocol('TCP'),

      // OTLP
      containerPort.new('otlp', 55680) + containerPort.withProtocol('TCP'),

      // Opencensus
      containerPort.new('opencensus', 55678) + containerPort.withProtocol('TCP'),
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
      retry_on_failure: {
        enabled: false,
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
