local k = import 'ksonnet-util/kausal.libsonnet';
local policyRule = k.rbac.v1.policyRule;
local serviceAccount = k.core.v1.serviceAccount;
local container = k.core.v1.container;
local deployment = k.apps.v1.deployment;

{
    new(name='grafana-agent-smoke', namespace='default', config={}):: {
        local k = (import 'ksonnet-util/kausal.libsonnet') { _config+:: { namespace: namespace } },

        local this = self,

        _images:: {
            agentsmoke: 'us.gcr.io/kubernetes-dev/grafana/agent-smoke:dev.smoke-framework',
        },

        _config:: {
            mutationFrequency: '5m',
            chaosFrequency: '30m',
            image: this._images.agentsmoke,
            pull_secret: '',
        } + config,

        rbac:
            k.util.rbac(name, [
                policyRule.withApiGroups(['apps']) +
                policyRule.withResources(['deployments/scale']) +
                policyRule.withVerbs(['get', 'update']),
                policyRule.withApiGroups(['']) +
                policyRule.withResources(['pods']) +
                policyRule.withVerbs(['list', 'delete']),
            ]) {
                service_account+:
                  serviceAccount.mixin.metadata.withNamespace(namespace),
            },

        container::
          container.new('agent-smoke', config.image) +
          container.withCommand('/bin/grafana-agent-smoke') +
          container.withArgsMixin(k.util.mapToFlags({
            'log.level': 'debug',
            'namespace': namespace,
            'mutation-frequency': config.mutationFrequency,
            'chaos-frequency': config.chaosFrequency,
          })),

        agentsmoke_deployment:
          deployment.new(name, 1, [self.container]) +
          deployment.mixin.metadata.withNamespace(namespace) +
          deployment.mixin.spec.template.spec.withServiceAccount(name) +
          deployment.spec.template.spec.withImagePullSecrets({ name: this._config.pull_secret }),
    },

    monitoring: (import './prometheus_monitoring.libsonnet'),
}
