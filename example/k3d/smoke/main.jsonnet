local monitoring = import './monitoring/main.jsonnet';
local cortex = import 'cortex/main.libsonnet';
local canary = import 'github.com/grafana/loki/production/ksonnet/loki-canary/loki-canary.libsonnet';
local vulture = import 'github.com/grafana/tempo/operations/jsonnet/microservices/vulture.libsonnet';
local tempo = import 'github.com/grafana/tempo/operations/jsonnet/single-binary/tempo.libsonnet';
local avalanche = import 'grafana-agent/smoke/avalanche/main.libsonnet';
local crow = import 'grafana-agent/smoke/crow/main.libsonnet';
local etcd = import 'grafana-agent/smoke/etcd/main.libsonnet';
local smoke = import 'grafana-agent/smoke/main.libsonnet';
local gragent = import 'grafana-agent/v2/main.libsonnet';
local k = import 'ksonnet-util/kausal.libsonnet';
local loki = import 'loki/main.libsonnet';

local namespace = k.core.v1.namespace;
local pvc = k.core.v1.persistentVolumeClaim;
local volumeMount = k.core.v1.volumeMount;
local containerPort = k.core.v1.containerPort;
local statefulset = k.apps.v1.statefulSet;
local service = k.core.v1.service;
local configMap = k.core.v1.configMap;
local deployment = k.apps.v1.deployment;
local daemonSet = k.apps.v1.daemonSet;

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

  tempo: tempo {
    _config+:: {
      namespace: 'smoke',
      tempo: {
        port: 3200,
        replicas: 1,
        headless_service_name: 'localhost',
      },
      pvc_size: '30Gi',
      pvc_storage_class: 'local-path',
      receivers: {
        jaeger: {
          protocols: {
            thrift_http: null,
          },
        },
        otlp: {
          protocols: {
            grpc: {
              endpoint: '0.0.0.0:4317',
            },
          },
        },
      },
    },
    tempo_config+: {
      querier: {
        frontend_worker: {
          frontend_address: 'localhost:9095',
        },
      },
    },
    tempo_statefulset+:
      statefulset.mixin.metadata.withNamespace('smoke'),
    tempo_service+:
      service.mixin.metadata.withNamespace('smoke'),
    tempo_headless_service+:
      service.mixin.metadata.withNamespace('smoke'),
    tempo_query_configmap+:
      configMap.mixin.metadata.withNamespace('smoke'),
    tempo_configmap+:
      configMap.mixin.metadata.withNamespace('smoke'),
  },

  loki: loki.new(namespace='smoke'),

  // https://grafana.com/docs/loki/latest/operations/loki-canary/
  canary: canary {
    loki_canary_args+:: {
      addr: 'loki:80',
      port: '80',
      tls: false,
      labelname: 'instance',
      labelvalue: '$(POD_NAME)',
      interval: '1s',
      'metric-test-interval': '30m',
      'metric-test-range': '2h',
      size: 1024,
      wait: '3m',
    },
    _config+:: {
      namespace: 'smoke',
    },
    loki_canary_daemonset+:
      daemonSet.mixin.metadata.withNamespace('smoke'),
  },

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

  vulture: vulture {
    _images+:: {
      tempo_vulture: 'grafana/tempo-vulture:latest',
    },
    _config+:: {
      vulture: {
        replicas: 1,
        tempoPushUrl: 'http://grafana-agent',
        tempoQueryUrl: 'http://tempo:3200',
        tempoOrgId: '',
        tempoRetentionDuration: '336h',
        tempoSearchBackoffDuration: '0s',
        tempoReadBackoffDuration: '10s',
        tempoWriteBackoffDuration: '10s',
      },
    },
    tempo_vulture_deployment+:
      deployment.mixin.metadata.withNamespace('smoke'),
  },

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
  }, {
    name: 'vulture',
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
        job_name: 'vulture',
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
  }, {
    name: 'canary',
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
        job_name: 'canary',
        kubernetes_sd_configs: [{ role: 'pod' }],
        tls_config: {
          ca_file: '/var/run/secrets/kubernetes.io/serviceaccount/ca.crt',
        },
        bearer_token_file: '/var/run/secrets/kubernetes.io/serviceaccount/token',

        relabel_configs: [
          {
            source_labels: ['__meta_kubernetes_namespace'],
            regex: 'smoke',
            action: 'keep',
          },
          {
            source_labels: ['__meta_kubernetes_pod_container_name'],
            regex: 'canary',
            action: 'keep',
          },
        ],
      },
    ],
  }],

  local logs_instances() = [{
    name: 'write-loki',
    clients: [{
      url: 'http://loki/loki/api/v1/push',
      basic_auth: {
        username: '104334',
        password: 'noauth',
      },
      external_labels: {
        cluster: 'grafana-agent',
      },

    }],
    scrape_configs: [{
      job_name: 'write-canary-output',
      kubernetes_sd_configs: [{ role: 'pod' }],
      pipeline_stages: [
        { cri: {} },
      ],
      relabel_configs: [
        {
          source_labels: ['__meta_kubernetes_namespace'],
          regex: 'smoke',
          action: 'keep',
        },
        {
          source_labels: ['__meta_kubernetes_pod_container_name'],
          regex: 'loki-canary',
          action: 'keep',
        },
        {
          action: 'replace',
          source_labels: ['__meta_kubernetes_pod_uid', '__meta_kubernetes_pod_container_name'],
          target_label: '__path__',
          separator: '/',
          replacement: '/var/log/pods/*$1/*.log',
        },
        {
          action: 'replace',
          source_labels: ['__meta_kubernetes_pod_name'],
          target_label: 'pod',
        },
        {
          action: 'replace',
          source_labels: ['__meta_kubernetes_pod_name'],
          target_label: 'instance',
        },
      ],
    }],
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
    gragent.withPortsMixin([
      containerPort.new('thrift-grpc', 14250) + containerPort.withProtocol('TCP'),
    ]) +
    gragent.withLogVolumeMounts() +
    gragent.withAgentConfig({
      server: { log_level: 'debug' },
      logs: {
        positions_directory: '/var/lib/agent/logs-positions',
        configs: logs_instances(),
      },

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
      traces: {
        configs: [
          {
            name: 'vulture',
            receivers: {
              jaeger: {
                protocols: {
                  grpc: null,
                },
              },
            },
            remote_write: [
              {
                endpoint: 'tempo:4317',
                insecure: true,
              },
            ],
            batch: {
              timeout: '5s',
              send_batch_size: 100,
            },
          },
        ],
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
    gragent.withLogVolumeMounts() +
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
