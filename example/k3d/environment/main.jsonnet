local collector = import 'collector/main.libsonnet';
local default = import 'default/main.libsonnet';
local etcd = import 'grafana-agent/smoke/etcd/main.libsonnet';
local agent_cluster = import 'grafana-agent/scraping-svc/main.libsonnet';
local k = import 'ksonnet-util/kausal.libsonnet';
local load_generator = import 'load-generator/main.libsonnet';

local loki_config = import 'default/loki_config.libsonnet';
local grafana_agent = import 'grafana-agent/v1/main.libsonnet';

local containerPort = k.core.v1.containerPort;
local ingress = k.networking.v1beta1.ingress;
local path = k.networking.v1beta1.httpIngressPath;
local rule = k.networking.v1beta1.ingressRule;
local service = k.core.v1.service;

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

  agent:
    local cluster_label = 'k3d-agent/daemonset';

    grafana_agent.new('grafana-agent', 'default') +
    grafana_agent.withImages(images) +
    grafana_agent.withMetricsConfig({
      wal_directory: '/var/lib/agent/data',
      global: {
        scrape_interval: '1m',
        external_labels: {
          cluster: cluster_label,
        },
      },
    }) +
    grafana_agent.withMetricsInstances(grafana_agent.scrapeInstanceKubernetes {
      // We want our cluster and label to remain static for this deployment, so
      // if they are overwritten by a metric we will change them to the values
      // set by external_labels.
      scrape_configs: std.map(function(config) config {
        relabel_configs+: [{
          target_label: 'cluster',
          replacement: cluster_label,
        }],
      }, super.scrape_configs),
    }) +
    grafana_agent.withRemoteWrite([{
      url: 'http://cortex.default.svc.cluster.local/api/prom/push',
    }]) +
    grafana_agent.withLogsConfig(loki_config) +
    grafana_agent.withLogsClients(grafana_agent.newLogsClient({
      scheme: 'http',
      hostname: 'loki.default.svc.cluster.local',
      external_labels: { cluster: cluster_label },
    })) +
    grafana_agent.withTracesConfig({
      receivers: {
        jaeger: {
          protocols: {
            thrift_http: null,
          },
        },
      },
      batch: {
        timeout: '5s',
        send_batch_size: 1000,
      },
    }) +
    grafana_agent.withPortsMixin([
      containerPort.new('thrift-http', 14268) + containerPort.withProtocol('TCP'),
      containerPort.new('otlp-lb', 4318) + containerPort.withProtocol('TCP'),
    ]) +
    grafana_agent.withTracesRemoteWrite([
      {
        endpoint: 'collector.default.svc.cluster.local:4317',
        insecure: true,
      },
    ]) +
    grafana_agent.withTracesTailSamplingConfig({
      policies: [{
        type: 'always_sample',
      }],
    }) +
    grafana_agent.withTracesLoadBalancingConfig({
      exporter: {
        insecure: true,
      },
      resolver: {
        dns: {
          hostname: 'grafana-agent.default.svc.cluster.local',
          port: 4318,
        },
      },
    }),

  // Need to run ETCD for agent_cluster
  etcd: etcd.new('default'),

  collector: collector.new('default'),

  load_generator: load_generator.new('default'),

  agent_cluster:
    agent_cluster.new('default', 'kube-system') +
    agent_cluster.withImagesMixin(images) +
    agent_cluster.withConfigMixin({
      local kvstore = {
        store: 'etcd',
        etcd: {
          endpoints: ['etcd.default.svc.cluster.local:2379'],
        },
      },

      agent_remote_write: [{
        url: 'http://cortex.default.svc.cluster.local/api/prom/push',
      }],

      agent_ring_kvstore: kvstore { prefix: 'agent/ring/' },
      agent_config_kvstore: kvstore { prefix: 'agent/configs/' },

      local cluster_label = 'k3d-agent/cluster',
      agent_config+: {
        metrics+: {
          global+: {
            external_labels+: {
              cluster: cluster_label,
            },
          },

          scraping_service+: {
            dangerous_allow_reading_files: true,
          },
        },
      },

      kubernetes_scrape_configs:
        (grafana_agent.scrapeInstanceKubernetes {
           // We want our cluster and label to remain static for this deployment, so
           // if they are overwritten by a metric we will change them to the values
           // set by external_labels.
           scrape_configs: std.map(function(config) config {
             relabel_configs+: [{
               target_label: 'cluster',
               replacement: cluster_label,
             }],
           }, super.scrape_configs),
         }).scrape_configs,
    }),
}
