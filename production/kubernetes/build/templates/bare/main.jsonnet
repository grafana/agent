local agent = import 'grafana-agent/v2/main.libsonnet';
local k = import 'ksonnet-util/kausal.libsonnet';

local pvc = k.core.v1.persistentVolumeClaim;
local volumeMount = k.core.v1.volumeMount;
local containerPort = k.core.v1.containerPort;

{
  agent:
    agent.new(name='grafana-agent', namespace='${NAMESPACE}') +
    agent.withStatefulSetController(
      replicas=1,
      volumeClaims=[
        pvc.new() +
        pvc.mixin.metadata.withName('agent-wal') +
        pvc.mixin.metadata.withNamespace('${NAMESPACE}') +
        pvc.mixin.spec.withAccessModes('ReadWriteOnce') +
        pvc.mixin.spec.resources.withRequests({ storage: '5Gi' }),
      ],
    ) +
    agent.withConfigHash(false) +
    agent.withArgsMixin({
      'enable-features': 'integrations-next'
    },) +
    // add dummy config or else will fail
    agent.withAgentConfig({
      server: { log_level: 'error' },
    }) +
    agent.withVolumeMountsMixin([volumeMount.new('agent-wal', '/var/lib/agent')]) +
    // headless svc needed by statefulset
    agent.withService() +
    {
      controller_service+: {
        spec+: {
          clusterIP: 'None',
        },
      },
    } +
    // hack to disable ConfigMap
    { configMap:: super.configMap },
}
