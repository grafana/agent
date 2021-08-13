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
  new(name='test-exporter', namespace='', config={}):: {
    local this = self,

    _config+:: {
      image: 'cortexproject/test-exporter:v1.9.0',
      args: {},
    } + config,

    container::
      container.new(name, this._config.image) +
      container.withPorts([
        containerPort.newNamed(name='http-metrics', containerPort=80),
      ]) +
      container.withArgsMixin(k.util.mapToFlags(this._config.args)),

    deployment:
      deployment.new(name, 1, [self.container]) +
      deployment.mixin.metadata.withNamespace(namespace),
  },
}
