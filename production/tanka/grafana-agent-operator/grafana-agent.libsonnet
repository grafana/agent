
local gen = import 'agent-operator-gen/main.libsonnet';
local ga = gen.monitoring.v1alpha1.grafanaAgent;

{
    local this = self,

    _images:: {
        agent: 'grafana/agent:v0.26.0-rc.0',
    },

    new(name='grafana-agent', namespace=''):: {
        local k = (import 'ksonnet-util/kausal.libsonnet') { _config+:: { namespace: namespace } },

        local policyRule = k.rbac.v1.policyRule,
        local serviceAccount = k.core.v1.serviceAccount,
        
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
    },

    withMetrics(metricsInstanceLabels):: {
        resource+:
            ga.spec.metrics.instanceSelector.withMatchLabels(metricsInstanceLabels)
    },

    withLogs(logsInstanceLabels):: {
        resource+:
            ga.spec.logs.instanceSelector.withMatchLabels(logsInstanceLabels)
    },

    withIntegration(integrationLabels):: {
        resource+:
            ga.spec.integrations.selector.withMatchLabels(integrationLabels)
    },

    withMetricsExternalLabels(externalLabels):: {
        resource+:
            ga.spec.metrics.withExternalLabels(externalLabels)
    },
}
