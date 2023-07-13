function(name='grafana-agent', namespace='') {
  local k = (import 'ksonnet-util/kausal.libsonnet') { _config+:: { namespace: namespace } },

  local container = k.core.v1.container,
  local configMap = k.core.v1.configMap,
  local containerPort = k.core.v1.containerPort,
  local policyRule = k.rbac.v1.policyRule,
  local serviceAccount = k.core.v1.serviceAccount,
  local envVar = k.core.v1.envVar,

  local this = self,

  _images:: {
    agent: 'grafana/agent:v0.35.0-rc.0',
    agentctl: 'grafana/agentctl:v0.35.0-rc.0',
  },
  _config:: {
    name: name,
    namespace: namespace,
    config_hash: true,
    agent_config: '',
    agent_port: 80,
    agent_args: {
      'config.file': '/etc/agent/agent.yaml',
      'server.http.address': '0.0.0.0:80',
      'config.expand-env': 'true',
    },
  },

  rbac: k.util.rbac(name, [
    policyRule.withApiGroups(['']) +
    policyRule.withResources(['nodes', 'nodes/proxy', 'services', 'endpoints', 'pods', 'events']) +
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
    container.withPorts(containerPort.new('http-metrics', this._config.agent_port)) +
    container.withArgsMixin(k.util.mapToFlags(this._config.agent_args)) +
    // `HOSTNAME` is required for promtail (logs) otherwise it will silently do nothing
    container.withEnvMixin([
      envVar.fromFieldPath('HOSTNAME', 'spec.nodeName'),
    ]),
}
