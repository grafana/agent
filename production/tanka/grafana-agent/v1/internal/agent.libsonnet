local k = import 'ksonnet-util/kausal.libsonnet';

local configMap = k.core.v1.configMap;
local container = k.core.v1.container;
local daemonSet = k.apps.v1.daemonSet;
local deployment = k.apps.v1.deployment;
local policyRule = k.rbac.v1.policyRule;

{
  newAgent(name='grafana-agent', namespace='default', image, config, use_daemonset=true):: {
    local controller = if use_daemonset then daemonSet else deployment,
    local k = (import 'ksonnet-util/kausal.libsonnet') { _config+:: { namespace: namespace } },

    _controller:: controller,
    _config_hash:: true,

    rbac:
      k.util.rbac(name, [
        // Need for k8s SD on Loki/Prometheus subsystems
        policyRule.withApiGroups(['']) +
        policyRule.withResources(['nodes', 'nodes/proxy', 'services', 'endpoints', 'pods']) +
        policyRule.withVerbs(['get', 'list', 'watch']),

        // Needed for Prometheus subsystem to scrape k8s API
        policyRule.withNonResourceUrls('/metrics') +
        policyRule.withVerbs(['get']),
      ]),

    config_map:
      configMap.new(name) +
      configMap.mixin.metadata.withNamespace(namespace) +
      configMap.withData({
        'agent.yaml': k.util.manifestYaml(config),
      }),

    container::
      container.new('agent', image) +
      container.withPorts(k.core.v1.containerPort.new('http-metrics', 8080)) +
      container.withArgsMixin(k.util.mapToFlags({
        'config.file': '/etc/agent/agent.yaml',
      })),

    agent:
      (
        if use_daemonset then daemonSet.new(name, [self.container])
        else deployment.new(name, 1, [self.container])
      ) +
      controller.mixin.metadata.withNamespace(namespace) +
      controller.mixin.spec.template.spec.withServiceAccount(name) +
      (
        if self._config_hash
        then controller.mixin.spec.template.metadata.withAnnotationsMixin({
          config_hash: std.md5(std.toString(config)),
        })
        else {}
      ) +
      k.util.configVolumeMount(name, '/etc/agent'),
  },

  withConfigHash(include):: { _config_hash:: include },
  withEnv(env):: {
    container+:: container.withEnv(env),
  },
}
