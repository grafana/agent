local config = import '../config.libsonnet';
local syncer = import './syncer.libsonnet';
local k = import 'ksonnet-util/kausal.libsonnet';

local containerPort = k.core.v1.containerPort;
local configMap = k.core.v1.configMap;
local container = k.core.v1.container;
local deployment = k.apps.v1.deployment;
local policyRule = k.rbac.v1.policyRule;

{
  new(namespace='default', kube_namespace='kube-system'):: config {
    local this = self,

    // Use the default config from the non-scraping-service mode
    // but change some of the defaults.
    _config+:: {
      agent_cluster_role_name: 'grafana-agent-cluster',
      agent_configmap_name: 'grafana-agent-cluster',
      agent_pod_name: 'grafana-agent-cluster',
      agent_replicas: 3,

      namespace: namespace,
      kube_namespace: kube_namespace,

      // Scraping service should not be using host filtering
      agent_host_filter: false,

      //
      // KVStore options
      //
      agent_config_kvstore: error 'must configure config kvstore',
      agent_ring_kvstore: error 'must configure ring kvstore',

      agent_config+: {
        metrics+: {
          // No configs are used in the scraping service mode.
          configs:: [],

          scraping_service: {
            enabled: true,
            kvstore: this._config.agent_config_kvstore,
            lifecycler: {
              ring: {
                kvstore: this._config.agent_ring_kvstore,
              },
            },
          },
        },
      },
    },

    rbac:
      // Need to do a hack here so ksonnet util has our configs :(
      (k { _config+: this._config }).util.rbac(this._config.agent_cluster_role_name, [
        policyRule.withApiGroups(['']) +
        policyRule.withResources(['nodes', 'nodes/proxy', 'services', 'endpoints', 'pods']) +
        policyRule.withVerbs(['get', 'list', 'watch']),

        policyRule.withNonResourceUrls('/metrics') +
        policyRule.withVerbs(['get']),
      ]),

    configMap:
      configMap.new(this._config.agent_configmap_name) +
      configMap.withData({
        'agent.yml': k.util.manifestYaml(this._config.agent_config),
      }),

    container::
      container.new('agent-cluster', this._images.agent) +
      container.withPorts(containerPort.new(name='http-metrics', port=80)) +
      container.withArgsMixin(k.util.mapToFlags({
        'config.file': '/etc/agent/agent.yml',
        'metrics.wal-directory': '/tmp/agent/data',
      })) +
      container.withEnv([
        k.core.v1.envVar.fromFieldPath('HOSTNAME', 'spec.nodeName'),
      ]) +
      container.mixin.securityContext.withPrivileged(true) +
      container.mixin.securityContext.withRunAsUser(0),

    deployment:
      deployment.new(this._config.agent_pod_name, this._config.agent_replicas, [this.container]) +
      deployment.mixin.spec.template.spec.withServiceAccount(this._config.agent_cluster_role_name) +
      deployment.mixin.spec.withMinReadySeconds(60) +
      deployment.mixin.spec.strategy.rollingUpdate.withMaxSurge(0) +
      deployment.mixin.spec.strategy.rollingUpdate.withMaxUnavailable(1) +
      deployment.mixin.spec.template.spec.withTerminationGracePeriodSeconds(4800) +
      k.util.configVolumeMount(this._config.agent_configmap_name, '/etc/agent'),

    service:
      k.util.serviceFor(this.deployment),

    // Create the cronjob that syncs configs to the API
    syncer:
      syncer.new(this._images.agentctl, this._config),
  },

  withImagesMixin(images):: { _images+: images },

  // withConfig overrides the config used for the agent.
  withConfig(config):: { _config: config },

  // withConfigMixin merges the provided config with the existing config.
  withConfigMixin(config):: { _config+: config },
}
