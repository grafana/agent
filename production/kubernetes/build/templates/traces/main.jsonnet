local agent = import 'grafana-agent/v1/main.libsonnet';

local k = import 'ksonnet-util/kausal.libsonnet';
local containerPort = k.core.v1.containerPort;

local newPort(name, portNumber, protocol='TCP') =
  // Port names for pods cannot be longer than 15 characters.
  if std.length(name) > 15 then
    error 'port name cannot be longer than 15 characters'
  else containerPort.new(name, portNumber) + containerPort.withProtocol(protocol);

{
  agent:
    agent.newDeployment('grafana-agent-traces', 'YOUR_NAMESPACE') +
    agent.withConfigHash(false) +
    agent.withImages({
      agent: (import 'version.libsonnet'),
    }) +
    agent.withTracesConfig({
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
      batch: {
        timeout: '5s',
        send_batch_size: 1000,
      },
    }) +
    agent.withPortsMixin([
      // Jaeger receiver
      newPort('thrift-compact', 6831, 'UDP'),
      newPort('thrift-binary', 6832, 'UDP'),
      newPort('thrift-http', 14268, 'TCP'),
      newPort('thrift-grpc', 14250, 'TCP'),

      // Zipkin
      newPort('zipkin', 9411, 'TCP'),

      // OTLP
      newPort('otlp', 55680, 'TCP'),

      // Opencensus
      newPort('opencensus', 55678, 'TCP'),
    ]) +
    agent.withTracesRemoteWrite([
      {
        endpoint: '${TEMPO_ENDPOINT}',
        basic_auth: {
          username: '${TEMPO_USERNAME}',
          password: '${TEMPO_PASSWORD}',
        },
        retry_on_failure: {
          enabled: false,
        },
      },
    ]) +
    agent.withTracesSamplingStrategies({
      default_strategy: {
        type: 'probabilistic',
        param: 0.001,
      },
    }) +
    agent.withTracesScrapeConfigs(agent.tracesScrapeKubernetes) + {
      agent+: {
        // Remove this block to generate ConfigMap
        config_map:: {},
      },
    },
}
