function(name='grafana-agent-operator', namespace='') {
    local k = (import 'ksonnet-util/kausal.libsonnet') { _config+:: { namespace: namespace } },

    local container = k.core.v1.container,
    local policyRule = k.rbac.v1.policyRule,
    local serviceAccount = k.core.v1.serviceAccount,
    local deployment = k.apps.v1.deployment,

    local this = self,

    _images:: {
      agent_operator: 'grafana/agent-operator:v0.26.0-rc.0',
    },

    _config:: {
      name: name,
      namespace: namespace,
      operator_args: {
        '-kubelet-service': 'default/kubelet'
      },
    },

    rbac: k.util.rbac(name, [
      policyRule.withApiGroups(['monitoring.grafana.com']) +
      policyRule.withResources(['grafanaagents', 'metricsinstances', 'logsintances', 'podlogs', 'integrations']) +
      policyRule.withVerbs(['get', 'list', 'watch']),

      policyRule.withApiGroups(['monitoring.coreos.com']) +
      policyRule.withResources(['podmonitors', 'probes', 'servicemonitors']) +
      policyRule.withVerbs(['get', 'list', 'watch']),

      policyRule.withApiGroups(['']) +
      policyRule.withResources(['namespaces', 'nodes']) +
      policyRule.withVerbs(['get', 'list', 'watch']),

      policyRule.withApiGroups(['']) +
      policyRule.withResources(['secrets', 'services', 'configmaps', 'endpoints']) +
      policyRule.withVerbs(['get', 'list', 'watch', 'create', 'update', 'patch', 'delete']),

      policyRule.withApiGroups(['apps']) +
      policyRule.withResources(['statefulsets', 'daemonsets', 'deployments']) +
      policyRule.withVerbs(['get', 'list', 'watch', 'create', 'update', 'patch', 'delete']),
      
    ]) {
      service_account+: serviceAccount.mixin.metadata.withNamespace(namespace),
    },

    container::
      container.new(name, this._images.agent_operator) +
      container.withArgsMixin(k.util.mapToFlags(this._config.operator_args)),

    controller:
      deployment.new(name, 1, [this.container]) +
      deployment.mixin.metadata.withNamespace(namespace) +
      deployment.mixin.spec.template.spec.withServiceAccount(name),

}
