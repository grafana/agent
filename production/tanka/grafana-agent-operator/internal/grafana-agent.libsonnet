function(name='grafana-agent', namespace='') {
    local this = self,
    local k = (import 'ksonnet-util/kausal.libsonnet') { _config+:: { namespace: namespace } },
    local gen = import 'agent-operator-gen/main.libsonnet',
    
    local ga = gen.monitoring.v1alpha1.grafanaAgent,

    local policyRule = k.rbac.v1.policyRule,
    local serviceAccount = k.core.v1.serviceAccount,

    _images:: {
      agent: 'grafana/agent:v0.26.0-rc.0',
    },

    rbac: k.util.rbac(name, [
        policyRule.withApiGroups(['']) +
        policyRule.withResources(['nodes', 'nodes/proxy', 'nodes/metrics', 'services', 'endpoints', 'pods', 'events']) +
        policyRule.withVerbs(['get', 'list', 'watch']),

        policyRule.withApiGroups(['networking.k8s.io']) +
        policyRule.withResources(['ingresses']) +
        policyRule.withVerbs(['get', 'list', 'watch']),

        policyRule.withNonResourceURLs(['/metrics', '/metrics/cadvisor']) +
        policyRule.withVerbs(['get']),
    ]) {
        service_account+: serviceAccount.mixin.metadata.withNamespace(namespace),
    },

    resource: ga.new(name) +
        ga.metadata.withNamespace(namespace) +
        ga.spec.withServiceAccountName(name) +
        ga.spec.withImage(this._images.agent)
}
