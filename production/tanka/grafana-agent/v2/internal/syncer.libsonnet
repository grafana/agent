local k = import 'ksonnet-util/kausal.libsonnet';

local cronJob = k.batch.v1beta1.cronJob;
local configMap = k.core.v1.configMap;
local container = k.core.v1.container;
local deployment = k.apps.v1.deployment;
local volumeMount = k.core.v1.volumeMount;
local volume = k.core.v1.volume;

function(
  name='grafana-agent-syncer',
  namespace='',
  config={},
) {
  local _config = {
    api: error 'api must be set',
    image: 'grafana/agentctl:v0.33.0',
    schedule: '*/5 * * * *',
    configs: [],
  } + config,

  local this = self,
  local _configs = std.foldl(
    function(agg, cfg)
      // Sanitize the name and remove / so every file goes into the same
      // folder.
      local name = std.strReplace(cfg.name, '/', '_');

      agg { ['%s.yml' % name]: k.util.manifestYaml(cfg) },
    _config.configs,
    {},
  ),

  configMap:
    configMap.new(name) +
    configMap.mixin.metadata.withNamespace(namespace) +
    configMap.withData(_configs),

  container::
    container.new(name, _config.image) +
    container.withArgsMixin([
      'config-sync',
      '--addr=%s' % _config.api,
      '/etc/configs',
    ]) +
    container.withVolumeMounts(volumeMount.new(name, '/etc/configs')),

  job:
    cronJob.new(name, _config.schedule, this.container) +
    cronJob.mixin.metadata.withNamespace(namespace) +
    cronJob.mixin.spec.withSuccessfulJobsHistoryLimit(1) +
    cronJob.mixin.spec.withFailedJobsHistoryLimit(3) +
    cronJob.mixin.spec.jobTemplate.spec.template.spec.withRestartPolicy('OnFailure') +
    cronJob.mixin.spec.jobTemplate.spec.template.spec.withActiveDeadlineSeconds(600) +
    cronJob.mixin.spec.jobTemplate.spec.withTtlSecondsAfterFinished(120) +
    cronJob.mixin.spec.jobTemplate.spec.template.spec.withVolumes([
      volume.fromConfigMap(
        name=name,
        configMapName=this.configMap.metadata.name,
      ),
    ]),
}
