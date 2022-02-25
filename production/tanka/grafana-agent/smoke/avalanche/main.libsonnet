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
  new(name='avalanche', replicas=1, namespace='', config={}):: {
    local this = self,

    _config+:: {
      image: 'quay.io/freshtracks.io/avalanche:latest',

      metric_count: 500,
      label_count: 10,
      series_count: 10,
      metricname_length: 5,
      labelname_length: 5,
      value_interval: 30,
      series_interval: 30,
      metric_interval: 120,
    } + config,

    container::
      container.new(name, this._config.image) +
      container.withPorts([
        containerPort.newNamed(name='http', containerPort=9001),
      ]) +
      container.withArgsMixin([
        '--metric-count=%d' % this._config.metric_count,
        '--label-count=%d' % this._config.label_count,
        '--series-count=%d' % this._config.series_count,
        '--metricname-length=%d' % this._config.metricname_length,
        '--labelname-length=%d' % this._config.labelname_length,
        '--value-interval=%d' % this._config.value_interval,
        '--series-interval=%d' % this._config.series_interval,
        '--metric-interval=%d' % this._config.metric_interval,
      ]),

    deployment:
      deployment.new(name, replicas, [self.container]) +
      deployment.mixin.metadata.withNamespace(namespace),
  },
}
