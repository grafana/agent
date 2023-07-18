local k = import 'ksonnet-util/kausal.libsonnet';

local cronJob = k.batch.v1.cronJob;
local configMap = k.core.v1.configMap;
local container = k.core.v1.container;
local deployment = k.apps.v1.deployment;
local volumeMount = k.core.v1.volumeMount;
local volume = k.core.v1.volume;

{
  new(agentctl_image, config):: {
    local this = self,

    local configs = std.foldl(
      function(agg, cfg)
        // Sanitize the name and remove / so every file goes into the same
        // folder.
        local name = std.strReplace(cfg.job_name, '/', '_');

        agg {
          ['%s.yml' % name]: k.util.manifestYaml(
            {
              scrape_configs: [cfg],
              remote_write: config.agent_remote_write,
            },
          ),
        },
      config.kubernetes_scrape_configs,
      {},
    ),

    configMap:
      configMap.new('agent-syncer') +
      configMap.withData(configs),

    container::
      container.new('agent-syncer', agentctl_image) +
      container.withArgsMixin([
        'config-sync',
        '--addr=http://%(agent_pod_name)s.%(namespace)s.svc.cluster.local:80' % config,
        '/etc/configs',
      ]) +
      container.withVolumeMounts([
        volumeMount.new('agent-syncer', '/etc/configs'),
      ]),

    syncer_job:
      cronJob.new('agent-syncer', '*/5 * * * *', this.container) +
      cronJob.mixin.spec.withSuccessfulJobsHistoryLimit(1) +
      cronJob.mixin.spec.withFailedJobsHistoryLimit(3) +
      cronJob.mixin.spec.jobTemplate.spec.template.spec.withRestartPolicy('OnFailure') +
      cronJob.mixin.spec.jobTemplate.spec.template.spec.withActiveDeadlineSeconds(600) +
      cronJob.mixin.spec.jobTemplate.spec.withTtlSecondsAfterFinished(120) +
      cronJob.mixin.spec.jobTemplate.spec.template.spec.withVolumes([
        volume.fromConfigMap(
          name='agent-syncer',
          configMapName=this.configMap.metadata.name,
        ),
      ]),
  },
}
