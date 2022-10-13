local agent = import 'grafana-agent/v2/main.libsonnet';
local k = import 'ksonnet-util/kausal.libsonnet';

local containerPort = k.core.v1.containerPort;

local newPort(name, portNumber, protocol='TCP') =
  // Port names for pods cannot be longer than 15 characters.
  if std.length(name) > 15 then
    error 'port name cannot be longer than 15 characters'
  else containerPort.new(name, portNumber) + containerPort.withProtocol(protocol);

{
  agent:
    agent.new(name='grafana-agent-traces', namespace='${NAMESPACE}') +
    agent.withDeploymentController(replicas=1) +
    agent.withConfigHash(false) +
    agent.withPortsMixin([
      // Jaeger receiver
      newPort('thrift-compact', 6831, 'UDP'),
      newPort('thrift-binary', 6832, 'UDP'),
      newPort('thrift-http', 14268, 'TCP'),
      newPort('thrift-grpc', 14250, 'TCP'),

      // Zipkin
      newPort('zipkin', 9411, 'TCP'),

      // OTLP
      newPort('otlp-grpc', 4317, 'TCP'),
      newPort('otlp-http', 4318, 'TCP'),

      // Opencensus
      newPort('opencensus', 55678, 'TCP'),
    ]) + 
    agent.withService() +
    // add dummy config or will fail
    agent.withAgentConfig({
      server: { log_level: 'error' },
    }) + 
    // remove configMap for generated manifests
    { configMap:: super.configMap }
}
