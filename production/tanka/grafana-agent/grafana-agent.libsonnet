local config = import 'config.libsonnet';
local k = import 'ksonnet-util/kausal.libsonnet';

k + config {
  local configMap = $.core.v1.configMap,
  local container = $.core.v1.container,
  local daemonSet = $.apps.v1.daemonSet,
  local deployment = $.core.v1.deployment,
  local policyRule = $.rbac.v1beta1.policyRule,

  agent_rbac:
    $.util.rbac($._config.agent_cluster_role_name, [
      policyRule.new() +
      policyRule.withApiGroups(['']) +
      policyRule.withResources(['nodes', 'nodes/proxy', 'services', 'endpoints', 'pods']) +
      policyRule.withVerbs(['get', 'list', 'watch']),

      policyRule.new() +
      policyRule.withNonResourceUrls('/metrics') +
      policyRule.withVerbs(['get']),
    ]),

  agent_config_map:
    configMap.new($._config.agent_configmap_name) +
    configMap.withData({
      'agent.yml': $.util.manifestYaml($._config.agent_config),
    }),

  agent_args:: {
    'config.file': '/etc/agent/agent.yml',
    'prometheus.wal-directory': '/tmp/agent/data',
  },

  agent_container::
    container.new('agent', $._images.agent) +
    container.withPorts($.core.v1.containerPort.new('http-metrics', 80)) +
    container.withArgsMixin($.util.mapToFlags($.agent_args)) +
    container.withEnv([
      container.envType.fromFieldPath('HOSTNAME', 'spec.nodeName'),
    ]) +
    container.mixin.securityContext.withPrivileged(true) +
    container.mixin.securityContext.withRunAsUser(0),

  local config_hash_mixin =
    if $._config.agent_config_hash_annotation then
      daemonSet.mixin.spec.template.metadata.withAnnotationsMixin({
        config_hash: std.md5(std.toString($._config.agent_config)),
      })
    else {},

  // TODO(rfratto): persistent storage for the WAL here is missing. hostVolume?
  agent_daemonset:
    daemonSet.new($._config.agent_pod_name, [$.agent_container]) +
    daemonSet.mixin.spec.template.spec.withServiceAccount($._config.agent_cluster_role_name) +
    config_hash_mixin +
    $.util.configVolumeMount($._config.agent_configmap_name, '/etc/agent'),


  local agent_deployment_configmap_name = $._config.agent_configmap_name + '_deployment',
  // If running on GKE, you cannot scrape API server pods, and must
  // instead scrape the API server service endpoints. On AKS this doesn't
  // work.
  agent_deployment_config_map:
    if $._config.agent_host_filter then
      configMap.new(agent_deployment_configmap_name) +
      configMap.withData({
        'agent.yml': $.util.manifestYaml($._config.agent_config {
          configs: [{
            name: 'agent',
            host_filter: $._config.agent_host_filter,
            remote_write: $._config.agent_remote_write,
            scrape_configs: [{
              job_name: 'default/kubernetes',
              kubernetes_sd_configs: [{
                role:
                  if $._config.scrape_api_server_endpoints
                  then 'endpoints'
                  else 'service',
              }],
              scheme: 'https',

              tls_config: {
                ca_file: '/var/run/secrets/kubernetes.io/serviceaccount/ca.crt',
                insecure_skip_verify: $._config.prometheus_insecure_skip_verify,
              },
              bearer_token_file: '/var/run/secrets/kubernetes.io/serviceaccount/token',

              relabel_configs: [{
                source_labels: ['__meta_kubernetes_service_label_component'],
                regex: 'apiserver',
                action: 'keep',
              }],

              // Drop some high cardinality metrics.
              metric_relabel_configs: [
                {
                  source_labels: ['__name__'],
                  regex: 'apiserver_admission_controller_admission_latencies_seconds_.*',
                  action: 'drop',
                },
                {
                  source_labels: ['__name__'],
                  regex: 'apiserver_admission_step_admission_latencies_seconds_.*',
                  action: 'drop',
                },
              ],
            }],
          }],
        }),
      })
    else {},

  agent_deployment:
    if $._config.agent_host_filter then
      deployment.new($._config.agent_pod_name, 1, [$.agent_container]) +
      deployment.mixin.spec.template.spec.withServiceAccount($._config.agent_cluster_role_name) +
      deployment.mixin.spec.withReplicas(1) +
      $.util.configVolumeMount(agent_deployment_configmap_name, '/etc/agent')
    else {},
}
