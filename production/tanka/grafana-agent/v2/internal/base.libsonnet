function(name='grafana-agent', namespace='') {
  local k = (import 'ksonnet-util/kausal.libsonnet') { _config+:: { namespace: namespace } },

  local container = k.core.v1.container,
  local configMap = k.core.v1.configMap,
  local containerPort = k.core.v1.containerPort,
  local policyRule = k.rbac.v1.policyRule,
  local serviceAccount = k.core.v1.serviceAccount,

  local this = self,

  _images:: {
    agent: 'grafana/agent:v0.21.1',
    agentctl: 'grafana/agentctl:v0.21.1',
  },
  _config:: {
    name: name,
    namespace: namespace,
    config_hash: true,
    agent_config: '',
  },

  rbac: k.util.rbac(name, [
    policyRule.withApiGroups(['']) +
    policyRule.withResources(['nodes', 'nodes/proxy', 'services', 'endpoints', 'pods']) +
    policyRule.withVerbs(['get', 'list', 'watch']),

    policyRule.withNonResourceUrls('/metrics') +
    policyRule.withVerbs(['get']),
  ]) {
    service_account+: serviceAccount.mixin.metadata.withNamespace(namespace),
  },

  configMap:
    configMap.new(name) +
    configMap.mixin.metadata.withNamespace(namespace) +
    configMap.withData({
      'agent.yaml': k.util.manifestYaml(this._config.agent_config),
    }),

  container::
    container.new(name, this._images.agent) +
    container.withPorts(containerPort.new('http-metrics', 80)) +
    container.withCommand('/bin/agent') +
    container.withArgsMixin(k.util.mapToFlags({
      'config.file': '/etc/agent/agent.yaml',
    })),
}
