local monitoring = import './monitoring/main.jsonnet';
local cortex = import 'cortex/main.libsonnet';
local tempo = import 'tempo/main.libsonnet';
local avalanche = import 'grafana-agent/smoke/avalanche/main.libsonnet';
local crow = import 'grafana-agent/smoke/crow/main.libsonnet';
local vulture = import 'grafana-agent/smoke/vulture/main.libsonnet';
local etcd = import 'grafana-agent/smoke/etcd/main.libsonnet';
local smoke = import 'grafana-agent/smoke/main.libsonnet';
local gragent = import 'grafana-agent/v2/main.libsonnet';
local k = import 'ksonnet-util/kausal.libsonnet';

local namespace = k.core.v1.namespace;
local pvc = k.core.v1.persistentVolumeClaim;
local volumeMount = k.core.v1.volumeMount;

local images = {
  agent: 'grafana/agent:main',
  agentctl: 'grafana/agentctl:main',
};

local new_crow(name, selector) =
  crow.new(name, namespace='smoke', config={
    args+: {
      'crow.prometheus-addr': 'http://cortex/api/prom',
      'crow.extra-selectors': selector,
    },
  });

local new_smoke(name) = smoke.new(name, namespace='smoke', config={
  mutationFrequency: '5m',
  chaosFrequency: '30m',
});


local smoke = {
  ns: namespace.new('smoke'),

  cortex: cortex.new('smoke'),

  tempo: tempo.new('smoke'),

  // Needed to run agent cluster
  etcd: etcd.new('smoke'),

  avalanche: avalanche.new(replicas=3, namespace='smoke', config={
    // We're going to be running a lot of these and we're not trying to test
    // for load, so reduce the cardinality and churn rate.
    metric_count: 1000,
    series_interval: 300,
    metric_interval: 600,
  }),

  smoke_test: new_smoke('smoke-test'),

  crows: [
    new_crow('crow-single', 'cluster="grafana-agent"'),
    new_crow('crow-cluster', 'cluster="grafana-agent-cluster"'),
  ],

  vulture: vulture.new(namespace='smoke'),

  local metric_instances(crow_name) = [{
    name: 'crow',
    remote_write: [
      {
        url: 'http://cortex/api/prom/push',
        write_relabel_configs: [
          {
            source_labels: ['__name__'],
            regex: 'avalanche_.*',
            action: 'drop',
          },
        ],
      },
      {
        url: 'http://smoke-test:19090/api/prom/push',
        write_relabel_configs: [
          {
            source_labels: ['__name__'],
            regex: 'avalanche_.*',
            action: 'keep',
          },
        ],
      },
    ],
    scrape_configs: [
      {
        job_name: 'crow',
        metrics_path: '/validate',

        kubernetes_sd_configs: [{ role: 'pod' }],
        tls_config: {
          ca_file: '/var/run/secrets/kubernetes.io/serviceaccount/ca.crt',
        },
        bearer_token_file: '/var/run/secrets/kubernetes.io/serviceaccount/token',

        relabel_configs: [{
          source_labels: ['__meta_kubernetes_namespace'],
          regex: 'smoke',
          action: 'keep',
        }, {
          source_labels: ['__meta_kubernetes_pod_container_name'],
          regex: crow_name,
          action: 'keep',
        }],
      },
    ],
  }, {
    name: 'avalanche',
    remote_write: [
      {
        url: 'http://cortex/api/prom/push',
        write_relabel_configs: [
          {
            source_labels: ['__name__'],
            regex: 'avalanche_.*',
            action: 'drop',
          },
        ],
      },
      {
        url: 'http://smoke-test:19090/api/prom/push',
        write_relabel_configs: [
          {
            source_labels: ['__name__'],
            regex: 'avalanche_.*',
            action: 'keep',
          },
        ],
      },
    ],
    scrape_configs: [
      {
        job_name: 'avalanche',
        kubernetes_sd_configs: [{ role: 'pod' }],
        tls_config: {
          ca_file: '/var/run/secrets/kubernetes.io/serviceaccount/ca.crt',
        },
        bearer_token_file: '/var/run/secrets/kubernetes.io/serviceaccount/token',

        relabel_configs: [{
          source_labels: ['__meta_kubernetes_namespace'],
          regex: 'smoke',
          action: 'keep',
        }, {
          source_labels: ['__meta_kubernetes_pod_container_name'],
          regex: 'avalanche',
          action: 'keep',
        }],
      },
    ],
  }],

  normal_agent:
    gragent.new(name='grafana-agent', namespace='smoke') +
    gragent.withImagesMixin(images) +
    gragent.withStatefulSetController(
      replicas=1,
      volumeClaims=[
        pvc.new() +
        pvc.mixin.metadata.withName('agent-wal') +
        pvc.mixin.metadata.withNamespace('smoke') +
        pvc.mixin.spec.withAccessModes('ReadWriteOnce') +
        pvc.mixin.spec.resources.withRequests({ storage: '5Gi' }),
      ],
    ) +
    gragent.withVolumeMountsMixin([volumeMount.new('agent-wal', '/var/lib/agent')]) +
    gragent.withService() +
    gragent.withAgentConfig({
      server: { log_level: 'debug' },

      prometheus: {
        global: {
          scrape_interval: '1m',
          external_labels: {
            cluster: 'grafana-agent',
          },
        },
        wal_directory: '/var/lib/agent/data',
        configs: metric_instances('crow-single'),
      },
    }),

  cluster_agent:
    gragent.new(name='grafana-agent-cluster', namespace='smoke') +
    gragent.withImagesMixin(images) +
    gragent.withStatefulSetController(
      replicas=3,
      volumeClaims=[
        pvc.new() +
        pvc.mixin.metadata.withName('agent-cluster-wal') +
        pvc.mixin.metadata.withNamespace('smoke') +
        pvc.mixin.spec.withAccessModes('ReadWriteOnce') +
        pvc.mixin.spec.resources.withRequests({ storage: '5Gi' }),
      ],
    ) +
    gragent.withVolumeMountsMixin([volumeMount.new('agent-cluster-wal', '/var/lib/agent')]) +
    gragent.withService() +
    gragent.withAgentConfig({
      server: { log_level: 'debug' },

      prometheus: {
        global: {
          scrape_interval: '1m',
          external_labels: {
            cluster: 'grafana-agent-cluster',
          },
        },
        wal_directory: '/var/lib/agent/data',

        scraping_service: {
          enabled: true,
          dangerous_allow_reading_files: true,
          kvstore: {
            store: 'etcd',
            etcd: { endpoints: ['etcd:2379'] },
          },
          lifecycler: {
            ring: {
              kvstore: {
                store: 'etcd',
                etcd: { endpoints: ['etcd:2379'] },
              },
            },
          },
        },
      },
    }),

  // Spawn a syncer so our cluster gets the same scrape jobs as our
  // normal agent.
  sycner: gragent.newSyncer(
    name='grafana-agent-syncer',
    namespace='smoke',
    config={
      image: images.agentctl,
      api: 'http://grafana-agent-cluster.smoke.svc.cluster.local',
      configs: metric_instances('crow-cluster'),
    }
  ),
};

{
  monitoring: monitoring,
  smoke: smoke,
}
