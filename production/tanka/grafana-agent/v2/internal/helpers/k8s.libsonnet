local k8s_tls_config(config) = {
  tls_config: {
    ca_file: '/var/run/secrets/kubernetes.io/serviceaccount/ca.crt',
    insecure_skip_verify: config.insecure_skip_verify,
  },
  bearer_token_file: '/var/run/secrets/kubernetes.io/serviceaccount/token',
};

local gen_scrape_config(job_name, pod_uid) = {
  job_name: job_name,
  pipeline_stages: [{
    docker: {},
  }],
  kubernetes_sd_configs: [{
    role: 'pod',
  }],

  relabel_configs: self.prelabel_config + [
    // Only scrape local pods; Promtail will drop targets with a __host__ label
    // that does not match the current host name.
    {
      source_labels: ['__meta_kubernetes_pod_node_name'],
      target_label: '__host__',
    },

    // Drop pods without a __service__ label.
    {
      source_labels: ['__service__'],
      action: 'drop',
      regex: '',
    },

    // Include all the other labels on the pod.
    // Perform this mapping before applying additional label replacement rules
    // to prevent a supplied label from overwriting any of the following labels.
    {
      action: 'labelmap',
      regex: '__meta_kubernetes_pod_label_(.+)',
    },

    // Rename jobs to be <namespace>/<service>.
    {
      source_labels: ['__meta_kubernetes_namespace', '__service__'],
      action: 'replace',
      separator: '/',
      target_label: 'job',
      replacement: '$1',
    },

    // But also include the namespace, pod, container as separate
    // labels. They uniquely identify a container. They are also
    // identical to the target labels configured in Prometheus
    // (but note that Loki does not use an instance label).
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

    // Kubernetes puts logs under subdirectories keyed pod UID and container_name.
    {
      source_labels: [pod_uid, '__meta_kubernetes_pod_container_name'],
      target_label: '__path__',
      separator: '/',
      replacement: '/var/log/pods/*$1/*.log',
    },
  ],
};

