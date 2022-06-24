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
  new(name='vulture', namespace='', config={}):: {
    local this = self,
    local port = 8080,

    _config+:: {
      image: 'grafana/tempo-vulture:latest',
      replicas: 1,
      tempoPushUrl: 'http://grafana-agent',
      tempoQueryUrl: 'http://tempo:3200',
      tempoOrgId: '',
      tempoRetentionDuration: '336h',
      tempoSearchBackoffDuration: '0s', // disable search
      tempoReadBackoffDuration: '10s',
      tempoWriteBackoffDuration: '10s',
    } + config,

    container::
      container.new(name, this._config.image) +
      container.withPorts([
        containerPort.newNamed(name='prom-metrics', containerPort=port),
      ]) +
      container.withArgs([
        '-prometheus-listen-address=:' + port,
        '-tempo-push-url=' + this._config.tempoPushUrl,
        '-tempo-query-url=' + this._config.tempoQueryUrl,
        '-tempo-org-id=' + this._config.tempoOrgId,
        '-tempo-retention-duration=' + this._config.tempoRetentionDuration,
        '-tempo-search-backoff-duration=' + this._config.tempoSearchBackoffDuration,
        '-tempo-read-backoff-duration=' + this._config.tempoReadBackoffDuration,
        '-tempo-write-backoff-duration=' + this._config.tempoWriteBackoffDuration,
      ]) +
      k.util.resourcesRequests('50m', '100Mi') +
      k.util.resourcesLimits('100m', '500Mi'),

    deployment:
      deployment.new(name, this._config.replicas, [self.container], {app: name}) +
      deployment.mixin.metadata.withNamespace(namespace),
  },
}
