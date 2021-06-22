local k = import 'ksonnet-util/kausal.libsonnet';

local configMap = k.core.v1.configMap;
local container = k.core.v1.container;
local containerPort = k.core.v1.containerPort;
local deployment = k.apps.v1.deployment;
local statefulSet = k.apps.v1.statefulSet;
local pvc = k.core.v1.persistentVolumeClaim;
local service = k.core.v1.service;
local volumeMount = k.core.v1.volumeMount;
local volume = k.core.v1.volume;
local serviceAccount = k.core.v1.serviceAccount;
local policyRule = k.rbac.v1.policyRule;

{
  new(namespace=''):: {
    local k = (import 'ksonnet-util/kausal.libsonnet') { _config+:: { namespace: namespace } },
    local this = self,

    _images:: {
      prom: 'prom/prometheus:v2.28.0',
    },

    _config:: {
      rule_files: ['/etc/prometheus/rules.yaml'],
    },
    _rules:: {},

    rbac:
      k.util.rbac('prometheus', [
        policyRule.withApiGroups(['']) +
        policyRule.withResources(['nodes', 'nodes/proxy', 'services', 'endpoints', 'pods']) +
        policyRule.withVerbs(['get', 'list', 'watch']),

        policyRule.withNonResourceUrls('/metrics') +
        policyRule.withVerbs(['get']),
      ]) {
        service_account+:
          serviceAccount.mixin.metadata.withNamespace(namespace),
      },

    configMap:
      configMap.new('prometheus') +
      configMap.mixin.metadata.withNamespace(namespace) +
      configMap.withData({
        'prometheus.yaml': k.util.manifestYaml(this._config),
        'rules.yaml': k.util.manifestYaml(this._rules),
      }),

    container::
      container.new('prometheus', this._images.prom) +
      container.withPorts([
        containerPort.newNamed(name='http-metrics', containerPort=9090),
      ]) +
      container.withVolumeMountsMixin(
        volumeMount.new('prometheus-data', '/data'),
      ) +
      container.withArgsMixin([
        '--config.file=/etc/prometheus/prometheus.yaml',
        '--storage.tsdb.path=/data',
      ]),

    pvc::
      { apiVersion: 'v1', kind: 'PersistentVolumeClaim' } +
      pvc.new() +
      pvc.mixin.metadata.withName('prometheus-data') +
      pvc.mixin.metadata.withNamespace(namespace) +
      pvc.mixin.spec.withAccessModes('ReadWriteOnce') +
      pvc.mixin.spec.resources.withRequests({ storage: '10Gi' }),

    statefulSet:
      statefulSet.new(
        name='prometheus',
        replicas=1,
        containers=[this.container],
        volumeClaims=[this.pvc]
      ) +
      statefulSet.mixin.spec.withServiceName('prometheus') +
      k.util.configMapVolumeMount(this.configMap, '/etc/prometheus') +
      statefulSet.mixin.spec.template.spec.withServiceAccountName('prometheus') +
      statefulSet.mixin.metadata.withNamespace(namespace),

    service:
      k.util.serviceFor(this.statefulSet) +
      service.mixin.metadata.withNamespace(namespace),
  },

  withConfigMixin(config={}):: { _config+:: config },
  withRulesMixin(rules={}):: { _rules+:: rules },
}
