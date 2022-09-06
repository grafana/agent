{
  new(name='grafana-agent-operator', namespace='', image='grafana/agent-operator:v0.26.0-rc.0', serviceAccount=''):: {
    local k = (import 'ksonnet-util/kausal.libsonnet'),

    local container = k.core.v1.container,
    local deployment = k.apps.v1.deployment,

    local this = self,

    container::
      container.new(name, image) +
      container.withArgsMixin(k.util.mapToFlags({'-kubelet-service': 'default/kubelet'})),

    controller:
      deployment.new(name, 1, [this.container]) +
      deployment.mixin.metadata.withNamespace(namespace) +
      deployment.mixin.spec.template.spec.withServiceAccount(name),

  },

  withRbac(name, namespace):: {
    local k = (import 'ksonnet-util/kausal.libsonnet') { _config+:: { namespace: namespace } },
    local policyRule = k.rbac.v1.policyRule,
    local serviceAccount = k.core.v1.serviceAccount,
    
    rbac: 
      k.util.rbac(name, [
          policyRule.withApiGroups(['monitoring.grafana.com']) +
          policyRule.withResources(['grafanaagents', 'metricsinstances', 'logsinstances', 'podlogs', 'integrations']) +
          policyRule.withVerbs(['get', 'list', 'watch']),

          policyRule.withApiGroups(['monitoring.grafana.com']) +
          policyRule.withResources(['grafanaagents/finalizers', 'metricsinstances/finalizers', 'logsinstances/finalizers', 'podlogs/finalizers', 'integrations/finalizers']) +
          policyRule.withVerbs(['get', 'list', 'watch', 'update']),

          policyRule.withApiGroups(['monitoring.coreos.com']) +
          policyRule.withResources(['podmonitors', 'probes', 'servicemonitors']) +
          policyRule.withVerbs(['get', 'list', 'watch']),

          policyRule.withApiGroups(['monitoring.coreos.com']) +
          policyRule.withResources(['podmonitors/finalizers', 'probes/finalizers', 'servicemonitors/finalizers']) +
          policyRule.withVerbs(['get', 'list', 'watch', 'update']),

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
  }
}
