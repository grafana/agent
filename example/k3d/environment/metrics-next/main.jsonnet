local default = import 'default/main.libsonnet';
local gragent = import 'grafana-agent/v2/main.libsonnet';
local k = import 'ksonnet-util/kausal.libsonnet';

local ingress = k.networking.v1beta1.ingress;
local pvc = k.core.v1.persistentVolumeClaim;
local path = k.networking.v1beta1.httpIngressPath;
local rule = k.networking.v1beta1.ingressRule;
local container = k.core.v1.container;
local volumeMount = k.core.v1.volumeMount;

local images = {
  agent: 'grafana/agent:latest',
  agentctl: 'grafana/agentctl:latest',
};

{
  default: default.new(namespace='default') {
    grafana+: {
      ingress+:
        ingress.new('grafana-ingress') +
        ingress.mixin.spec.withRules([
          rule.withHost('grafana.k3d.localhost') +
          rule.http.withPaths([
            path.withPath('/')
            + path.backend.withServiceName('grafana')
            + path.backend.withServicePort(80),
          ]),
        ]),
    },
  },

  logging_agents:
    gragent.new(name='grafana-agent-logs', namespace='default') +
    gragent.withImagesMixin(images) +
    gragent.withDaemonSetController() +
    gragent.withLogVolumeMounts() +
    gragent.withLogPermissions() +
    gragent.withAgentConfig({
      server: { log_level: 'debug' },

      logs: {
        positions_directory: '/tmp/loki-positions',
        configs: [{
          name: 'default',
          clients: [{
            url: 'http://loki.default.svc.cluster.local/loki/api/v1/push',
            external_labels: { cluster: 'k3d' },
          }],
          scrape_configs: [
            x { pipeline_stages: [{ cri: {} }] }
            for x
            in gragent.newKubernetesLogs()
          ],
        }],
      },
    }),


  agent_cluster:
    gragent.new(name='grafana-agent', namespace='default') {
      container+:: container.withArgsMixin(k.util.mapToFlags({
        'cluster.enable': 'true',
        'cluster.discover-peers': 'provider=k8s namespace=default label_selector="name=grafana-agent"',

        'enable-features': 'metrics-next',
        'metrics.wal.directory': '/var/lib/agent',
      })),
    } +
    gragent.withImagesMixin(images) +
    gragent.withStatefulSetController(
      replicas=3,
      volumeClaims=[
        pvc.new() +
        pvc.mixin.metadata.withName('agent-wal') +
        pvc.mixin.metadata.withNamespace('smoke') +
        pvc.mixin.spec.withAccessModes('ReadWriteOnce') +
        pvc.mixin.spec.resources.withRequests({ storage: '5Gi' }),
      ],
    ) +
    gragent.withVolumeMountsMixin([volumeMount.new('agent-wal', '/var/lib/agent')]) +
    gragent.withAgentConfig({
      server: { log_level: 'debug' },

      metrics: {
        global: {
          scrape_interval: '15s',
          external_labels: { cluster: 'k3d' },
        },
        configs: [{
          name: 'default',
          scrape_configs: gragent.newKubernetesMetrics({}),
          remote_write: [{
            url: 'http://cortex.default.svc.cluster.local/api/prom/push',
          }],
        }],
      },
    }),
}
