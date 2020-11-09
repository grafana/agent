local agent_cluster = import 'grafana-agent/scraping-svc/main.libsonnet';
local default = import 'default/main.libsonnet';
local etcd = import 'etcd/main.libsonnet';
local grafana_agent = import 'grafana-agent/v1/main.libsonnet';
local k = import 'ksonnet-util/kausal.libsonnet';

local containerPort = k.core.v1.containerPort;

local service = k.core.v1.service;
local images = {
  agent: 'grafana/agent:latest',
  agentctl: 'grafana/agentctl:latest',
};

{
  default: default.new(namespace='default') {
    grafana+: {
      // Expose Grafana on 30080 on the k3d agent, which is exposed to the host
      // machine.
      service+:
        local bindNodePort(port) = port { nodePort: port.port + 30000 };
        service.mixin.spec.withPorts([
          { name: 'grafana', nodePort: 30080, port: 30080, targetPort: 80 },
        ]) +
        service.mixin.spec.withType('NodePort'),
    },
  },

  agent:
    local cluster_label = 'k3d-agent/daemonset';

    grafana_agent.new('grafana-agent', 'default') +
    grafana_agent.withImages(images) +

    //
    // Configure Prometheus 
    //

    grafana_agent.withPrometheusConfig({
      wal_directory: '/var/lib/agent/data',
      global: {
        scrape_interval: '15s',
        external_labels: {
          cluster: cluster_label,
        },
      },
    }) +
    grafana_agent.withPrometheusInstances(grafana_agent.scrapeInstanceKubernetes {
      // We want our cluster and label to remain static for this deployment, so
      // if they are overwritten by a metric we will change them to the values
      // set by external_labels.
      scrape_configs: std.map(function(config) config {
        relabel_configs+: [{
          target_label: 'cluster', replacement: cluster_label,
        }]
      }, super.scrape_configs),
    }) +
    grafana_agent.withRemoteWrite([{
      url: 'http://cortex.default.svc.cluster.local/api/prom/push',
    }]) +

    //
    // Configure Tempo
    //

    grafana_agent.withEnv([
      // Configure the Agent to send spans to itself using thrift_http.
      { name: 'JAEGER_AGENT_HOST', value: 'localhost' },
      { name: 'JAEGER_AGENT_PORT', value: '6831' },

      // Send every span. You wouldn't want this in prod, but our span
      // throughout here is low enough that we need a high sampling rate.
      { name: 'JAEGER_SAMPLER_TYPE', value: 'const' },
      { name: 'JAEGER_SAMPLER_PARAM', value: '1' },
    ]) +
    grafana_agent.withTempoConfig({
      receivers: {
        jaeger: {
          protocols: {
            thrift_http: null,
            thrift_binary: null,
            thrift_compact: null,
            grpc: null,
          },
        },
      },
    }) + 
    grafana_agent.withPortsMixin([
      // Jaeger receiver
      containerPort.new('thrift-c', 6831) + containerPort.withProtocol('UDP'),
      containerPort.new('thrift-bin', 6832) + containerPort.withProtocol('UDP'),
      containerPort.new('thrift-http', 14268) + containerPort.withProtocol('TCP'),
      containerPort.new('thrift-grpc', 14250) + containerPort.withProtocol('TCP'),
    ]) +
    grafana_agent.withTempoPushConfig({
      endpoint: 'tempo:55680',
      insecure: true,
      batch: {
        timeout: '5s',
        send_batch_size: 100,
      },
    }) + 
    grafana_agent.withTempoScrapeConfigs([
      cfg {
        // We want our cluster and label to remain static for this deployment, so
        // if they are overwritten by a metric we will change them to the values
        // set by external_labels.
        relabel_configs+: [{
          target_label: 'cluster', replacement: cluster_label,
        }]
      }
      for cfg in grafana_agent.tempoScrapeKubernetes
    ]),

  // Need to run ETCD for agent_cluster
  etcd: etcd.new('default'),

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
        prometheus+: {
          global+: {
            external_labels+: {
              cluster: cluster_label,
            },
          },
        },
      },

      // We want our cluster and agent labels to remain static
      // for this deployment, so if they are overwritten by a metric
      // we will change them to the values set by external_labels.
      kubernetes_scrape_configs: std.map(function(config) config {
        relabel_configs+: [
          { target_label: 'cluster', replacement: cluster_label },
        ],
      }, super.deployment_scrape_configs + super.kubernetes_scrape_configs),
    }),
}
