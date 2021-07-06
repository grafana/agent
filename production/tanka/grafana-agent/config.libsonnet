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

      prometheus: {
        global: {
          scrape_interval: '15s',
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

    //
    // We have two optional extension points for scrape config. One for the
    // statefulset that holds all the agents attached to a node
    // (kubernetes_scrape_configs) and One for the single replica deployment
    // that is used to scrape jobs that don't work with host filtering mode
    // (deployment_scrape_configs) the later is only used when host_filter =
    // true.
    deployment_scrape_configs: [
      {
        job_name: 'default/kubernetes',
        kubernetes_sd_configs: [{
          role:
            if $._config.scrape_api_server_endpoints
            then 'endpoints'
            else 'service',
        }],
        scheme: 'https',

        tls_config: {
          ca_file: '/var/run/secrets/kubernetes.io/serviceaccount/ca.crt',
          insecure_skip_verify: $._config.prometheus_insecure_skip_verify,
          server_name: 'kubernetes',
        },
        bearer_token_file: '/var/run/secrets/kubernetes.io/serviceaccount/token',
        relabel_configs: [{
          source_labels: ['__meta_kubernetes_service_label_component'],
          regex: 'apiserver',
          action: 'keep',
        }],

        // Keep limited set of metrics to reduce default usage, drop all others
        metric_relabel_configs: [
          {
            source_labels: ['__name__'],
            regex: 'workqueue_queue_duration_seconds_bucket|process_cpu_seconds_total|process_resident_memory_bytes|workqueue_depth|rest_client_request_duration_seconds_bucket|workqueue_adds_total|up|rest_client_requests_total|apiserver_request_total|go_goroutines',
            action: 'keep',
          },
        ],

      },
    ],
    kubernetes_scrape_configs: [
      {
        job_name: 'kubernetes-pods',
        kubernetes_sd_configs: [{
          role: 'pod',
        }],

        // You can specify the following annotations (on pods):
        //   prometheus.io/scrape: false - don't scrape this pod
        //   prometheus.io/scheme: https - use https for scraping
        //   prometheus.io/port - scrape this port
        //   prometheus.io/path - scrape this path
        //   prometheus.io/param-<parameter> - send ?parameter=value with the scrape
        relabel_configs: [
          // Drop anything annotated with prometheus.io/scrape=false
          {
            source_labels: ['__meta_kubernetes_pod_annotation_prometheus_io_scrape'],
            action: 'drop',
            regex: 'false',
          },

          // Drop any endpoint whose pod port name does not end with metrics
          {
            source_labels: ['__meta_kubernetes_pod_container_port_name'],
            action: 'keep',
            regex: '.*-metrics',
          },

          // Allow pods to override the scrape scheme with prometheus.io/scheme=https
          {
            source_labels: ['__meta_kubernetes_pod_annotation_prometheus_io_scheme'],
            action: 'replace',
            target_label: '__scheme__',
            regex: '(https?)',
            replacement: '$1',
          },

          // Allow service to override the scrape path with prometheus.io/path=/other_metrics_path
          {
            source_labels: ['__meta_kubernetes_pod_annotation_prometheus_io_path'],
            action: 'replace',
            target_label: '__metrics_path__',
            regex: '(.+)',
            replacement: '$1',
          },

          // Allow services to override the scrape port with prometheus.io/port=1234
          {
            source_labels: ['__address__', '__meta_kubernetes_pod_annotation_prometheus_io_port'],
            action: 'replace',
            target_label: '__address__',
            regex: '(.+?)(\\:\\d+)?;(\\d+)',
            replacement: '$1:$3',
          },

          // Drop pods without a name label
          {
            source_labels: ['__meta_kubernetes_pod_label_name'],
            action: 'drop',
            regex: '',
          },

          // Rename jobs to be <namespace>/<name, from pod name label>
          {
            source_labels: ['__meta_kubernetes_namespace', '__meta_kubernetes_pod_label_name'],
            action: 'replace',
            separator: '/',
            target_label: 'job',
            replacement: '$1',
          },

          // But also include the namespace as a separate label for routing alerts
          {
            source_labels: ['__meta_kubernetes_namespace'],
            action: 'replace',
            target_label: 'namespace',
          },
          {
            source_labels: ['__meta_kubernetes_pod_name'],
            action: 'replace',
            target_label: 'pod',  // Not 'pod_name', which disappeared in K8s 1.16.
          },
          {
            source_labels: ['__meta_kubernetes_pod_container_name'],
            action: 'replace',
            target_label: 'container',  // Not 'container_name', which disappeared in K8s 1.16.
          },

          // Rename instances to the concatenation of pod:container:port.
          // All three components are needed to guarantee a unique instance label.
          {
            source_labels: [
              '__meta_kubernetes_pod_name',
              '__meta_kubernetes_pod_container_name',
              '__meta_kubernetes_pod_container_port_name',
            ],
            action: 'replace',
            separator: ':',
            target_label: 'instance',
          },

          // Map prometheus.io/param-<name>=value fields to __param_<name>=value
          {
            regex: '__meta_kubernetes_pod_annotation_prometheus_io_param_(.+)',
            action: 'labelmap',
            replacement: '__param_$1',
          },

          // Drop pods with phase Succeeded or Failed
          {
            source_labels: ['__meta_kubernetes_pod_phase'],
            action: 'drop',
            regex: 'Succeeded|Failed',
          },
        ],
      },

      // A separate scrape config for kube-state-metrics which doesn't add a
      // namespace label and instead takes the namespace label from the
      // exported timeseries. This prevents the exported namespace label from
      // being renamed to exported_namespace and allows us to route alerts
      // based on namespace.
      {
        job_name: '%s/kube-state-metrics' % $._config.namespace,
        kubernetes_sd_configs: [{
          role: 'pod',
          namespaces: {
            names: [$._config.namespace],
          },
        }],

        relabel_configs: [
          // Drop anything whose service is not kube-state-metrics
          {
            source_labels: ['__meta_kubernetes_pod_label_name'],
            regex: 'kube-state-metrics',
            action: 'keep',
          },

          // Rename instances to the concatenation of pod:container:port.
          // In the specific case of KSM, we could leave out the container
          // name and still have a unique instance label, but we leave it
          // in here for consistency with the normal pod scraping.
          {
            source_labels: [
              '__meta_kubernetes_pod_name',
              '__meta_kubernetes_pod_container_name',
              '__meta_kubernetes_pod_container_port_name',
            ],
            action: 'replace',
            separator: ':',
            target_label: 'instance',
          },
        ],
      },

      // A separate scrape config for node-exporter which maps the nodename
      // onto the instance label.
      {
        job_name: '%s/node-exporter' % $._config.namespace,
        kubernetes_sd_configs: [{
          role: 'pod',
          namespaces: {
            names: [$._config.namespace],
          },
        }],

        relabel_configs: [
          // Drop anything whose name is not node-exporter.
          {
            source_labels: ['__meta_kubernetes_pod_label_name'],
            regex: 'node-exporter',
            action: 'keep',
          },

          // Rename instances to be the node name.
          {
            source_labels: ['__meta_kubernetes_pod_node_name'],
            action: 'replace',
            target_label: 'instance',
          },

          // But also include the namespace as a separate label, for
          // routing alerts.
          {
            source_labels: ['__meta_kubernetes_namespace'],
            action: 'replace',
            target_label: 'namespace',
          },
        ],
      },

      // This scrape config gathers all kubelet metrics.
      {
        job_name: 'kube-system/kubelet',
        kubernetes_sd_configs: [{
          role: 'node',
        }],

        tls_config: {
          ca_file: '/var/run/secrets/kubernetes.io/serviceaccount/ca.crt',
          insecure_skip_verify: $._config.prometheus_insecure_skip_verify,
        },
        bearer_token_file: '/var/run/secrets/kubernetes.io/serviceaccount/token',

        relabel_configs: [
          {
            target_label: '__address__',
            replacement: $._config.prometheus_kubernetes_api_server_address,
          },
          {
            target_label: '__scheme__',
            replacement: 'https',
          },
          {
            source_labels: ['__meta_kubernetes_node_name'],
            regex: '(.+)',
            target_label: '__metrics_path__',
            replacement: '/api/v1/nodes/$1/proxy/metrics',
          },
        ],
      },

      // As of k8s 1.7.3, cAdvisor metrics are available via kubelet using
      // the /metrics/cadvisor path.
      {
        job_name: 'kube-system/cadvisor',
        kubernetes_sd_configs: [{
          role: 'node',
        }],
        scheme: 'https',

        tls_config: {
          ca_file: '/var/run/secrets/kubernetes.io/serviceaccount/ca.crt',
          insecure_skip_verify: $._config.prometheus_insecure_skip_verify,
        },
        bearer_token_file: '/var/run/secrets/kubernetes.io/serviceaccount/token',

        relabel_configs: [
          {
            target_label: '__address__',
            replacement: $._config.prometheus_kubernetes_api_server_address,
          },
          {
            source_labels: ['__meta_kubernetes_node_name'],
            regex: '(.+)',
            target_label: '__metrics_path__',
            replacement: '/api/v1/nodes/$1/proxy/metrics/cadvisor',
          },
        ],

        metric_relabel_configs: [
          // Drop container_* metrics with no image.
          {
            source_labels: ['__name__', 'image'],
            regex: 'container_([a-z_]+);',
            action: 'drop',
          },

          // Drop a bunch of metrics which are disabled but still sent,
          // see https://github.com/google/cadvisor/issues/1925.
          {
            source_labels: ['__name__'],
            regex: 'container_(network_tcp_usage_total|network_udp_usage_total|tasks_state|cpu_load_average_10s)',
            action: 'drop',
          },
        ],
      },
    ],

    agent_remote_write: [],
  },
}
