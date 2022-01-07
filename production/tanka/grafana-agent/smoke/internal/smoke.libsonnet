local k = import 'ksonnet-util/kausal.libsonnet';

local policyRule = k.rbac.v1.policyRule;
local serviceAccount = k.core.v1.serviceAccount;
local container = k.core.v1.container;
local deployment = k.apps.v1.deployment;

{
    newSmoke(name='grafana-agent-smoke', namespace='default', image):: {
        local k = (import 'ksonnet-util/kausal.libsonnet') { _config+:: { namespace: namespace } },

        rbac:
            k.util.rbac(name, [
                policyRule.withApiGroups(['apps']) +
                policyRule.withResources(['deployments/scale']) +
                policyRule.withVerbs(['get', 'update']),
            ]) {
                service_account+:
                  serviceAccount.mixin.metadata.withNamespace(namespace),
            },

        container::
          container.new('agent-smoke', image) +
          container.withCommand('/bin/grafana-agent-smoke') +
          container.withArgsMixin(k.util.mapToFlags({
            'log.level': 'debug',
            'namespace': namespace,
          })),

        agentsmoke:
          deployment.new(name, 1, [self.container]) +
          deployment.mixin.metadata.withNamespace(namespace) +
          deployment.mixin.spec.template.spec.withServiceAccount(name),
    },
}
