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
      container.withArgsMixin(
        '--config=/etc/collector/config.yaml',
      ),

    deployment:
      deployment.new('collector', 1, [self.container]) +
      deployment.mixin.metadata.withNamespace(namespace) +
      k.util.configMapVolumeMount(this.configMap, '/etc/collector'),


    service:
      k.util.serviceFor(self.deployment) +
      service.mixin.metadata.withNamespace(namespace),
  },
}
