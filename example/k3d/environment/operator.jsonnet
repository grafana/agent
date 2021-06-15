local k = import 'ksonnet-util/kausal.libsonnet';

local images = {
  operator: 'grafana/agent-operator:latest',
  agent: 'grafana/agent:latest',
};

local configMap = k.core.v1.configMap;
local container = k.core.v1.container;
local daemonSet = k.apps.v1.daemonSet;
local deployment = k.apps.v1.deployment;
local policyRule = k.rbac.v1.policyRule;
local serviceAccount = k.core.v1.serviceAccount;
local namespace = k.core.v1.namespace;
local secret = k.core.v1.secret;

{
  local operator_namespace = 'operator',
  local k = (import 'ksonnet-util/kausal.libsonnet') {
    _config+:: { namespace: operator_namespace },
  },

  // CRDs for the operator.
  crds: std.map(std.native('parseYaml'), [
    importstr '../../../production/operator/crds/monitoring.coreos.com_probes.yaml',
    importstr '../../../production/operator/crds/monitoring.coreos.com_podmonitors.yaml',
    importstr '../../../production/operator/crds/monitoring.coreos.com_servicemonitors.yaml',
    importstr '../../../production/operator/crds/monitoring.grafana.com_grafana-agents.yaml',
    importstr '../../../production/operator/crds/monitoring.grafana.com_prometheus-instances.yaml',
  ]),

  // Deployment of the operator itself.
  operator: {
    namespace: namespace.new(operator_namespace),

    container:: container.new('operator', images.operator),

    deployment:
      deployment.new('operator', 1, [self.container]) +
      deployment.mixin.metadata.withNamespace('operator') +
      deployment.mixin.spec.template.spec.withServiceAccount('operator'),

    local verbs_read = ['get', 'list', 'watch'],
    local verbs_all = ['get', 'list', 'watch', 'create', 'update', 'patch', 'delete'],

    rbac: k.util.rbac('operator', [
      // Read resources

      policyRule.withApiGroups(['monitoring.grafana.com']) +
      policyRule.withResources(['grafana-agents', 'prometheus-instances']) +
      policyRule.withVerbs(verbs_read),

      policyRule.withApiGroups(['monitoring.coreos.com']) +
      policyRule.withResources(['podmonitors', 'probes', 'servicemonitors']) +
      policyRule.withVerbs(verbs_read),

      policyRule.withApiGroups(['']) +
      policyRule.withResources(['namespaces']) +
      policyRule.withVerbs(verbs_read),

      // Read + written resources

      policyRule.withApiGroups(['']) +
      policyRule.withResources(['secrets', 'services']) +
      policyRule.withVerbs(verbs_all),

      policyRule.withApiGroups(['apps']) +
      policyRule.withResources(['statefulsets']) +
      policyRule.withVerbs(verbs_all),
    ]) {
      service_account+:
        serviceAccount.mixin.metadata.withNamespace(operator_namespace),
    },
  },

  // Deployment of the Agent.
  operator_agent_deployment: {
    rbac: k.util.rbac('agent', [
      policyRule.withApiGroups(['']) +
      policyRule.withResources(['nodes', 'nodes/proxy', 'services', 'endpoints', 'pods']) +
      policyRule.withVerbs(['get', 'list', 'watch']),

      policyRule.withNonResourceURLs(['/metrics']) +
      policyRule.withVerbs(['get']),
    ]) {
      service_account+:
        serviceAccount.mixin.metadata.withNamespace(operator_namespace),
    },

    agent: {
      apiVersion: 'monitoring.grafana.com/v1alpha1',
      kind: 'GrafanaAgent',
      metadata: {
        name: 'agent',
        namespace: operator_namespace,
        labels: { name: 'grafana-agent' },
      },
      spec: {
        image: images.agent,
        serviceAccountName: 'agent',
        podMetadata: {
          labels: {
            name: 'grafana-agent',
          },
        },
        prometheus: {
          instanceSelector: {
            matchLabels: { agent: 'agent' },
          },
        },
      },
    },

    writer: {
      apiVersion: 'monitoring.grafana.com/v1alpha1',
      kind: 'PrometheusInstance',
      metadata: {
        name: 'primary',
        namespace: operator_namespace,
        labels: { agent: 'agent', name: 'grafana-agent' },
      },
      spec: {
        remoteWrite: [{
          url: 'http://cortex.default.svc.cluster.local/api/prom/push',
        }],

        local instanceSelector = { matchLabels: { instance: 'primary' } },
        serviceMonitorSelector: instanceSelector,
        podMonitorSelector: instanceSelector,
        probeSelector: instanceSelector,

        // Load in role: node jobs from a secret
        additionalScrapeConfigs: {
          name: 'system-jobs',
          key: 'jobs.yaml',
        },
      },
    },

    jobs: [
      // Collect from Kubernetes
      {
        apiVersion: 'monitoring.coreos.com/v1',
        kind: 'ServiceMonitor',
        metadata: {
          name: 'kubernetes',
          namespace: operator_namespace,
          labels: { instance: 'primary', name: 'grafana-agent' },
        },
        spec: {
          namespaceSelector: { matchNames: ['default'] },
          selector: {
            matchLabels: { component: 'apiserver' },
          },
          endpoints: [{
            port: 'https',
            scheme: 'https',
            tlsConfig: {
              caFile: '/var/run/secrets/kubernetes.io/serviceaccount/ca.crt',
              serverName: 'kubernetes',
            },
            bearerTokenFile: '/var/run/secrets/kubernetes.io/serviceaccount/token',
            metricRelabelings: [{
              sourceLabels: ['__name__'],
              regex: 'workqueue_queue_duration_seconds_bucket|process_cpu_seconds_total|process_resident_memory_bytes|workqueue_depth|rest_client_request_duration_seconds_bucket|workqueue_adds_total|up|rest_client_requests_total|apiserver_request_total|go_goroutines',
              action: 'keep',
            }],
            relabelings: [{
              targetLabel: 'job',
              replacement: 'default/kubernetes',
            }],
          }],
        },
      },
      // Collect from all pods
      {
        apiVersion: 'monitoring.coreos.com/v1',
        kind: 'PodMonitor',
        metadata: {
          name: 'kubernetes-pods',
          namespace: operator_namespace,
          labels: { instance: 'primary', name: 'grafana-agent' },
        },
        spec: {
          selector: {
            matchExpressions: [{
              key: 'name',
              operator: 'Exists',
            }],
          },
          namespaceSelector: { any: true },
          podMetricsEndpoints: [{
            port: '.*-metrics',
            relabelings: [
              {
                sourceLabels: ['__meta_kubernetes_namespace', '__meta_kubernetes_pod_label_name'],
                action: 'replace',
                separator: '/',
                targetLabel: 'job',
                replacement: '$1',
              },
              {
                // Rename instances to the concatenation of pod:container:port.
                // All three components are needed to guarantee a unique instance label.
                sourceLabels: [
                  '__meta_kubernetes_pod_name',
                  '__meta_kubernetes_pod_container_name',
                  '__meta_kubernetes_pod_container_port_name',
                ],
                action: 'replace',
                separator: ':',
                targetLabel: 'instance',
              },
            ],
          }],
        },
      },
    ],

    system_jobs: secret.new('system-jobs', {
      'jobs.yaml': std.base64(|||
        - job_name: kube-system/kubelet
          kubernetes_sd_configs:
          - role: 'node'
          tls_config:
            ca_file: /var/run/secrets/kubernetes.io/serviceaccount/ca.crt
          bearer_token_file: /var/run/secrets/kubernetes.io/serviceaccount/token
          scheme: https
          relabel_configs:
          - target_label: __address__
            replacement: kubernetes.default.svc.cluster.local:443
          - source_labels: [__meta_kubernetes_node_name]
            regex: (.+)
            target_label: __metrics_path__
            replacement: /api/v1/nodes/$1/proxy/metrics
          - action: hashmod
            modulus: $(SHARDS)
            source_labels:
            - __address__
            target_label: __tmp_hash
          - action: keep
            regex: $(SHARD)
            source_labels:
            - __tmp_hash

        - job_name: kube-system/cadvisor
          kubernetes_sd_configs:
          - role: 'node'
          tls_config:
            ca_file: /var/run/secrets/kubernetes.io/serviceaccount/ca.crt
          bearer_token_file: /var/run/secrets/kubernetes.io/serviceaccount/token
          scheme: https
          relabel_configs:
          - target_label: __address__
            replacement: kubernetes.default.svc.cluster.local:443
          - source_labels: [__meta_kubernetes_node_name]
            regex: (.+)
            target_label: __metrics_path__
            replacement: /api/v1/nodes/$1/proxy/metrics/cadvisor
          - action: hashmod
            modulus: $(SHARDS)
            source_labels:
            - __address__
            target_label: __tmp_hash
          - action: keep
            regex: $(SHARD)
            source_labels:
            - __tmp_hash
          metric_relabel_configs:
          - source_labels: ['__name__', 'image']
            regex: 'container_([a-z_]+);'
            action: 'drop'
          - source_labels: ['__name__']
            regex: 'container_(network_tcp_usage_total|network_udp_usage_total|tasks_state|cpu_load_average_10s)'
            action: 'drop'
      |||),
    }) + secret.metadata.withNamespace(operator_namespace),
  },
}
