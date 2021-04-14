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
      collector: 'otel/opentelemetry-collector:0.9.0',
    },
    _config:: (import './collector-config.libsonnet'),

    configMap:
      configMap.new('collector') +
      configMap.mixin.metadata.withNamespace(namespace) +
      configMap.withData({
        'config.yaml': k.util.manifestYaml(this._config),
      }),

    container::
      container.new('collector', this._images.collector) +
      container.withPorts([
        containerPort.newNamed(name='grpc', containerPort=55680),
      ]) +
      container.withVolumeMountsMixin(
        volumeMount.new('collector-data', '/tmp/collector'),
      ) +
      container.withArgsMixin(
          '--config=/etc/collector/config.yaml',
      ),

    pvc:
      { apiVersion: 'v1', kind: 'PersistentVolumeClaim' } +
      pvc.new() +
      pvc.mixin.metadata.withName('collector-data') +
      pvc.mixin.metadata.withNamespace(namespace) +
      pvc.mixin.spec.withAccessModes('ReadWriteOnce') +
      pvc.mixin.spec.resources.withRequests({ storage: '10Gi' }),

    deployment:
      deployment.new('collector', 1, [self.container]) +
      deployment.mixin.metadata.withNamespace(namespace) +
      deployment.mixin.spec.template.spec.withVolumesMixin([
        volume.fromPersistentVolumeClaim('collector-data', 'collector-data'),
      ]) +
      k.util.configMapVolumeMount(this.configMap, '/etc/collector'),


    service:
      k.util.serviceFor(self.deployment) +
      service.mixin.metadata.withNamespace(namespace),
  },
}
