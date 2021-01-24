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

        // Drop some high cardinality metrics.
        metric_relabel_configs: [
          {
            action: 'drop',
            regex: 'kubelet_(pod_worker_latency_microseconds|pod_start_latency_microseconds|cgroup_manager_latency_microseconds|pod_worker_start_latency_microseconds|pleg_relist_latency_microseconds|pleg_relist_interval_microseconds|runtime_operations|runtime_operations_latency_microseconds|runtime_operations_errors|eviction_stats_age_microseconds|device_plugin_registration_count|device_plugin_alloc_latency_microseconds|network_plugin_operations_latency_microseconds)',
            source_labels: ['__name__'],
          },
          {
            action: 'drop',
            regex: 'scheduler_(e2e_scheduling_latency_microseconds|scheduling_algorithm_predicate_evaluation|scheduling_algorithm_priority_evaluation|scheduling_algorithm_preemption_evaluation|scheduling_algorithm_latency_microseconds|binding_latency_microseconds|scheduling_latency_seconds)',
            source_labels: ['__name__'],
          },
          {
            action: 'drop',
            regex: 'apiserver_(request_count|request_latencies|request_latencies_summary|dropped_requests|storage_data_key_generation_latencies_microseconds|storage_transformation_failures_total|storage_transformation_latencies_microseconds|proxy_tunnel_sync_latency_secs)',
            source_labels: ['__name__'],
          },
          {
            action: 'drop',
            regex: 'kubelet_docker_(operations|operations_latency_microseconds|operations_errors|operations_timeout)',
            source_labels: ['__name__'],
          },
          {
            action: 'drop',
            regex: 'reflector_(items_per_list|items_per_watch|list_duration_seconds|lists_total|short_watches_total|watch_duration_seconds|watches_total)',
            source_labels: ['__name__'],
          },
          {
            action: 'drop',
            regex: 'etcd_(helper_cache_hit_count|helper_cache_miss_count|helper_cache_entry_count|request_cache_get_latencies_summary|request_cache_add_latencies_summary|request_latencies_summary)',
            source_labels: ['__name__'],
          },
          {
            action: 'drop',
            regex: 'transformation_(transformation_latencies_microseconds|failures_total)',
            source_labels: ['__name__'],
          },
          {
            action: 'drop',
            regex: '(admission_quota_controller_adds|crd_autoregistration_controller_work_duration|APIServiceOpenAPIAggregationControllerQueue1_adds|AvailableConditionController_retries|crd_openapi_controller_unfinished_work_seconds|APIServiceRegistrationController_retries|admission_quota_controller_longest_running_processor_microseconds|crdEstablishing_longest_running_processor_microseconds|crdEstablishing_unfinished_work_seconds|crd_openapi_controller_adds|crd_autoregistration_controller_retries|crd_finalizer_queue_latency|AvailableConditionController_work_duration|non_structural_schema_condition_controller_depth|crd_autoregistration_controller_unfinished_work_seconds|AvailableConditionController_adds|DiscoveryController_longest_running_processor_microseconds|autoregister_queue_latency|crd_autoregistration_controller_adds|non_structural_schema_condition_controller_work_duration|APIServiceRegistrationController_adds|crd_finalizer_work_duration|crd_naming_condition_controller_unfinished_work_seconds|crd_openapi_controller_longest_running_processor_microseconds|DiscoveryController_adds|crd_autoregistration_controller_longest_running_processor_microseconds|autoregister_unfinished_work_seconds|crd_naming_condition_controller_queue_latency|crd_naming_condition_controller_retries|non_structural_schema_condition_controller_queue_latency|crd_naming_condition_controller_depth|AvailableConditionController_longest_running_processor_microseconds|crdEstablishing_depth|crd_finalizer_longest_running_processor_microseconds|crd_naming_condition_controller_adds|APIServiceOpenAPIAggregationControllerQueue1_longest_running_processor_microseconds|DiscoveryController_queue_latency|DiscoveryController_unfinished_work_seconds|crd_openapi_controller_depth|APIServiceOpenAPIAggregationControllerQueue1_queue_latency|APIServiceOpenAPIAggregationControllerQueue1_unfinished_work_seconds|DiscoveryController_work_duration|autoregister_adds|crd_autoregistration_controller_queue_latency|crd_finalizer_retries|AvailableConditionController_unfinished_work_seconds|autoregister_longest_running_processor_microseconds|non_structural_schema_condition_controller_unfinished_work_seconds|APIServiceOpenAPIAggregationControllerQueue1_depth|AvailableConditionController_depth|DiscoveryController_retries|admission_quota_controller_depth|crdEstablishing_adds|APIServiceOpenAPIAggregationControllerQueue1_retries|crdEstablishing_queue_latency|non_structural_schema_condition_controller_longest_running_processor_microseconds|autoregister_work_duration|crd_openapi_controller_retries|APIServiceRegistrationController_work_duration|crdEstablishing_work_duration|crd_finalizer_adds|crd_finalizer_depth|crd_openapi_controller_queue_latency|APIServiceOpenAPIAggregationControllerQueue1_work_duration|APIServiceRegistrationController_queue_latency|crd_autoregistration_controller_depth|AvailableConditionController_queue_latency|admission_quota_controller_queue_latency|crd_naming_condition_controller_work_duration|crd_openapi_controller_work_duration|DiscoveryController_depth|crd_naming_condition_controller_longest_running_processor_microseconds|APIServiceRegistrationController_depth|APIServiceRegistrationController_longest_running_processor_microseconds|crd_finalizer_unfinished_work_seconds|crdEstablishing_retries|admission_quota_controller_unfinished_work_seconds|non_structural_schema_condition_controller_adds|APIServiceRegistrationController_unfinished_work_seconds|admission_quota_controller_work_duration|autoregister_depth|autoregister_retries|kubeproxy_sync_proxy_rules_latency_microseconds|rest_client_request_latency_seconds|non_structural_schema_condition_controller_retries)',
            source_labels: ['__name__'],
          },
          {
            action: 'drop',
            regex: 'etcd_(debugging|disk|server).*',
            source_labels: ['__name__'],
          },
          {
            action: 'drop',
            regex: 'apiserver_admission_controller_admission_latencies_seconds_.*',
            source_labels: ['__name__'],
          },
          {
            action: 'drop',
            regex: 'apiserver_admission_step_admission_latencies_seconds_.*',
            source_labels: ['__name__'],
          },
          {
            action: 'drop',
            regex: 'apiserver_request_duration_seconds_bucket;(0.15|0.25|0.3|0.35|0.4|0.45|0.6|0.7|0.8|0.9|1.25|1.5|1.75|2.5|3|3.5|4.5|6|7|8|9|15|25|30|50)',
            source_labels: ['__name__', 'le'],
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

        tls_config: {
          ca_file: '/var/run/secrets/kubernetes.io/serviceaccount/ca.crt',
          insecure_skip_verify: $._config.prometheus_insecure_skip_verify,
        },
        bearer_token_file: '/var/run/secrets/kubernetes.io/serviceaccount/token',

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

        tls_config: {
          ca_file: '/var/run/secrets/kubernetes.io/serviceaccount/ca.crt',
          insecure_skip_verify: $._config.prometheus_insecure_skip_verify,
        },
        bearer_token_file: '/var/run/secrets/kubernetes.io/serviceaccount/token',

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

        tls_config: {
          ca_file: '/var/run/secrets/kubernetes.io/serviceaccount/ca.crt',
          insecure_skip_verify: $._config.prometheus_insecure_skip_verify,
        },
        bearer_token_file: '/var/run/secrets/kubernetes.io/serviceaccount/token',

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
            replacement: '/api/v1/nodes/${1}/proxy/metrics',
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
            replacement: '/api/v1/nodes/${1}/proxy/metrics/cadvisor',
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