{
  metrics(config)::
    local _config = {
      scrape_api_server_endpoints: false,
      insecure_skip_verify: false,

      cluster_dns_tld: 'local',
      cluster_dns_suffix: 'cluster.' + self.cluster_dns_tld,
      kubernetes_api_server_address: 'kubernetes.default.svc.%(cluster_dns_suffix)s:443' % self,

      ksm_namespace: 'kube-system',
      node_exporter_namespace: 'kube-system',
    } + config;

    [
      k8s_tls_config(_config) {
        job_name: 'default/kubernetes',
        kubernetes_sd_configs: [{
          role: if _config.scrape_api_server_endpoints then 'endpoints' else 'service',
        }],
        scheme: 'https',
        tls_config+: {
          server_name: 'kubernetes',
        },

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
      // namespace label and instead takes the namespace label from the exported
      // timeseries. This prevents the exported namespace label from being
      // renamed to exported_namesapce and allows us to route alerts based on
      // namespace.
      {
        job_name: '%s/kube-state-metrics' % _config.ksm_namespace,
        kubernetes_sd_configs: [{
          role: 'pod',
          namespaces: {
            names: [_config.ksm_namespace],
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
      {
        job_name: '%s/node-exporter' % _config.node_exporter_namespace,
        kubernetes_sd_configs: [{
          role: 'pod',
          namespaces: {
            names: [_config.node_exporter_namespace],
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
      k8s_tls_config(_config) {
        job_name: 'kube-system/kubelet',
        kubernetes_sd_configs: [{ role: 'node' }],

        relabel_configs: [
          {
            target_label: '__address__',
            replacement: _config.kubernetes_api_server_address,
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

      // As of k8s 1.7.3, cAdvisor metrics are available via kubelet using the
      // /metrics/cadvisor path.
      k8s_tls_config(_config) {
        job_name: 'kube-system/cadvisor',
        kubernetes_sd_configs: [{
          role: 'node',
        }],
        scheme: 'https',

        relabel_configs: [
          {
            target_label: '__address__',
            replacement: _config.kubernetes_api_server_address,
          },
          {
            source_labels: ['__meta_kubernetes_node_name'],
            regex: '(.+)',
            target_label: '__metrics_path__',
            replacement: '/api/v1/nodes/$1/proxy/metrics/cadvisor',
          },
        ],

        metric_relabel_configs: [
          // Let system processes like kubelet survive the next rule by giving them a fake image.
          {
            source_labels: ['__name__', 'id'],
            regex: 'container_([a-z_]+);/system.slice/(.+)',
            target_label: 'image',
            replacement: '$2',
          },

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

  logs(config={}):: [
    // Scrape config to scrape any pods with a 'name' label.
    gen_scrape_config('kubernetes-pods-name', '__meta_kubernetes_pod_uid') {
      prelabel_config:: [
        // Use name label as __service__.
        {
          source_labels: ['__meta_kubernetes_pod_label_name'],
          target_label: '__service__',
        },
      ],
    },

    // Scrape config to scrape any pods with an 'app' label.
    gen_scrape_config('kubernetes-pods-app', '__meta_kubernetes_pod_uid') {
      prelabel_config:: [
        // Drop pods with a 'name' label.  They will have already been added by
        // the scrape_config that matches on the 'name' label
        {
          source_labels: ['__meta_kubernetes_pod_label_name'],
          action: 'drop',
          regex: '.+',
        },

        // Use app label as the __service__.
        {
          source_labels: ['__meta_kubernetes_pod_label_app'],
          target_label: '__service__',
        },
      ],
    },

    // Scrape config to scrape any pods with a direct controller (eg
    // StatefulSets).
    gen_scrape_config('kubernetes-pods-direct-controllers', '__meta_kubernetes_pod_uid') {
      prelabel_config:: [
        // Drop pods with a 'name' or 'app' label.  They will have already been added by
        // the scrape_config that matches above.
        {
          source_labels: ['__meta_kubernetes_pod_label_name', '__meta_kubernetes_pod_label_app'],
          separator: '',
          action: 'drop',
          regex: '.+',
        },

        // Drop pods with an indirect controller. eg Deployments create replicaSets
        // which then create pods.
        {
          source_labels: ['__meta_kubernetes_pod_controller_name'],
          action: 'drop',
          regex: '[0-9a-z-.]+-[0-9a-f]{8,10}',
        },

        // Use controller name as __service__.
        {
          source_labels: ['__meta_kubernetes_pod_controller_name'],
          target_label: '__service__',
        },
      ],
    },

    // Scrape config to scrape any pods with an indirect controller (eg
    // Deployments).
    gen_scrape_config('kubernetes-pods-indirect-controller', '__meta_kubernetes_pod_uid') {
      prelabel_config:: [
        // Drop pods with a 'name' or 'app' label.  They will have already been added by
        // the scrape_config that matches above.
        {
          source_labels: ['__meta_kubernetes_pod_label_name', '__meta_kubernetes_pod_label_app'],
          separator: '',
          action: 'drop',
          regex: '.+',
        },

        // Drop pods not from an indirect controller. eg StatefulSets, DaemonSets
        {
          source_labels: ['__meta_kubernetes_pod_controller_name'],
          regex: '[0-9a-z-.]+-[0-9a-f]{8,10}',
          action: 'keep',
        },

        // Put the indirect controller name into a temp label.
        {
          source_labels: ['__meta_kubernetes_pod_controller_name'],
          action: 'replace',
          regex: '([0-9a-z-.]+)-[0-9a-f]{8,10}',
          target_label: '__service__',
        },
      ],
    },

    // Scrape config to scrape any control plane static pods (e.g. kube-apiserver
    // etcd, kube-controller-manager & kube-scheduler)
    gen_scrape_config('kubernetes-pods-static', '__meta_kubernetes_pod_annotation_kubernetes_io_config_mirror') {
      prelabel_config:: [
        // Ignore pods that aren't mirror pods
        {
          action: 'drop',
          source_labels: ['__meta_kubernetes_pod_annotation_kubernetes_io_config_mirror'],
          regex: '',
        },

        // Static control plane pods usually have a component label that identifies them
        {
          action: 'replace',
          source_labels: ['__meta_kubernetes_pod_label_component'],
          target_label: '__service__',
        },
      ],
    },
  ],

  traces(config={}):: [
    {
      bearer_token_file: '/var/run/secrets/kubernetes.io/serviceaccount/token',
      job_name: 'kubernetes-pods',
      kubernetes_sd_configs: [{ role: 'pod' }],
      relabel_configs: [
        {
          action: 'replace',
          source_labels: ['__meta_kubernetes_namespace'],
          target_label: 'namespace',
        },
        {
          action: 'replace',
          source_labels: ['__meta_kubernetes_pod_name'],
          target_label: 'pod',
        },
        {
          action: 'replace',
          source_labels: ['__meta_kubernetes_pod_container_name'],
          target_label: 'container',
        },
      ],
      tls_config: {
        ca_file: '/var/run/secrets/kubernetes.io/serviceaccount/ca.crt',
        insecure_skip_verify: false,
      },
    },
  ],
}
