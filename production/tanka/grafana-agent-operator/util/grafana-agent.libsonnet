{
    withRbac(name, namespace):: {
        local k = (import 'ksonnet-util/kausal.libsonnet') + { _config+:: { namespace: namespace } },
        local policyRule = k.rbac.v1.policyRule,
        local serviceAccount = k.core.v1.serviceAccount,
        
        rbac: 
            k.util.rbac(name, [
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
    }
}
