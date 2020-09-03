local k8s_tls_config(config) = {
  tls_config: {
    ca_file: '/var/run/secrets/kubernetes.io/serviceaccount/ca.crt',
    insecure_skip_verify: config.insecure_skip_verify,
  },
  bearer_token_file: '/var/run/secrets/kubernetes.io/serviceaccount/token',
};

{
  kubernetesScrapeInstanceConfig:: {
    scrape_api_server_endpoints: false,
    insecure_skip_verify: false,

    cluster_dns_tld: 'local',
    cluster_dns_suffix: 'cluster.' + self.cluster_dns_tld,
    kubernetes_api_server_address: 'kubernetes.default.svc.%(cluster_dns_suffix)s:443' % self,
  },

  newKubernetesScrapeInstance(config, namespace='default'):: {
    name: 'kubernetes',
    scrape_configs: [
      k8s_tls_config(config) {
        job_name: 'default/kubernetes',
        kubernetes_sd_configs: [{
          role: if config.scrape_api_server_endpoints then 'endpoints' else 'service',
        }],
        scheme: 'https',

        relabel_configs: [{
          source_labels: ['__meta_kubernetes_service_label_component'],
          regex: 'apiserver',
          action: 'keep',
        }],

        // Drop some high cardinality metrics.
        metric_relabel_configs: [
          {
            source_labels: ['__name__'],
            regex: 'apiserver_admission_controller_admission_latencies_seconds_.*',
            action: 'drop',
          },
          {
            source_labels: ['__name__'],
            regex: 'apiserver_admission_step_admission_latencies_seconds_.*',
            action: 'drop',
          },
        ],
      },

      k8s_tls_config(config) {
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
      // namespace label and instead takes the namespace label from the exported
      // timeseries. This prevents the exported namespace label from being
      // renamed to exported_namesapce and allows us to route alerts based on
      // namespace.
      k8s_tls_config(config) {
        job_name: '%s/kube-state-metrics' % namespace,
        kubernetes_sd_configs: [{
          role: 'pod',
          namespaces: {
            names: [namespace],
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

      // A separate scrape config for node-exporter which maps the node name
      // onto the instance label.
      k8s_tls_config(config) {
        job_name: '%s/node-exporter' % namespace,
        kubernetes_sd_configs: [{
          role: 'pod',
          namespaces: {
            names: [namespace],
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
      k8s_tls_config(config) {
        job_name: 'kube-system/kubelet',
        kubernetes_sd_configs: [{
          role: 'node',
        }],

        relabel_configs: [
          {
            target_label: '__address__',
            replacement: config.kubernetes_api_server_address,
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

      // As of k8s 1.7.3, cAdvisor metrics are available via kubelet using the
      // /metrics/cadvisor path.
      k8s_tls_config(config) {
        job_name: 'kube-system/cadvisor',
        kubernetes_sd_configs: [{
          role: 'node',
        }],
        scheme: 'https',

        relabel_configs: [
          {
            target_label: '__address__',
            replacement: config.kubernetes_api_server_address,
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
  },
}
