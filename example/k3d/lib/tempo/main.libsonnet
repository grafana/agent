(import 'configmap.libsonnet') +
(import 'config.libsonnet') +

local k = import 'ksonnet-util/kausal.libsonnet';

local configMap = k.core.v1.configMap;
local container = k.core.v1.container;
local containerPort = k.core.v1.containerPort;
local deployment = k.apps.v1.deployment;
local pvc = k.core.v1.persistentVolumeClaim;
local service = k.core.v1.service;
local statefulset = k.apps.v1.statefulSet;
local volumeMount = k.core.v1.volumeMount;
local volume = k.core.v1.volume;

{
  new(namespace=''):: {
    local this = self,

    local tempo_config_volume = 'tempo-conf',
    local tempo_query_config_volume = 'tempo-query-conf',
    local tempo_data_volume = 'tempo-data',

    configmap:
      configMap.new('tempo') +
      configMap.withData({
        'tempo.yaml': k.util.manifestYaml($.tempo_config),
      }) +
      configMap.withDataMixin({
        'overrides.yaml': |||
          overrides:
        |||,
      }) + 
      configMap.mixin.metadata.withNamespace(namespace),

    pvc::
      pvc.new() +
      pvc.mixin.spec.resources
      .withRequests({ storage: $._config.pvc_size }) +
      pvc.mixin.spec
      .withAccessModes(['ReadWriteOnce']) +
      pvc.mixin.spec
      .withStorageClassName($._config.pvc_storage_class) +
      pvc.mixin.metadata
      .withLabels({ app: 'tempo' }) +
      pvc.mixin.metadata
      .withNamespace(namespace) +
      pvc.mixin.metadata
      .withName(tempo_data_volume) +
      { kind: 'PersistentVolumeClaim', apiVersion: 'v1' },

    container::
      container.new('tempo', $._images.tempo) +
      container.withPorts([
        containerPort.new('prom-metrics', $._config.tempo.port),
        containerPort.new('memberlist', 9095),
        containerPort.new('otlp', 4317),
      ]) +
      container.withArgs([
        '-target=scalable-single-binary',
        '-config.file=/conf/tempo.yaml',
        '-mem-ballast-size-mbs=' + $._config.ballast_size_mbs,
      ]) +
      container.withVolumeMounts([
        volumeMount.new(tempo_config_volume, '/conf'),
        volumeMount.new(tempo_data_volume, '/var/tempo'),
      ]) +
      k.util.resourcesRequests('3', '3Gi') +
      k.util.resourcesLimits('5', '5Gi'),

    statefulset:
      statefulset.new('tempo',
                      $._config.tempo.replicas,
                      [
                        self.container,
                        self.query_container,
                      ],
                      self.pvc,
                      { app: 'tempo' }) +
      statefulset.mixin.spec.withServiceName('tempo') +
      statefulset.mixin.spec.template.metadata.withAnnotations({
        config_hash: std.md5(std.toString(this.configmap.data['tempo.yaml'])),
      }) +
      statefulset.mixin.metadata.withLabels({ app: $._config.tempo.headless_service_name, name: 'tempo' }) +
      statefulset.mixin.spec.selector.withMatchLabels({ name: 'tempo' }) +
      statefulset.mixin.spec.template.metadata.withLabels({ name: 'tempo', app: $._config.tempo.headless_service_name }) +
      statefulset.mixin.spec.template.spec.withVolumes([
        volume.fromConfigMap(tempo_query_config_volume, this.query_configmap.metadata.name),
        volume.fromConfigMap(tempo_config_volume, this.configmap.metadata.name),
      ]) +
      statefulset.mixin.metadata.withNamespace(namespace),

  service:
    k.util.serviceFor(this.statefulset) +
    service.mixin.metadata.withNamespace(namespace),

    headless_service:
      service.new(
        $._config.tempo.headless_service_name,
        { app: $._config.tempo.headless_service_name },
        []
      ) +
      service.mixin.spec.withClusterIP('None') +
      service.mixin.spec.withPublishNotReadyAddresses(true)  +
      service.mixin.metadata.withNamespace(namespace),
  }
}
