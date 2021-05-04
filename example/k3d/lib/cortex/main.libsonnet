local k = import 'ksonnet-util/kausal.libsonnet';

local configMap = k.core.v1.configMap;
local container = k.core.v1.container;
local containerPort = k.core.v1.containerPort;
local deployment = k.apps.v1.deployment;
local pvc = k.core.v1.persistentVolumeClaim;
local service = k.core.v1.service;
local volumeMount = k.core.v1.volumeMount;
local volume = k.core.v1.volume;

{
  new(namespace=''):: {
    local this = self,

    _images:: {
      cortex: 'cortexproject/cortex:v1.8.1',
    },
    _config:: (import './cortex-config.libsonnet'),

    configMap:
      configMap.new('cortex-config') +
      configMap.mixin.metadata.withNamespace(namespace) +
      configMap.withData({
        'config.yaml': k.util.manifestYaml(this._config),
      }),

    container::
      container.new('cortex', this._images.cortex) +
      container.withPorts([
        containerPort.newNamed(name='http-metrics', containerPort=80),
        containerPort.newNamed(name='grpc', containerPort=9095),
      ]) +
      container.withVolumeMountsMixin(
        volumeMount.new('cortex-data', '/tmp/cortex'),
      ) +
      container.withArgsMixin(
        k.util.mapToFlags({
          'config.file': '/etc/cortex/config.yaml',
        }),
      ),

    pvc:
      { apiVersion: 'v1', kind: 'PersistentVolumeClaim' } +
      pvc.new() +
      pvc.mixin.metadata.withName('cortex-data') +
      pvc.mixin.metadata.withNamespace(namespace) +
      pvc.mixin.spec.withAccessModes('ReadWriteOnce') +
      pvc.mixin.spec.resources.withRequests({ storage: '10Gi' }),

    deployment:
      deployment.new('cortex', 1, [this.container]) +
      deployment.mixin.metadata.withNamespace(namespace) +
      deployment.mixin.spec.template.spec.withVolumesMixin([
        volume.fromPersistentVolumeClaim('cortex-data', 'cortex-data'),
      ]) +
      k.util.configMapVolumeMount(this.configMap, '/etc/cortex') +
      deployment.mixin.spec.template.spec.withTerminationGracePeriodSeconds(4800),

    service:
      k.util.serviceFor(this.deployment) +
      service.mixin.metadata.withNamespace(namespace),
  },
}
