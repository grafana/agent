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
      load_generator: 'omnition/synthetic-load-generator:1.0.25',
    },
    _config:: (import './load-generator-config.json'),

    configMap:
      configMap.new('load-generator') +
      configMap.mixin.metadata.withNamespace(namespace) +
      configMap.withData({
        'config.json': std.toString(this._config),
      }),

    container::
      container.new('load-generator', this._images.load_generator) +
      container.withPorts([
        containerPort.newNamed(name='grpc', containerPort=55680),
      ]) +
      container.withEnvMixin([
        {
          name: 'TOPOLOGY_FILE',
          value: '/etc/load-generator/config.json',
        },
        {
          name: 'JAEGER_COLLECTOR_URL',
          value: 'http://grafana-agent.default.svc.cluster.local:14268',
        },
      ]),

    deployment:
      deployment.new('load-generator', 1, [self.container]) +
      deployment.mixin.metadata.withNamespace(namespace) +
      k.util.configMapVolumeMount(this.configMap, '/etc/load-generator'),


    service:
      k.util.serviceFor(self.deployment) +
      service.mixin.metadata.withNamespace(namespace),
  },
}
