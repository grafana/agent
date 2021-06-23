local monitoring = import './monitoring/main.jsonnet';
local avalanche = import 'avalanche/main.libsonnet';
local cortex = import 'cortex/main.libsonnet';
local etcd = import 'etcd/main.libsonnet';
local gragent = import 'grafana-agent/v2/main.libsonnet';
local k = import 'ksonnet-util/kausal.libsonnet';
local test_exporter = import 'test-exporter/main.libsonnet';

local namespace = k.core.v1.namespace;
local pvc = k.core.v1.persistentVolumeClaim;
local volumeMount = k.core.v1.volumeMount;

local images = {
  agent: 'grafana/agent:latest',
  agentctl: 'grafana/agentctl:latest',
};

local new_test_exporter(name, selector) =
  test_exporter.new(name, namespace='smoke', config={
    args: {
      'prometheus-address': 'http://cortex/api/prom',
      'scrape-interval': '15s',
      'extra-selectors': selector,
    },
  });


local smoke = {
  ns: namespace.new('smoke'),

  cortex: cortex.new('smoke'),

  // Needed to run agent cluster
  etcd: etcd.new('smoke'),

  avalanche: avalanche.new(replicas=3, namespace='smoke', config={
    // We're going to be running a lot of these and we're not trying to test
    // for load, so reduce the cardinality and churn rate.
    metric_count: 50,
    series_interval: 300,
    metric_interval: 600,
  }),

  test_exporters: [
    new_test_exporter('test-exporter-single', 'cluster="grafana-agent"'),
    new_test_exporter('test-exporter-cluster', 'cluster="grafana-agent-cluster"'),
  ],

  local metric_instances(test_exporter_name) = [{
    name: 'test-exporter',
    remote_write: [{ url: 'http://cortex/api/prom/push' }],
    scrape_configs: [
      {
        job_name: 'test-exporter',
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
          regex: test_exporter_name,
          action: 'keep',
        }],
      },
    ],
  }, {
    name: 'avalanche',
    remote_write: [{ url: 'http://cortex/api/prom/push' }],
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
    gragent.withAgentConfig({
      server: { log_level: 'debug' },

      prometheus: {
        global: {
          scrape_interval: '15s',
          external_labels: {
            cluster: 'grafana-agent',
          },
        },
        wal_directory: '/var/lib/agent/data',
        configs: metric_instances('test-exporter-single'),
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
    gragent.withAgentConfig({
      server: { log_level: 'debug' },

      prometheus: {
        global: {
          scrape_interval: '15s',
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
    name='grafana-agent-sycner',
    namespace='smoke',
    config={
      image: images.agentctl,
      api: 'http://grafana-agent-cluster.smoke.svc.cluster.local',
      configs: metric_instances('test-exporter-cluster'),
    }
  ),
};

{
  monitoring: monitoring,
  smoke: smoke,
}
