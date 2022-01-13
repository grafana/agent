local k8s_v2 = import './v2/internal/helpers/k8s.libsonnet';

{
  _images+:: {
    agent: 'grafana/agent:latest',
    agentctl: 'grafana/agentctl:latest',
  },

  _config+:: {
    //
    // Deployment options
    //
    agent_cluster_role_name: 'grafana-agent',
    agent_configmap_name: 'grafana-agent',
    agent_deployment_configmap_name: self.agent_configmap_name + '-deployment',
    agent_pod_name: 'grafana-agent',
    agent_deployment_pod_name: self.agent_pod_name + '-deployment',

    cluster_dns_tld: 'local',
    cluster_dns_suffix: 'cluster.' + self.cluster_dns_tld,
    cluster_name: error 'must specify cluster name',
    namespace: error 'must specify namespace',

    agent_config_hash_annotation: true,

    //
    // Prometheus instance options
    //

    // Enabling this causes the agent to only scrape metrics on the same node
    // on which it is currently running.
    //
    // Take CAUTION when disabling this! If the agent is deployed
    // as a DaemonSet (like it is here by default), then disabling this will
    // scrape all metrics multiple times, once per node, leading to
    // duplicate samples being rejected and might hit limits.
    agent_host_filter: true,

    // The directory where the WAL is stored for all instances.
    agent_wal_dir: '/var/lib/agent/data',

    prometheus_kubernetes_api_server_address: 'kubernetes.default.svc.%(cluster_dns_suffix)s:443' % self,
    prometheus_insecure_skip_verify: false,
    scrape_api_server_endpoints: true,

    //
    // Config passed to the agent
    //
    // agent_config is rendered as a YAML and is the configuration file used
    // to control the agent. A single instance is hard-coded and its
    // scrape_configs are defined below.
    //
    // deployment_agent_config is a copy of `agent_config` that is used by the
    // single-replica deployment to scrape jobs that don't work in host
    // filtering mode.
    agent_config: {
      server: {
        log_level: 'info',
      },

      metrics: {
        global: {
          scrape_interval: '1m',
        },

        wal_directory: $._config.agent_wal_dir,

        configs: [{
          name: 'agent',

          host_filter: $._config.agent_host_filter,

          scrape_configs:
            if $._config.agent_host_filter then
              $._config.kubernetes_scrape_configs
            else
              $._config.kubernetes_scrape_configs + $._config.deployment_scrape_configs,
          remote_write: $._config.agent_remote_write,
        }],
      },
    },
    deployment_agent_config: self.agent_config {
      prometheus+: {
        configs: [{
          name: 'agent',

          host_filter: false,

          scrape_configs: $._config.deployment_scrape_configs,
          remote_write: $._config.agent_remote_write,
        }],
      },

    },

    local all_scrape_configs = k8s_v2.metrics({
      scrape_api_server_endpoints: $._config.scrape_api_server_endpoints,
      insecure_skip_verify: $._config.prometheus_insecure_skip_verify,
      kubernetes_api_server_address: $._config.prometheus_kubernetes_api_server_address,
      ksm_namespace: $._config.namespace,
      node_exporter_namespace: $._config.namespace,
    }),

    // We have two optional extension points for scrape config. One for the
    // statefulset that holds all the agents attached to a node
    // (kubernetes_scrape_configs) and One for the single replica deployment
    // that is used to scrape jobs that don't work with host filtering mode
    // (deployment_scrape_configs) the later is only used when host_filter =
    // true.
    deployment_scrape_configs:
      std.filter(function(job) job.job_name == 'default/kubernetes', all_scrape_configs),
    kubernetes_scrape_configs:
      std.filter(function(job) job.job_name != 'default/kubernetes', all_scrape_configs),

    agent_remote_write: [],
  },
}
