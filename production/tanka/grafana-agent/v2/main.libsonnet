local k = import 'ksonnet-util/kausal.libsonnet';
local container = k.core.v1.container;

{
  new(name='grafana-agent', namespace='')::
    (import './internal/base.libsonnet')(name, namespace),

  // Controllers
  withDeploymentController(replicas=1)::
    (import './internal/controllers/deployment.libsonnet')(replicas),
  withDaemonSetController()::
    (import './internal/controllers/daemonset.libsonnet')(),
  withStatefulSetController(replicas=1, volumeClaims=[])::
    (import './internal/controllers/statefulset.libsonnet')(replicas, volumeClaims),

  // Syncer
  newSyncer(name='grafana-agent-syncer', namespace='', config={})::
    (import './internal/syncer.libsonnet')(name, namespace, config),

  // General
  withAgentConfig(config):: { _config+: { agent_config: config } },
  withImagesMixin(images):: { _images+: images },
  withConfigHash(include=true):: { _config+: { config_hash: include } },
  withPortsMixin(ports=[]):: { container+:: container.withPortsMixin(ports) },
  withVolumeMountsMixin(mounts=[]):: { container+:: container.withVolumeMountsMixin(mounts) },
  withVolumesMixin(volumes=[]):: {
    controller+: self.controller.mixin.spec.template.spec.withVolumesMixin(volumes),
  },

  // Helpers
  newKubernetesMetrics(config)::
    (import './internal/helpers/k8s.libsonnet').metrics(config),
  newKubernetesLogs(config)::
    (import './internal/helpers/k8s.libsonnet').logs(config),
  newKubernetesTraces(config)::
    (import './internal/helpers/k8s.libsonnet').traces(config),
  withLogVolumeMounts(config)::
    (import './internal/helpers/logs.libsonnet').volumeMounts(config),
  withLogPermissions(config)::
    (import './internal/helpers/logs.libsonnet').permissions(config),
  withService(config)::
    (import './internal/helpers/service.libsonnet').service(config),
}
