local k = import 'ksonnet-util/kausal.libsonnet';
local policyRule = k.rbac.v1.policyRule;
local serviceAccount = k.core.v1.serviceAccount;
local container = k.core.v1.container;
local containerPort = k.core.v1.containerPort;
local deployment = k.apps.v1.deployment;
local service = k.core.v1.service;
local util = k.util;

{
  new(name='grafana-agent-smoke', namespace='default', config={}):: {
    local k = (import 'ksonnet-util/kausal.libsonnet') { _config+:: { namespace: namespace } },

    local this = self,

    _images:: {
      agentsmoke: 'us.gcr.io/kubernetes-dev/grafana/agent-smoke:main',
    },

    _config:: {
      mutationFrequency: '5m',
      chaosFrequency: '30m',
      image: this._images.agentsmoke,
      pull_secret: '',
      podPrefix: 'grafana-agent',
      simulateErrors: true,
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
      container.new('agent-smoke', this._config.image) +
      container.withPorts([
        containerPort.newNamed(name='remote-write', containerPort=19090),
      ]) +
      container.withArgsMixin(k.util.mapToFlags({
        'log.level': 'debug',
        namespace: namespace,
        'mutation-frequency': this._config.mutationFrequency,
        'chaos-frequency': this._config.chaosFrequency,
        'pod-prefix': this._config.podPrefix,
        'fake-remote-write': true,
        'simulate-errors': this._config.simulateErrors,
      })),

    agentsmoke_deployment:
      deployment.new(name, 1, [self.container]) +
      deployment.mixin.metadata.withNamespace(namespace) +
      deployment.mixin.spec.template.spec.withServiceAccount(name) +
      deployment.spec.template.spec.withImagePullSecrets({ name: this._config.pull_secret }),

    service:
      util.serviceFor(self.agentsmoke_deployment) +
      service.mixin.metadata.withNamespace(namespace),
  },

  monitoring: (import './prometheus_monitoring.libsonnet'),
}
