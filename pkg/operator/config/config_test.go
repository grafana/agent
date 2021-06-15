package config

import (
	"fmt"
	"testing"

	"github.com/grafana/agent/pkg/operator/assets"
	"github.com/grafana/agent/pkg/util"
	prom_v1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8s_yaml "sigs.k8s.io/yaml"

	grafana "github.com/grafana/agent/pkg/operator/apis/monitoring/v1alpha1"
)

func TestBuildConfig(t *testing.T) {
	var store = make(assets.SecretStore)

	store[assets.Key("/secrets/default/example-secret/key")] = "somesecret"
	store[assets.Key("/configMaps/default/example-cm/key")] = "somecm"

	tt := []struct {
		input  string
		expect string
	}{
		{
			input: util.Untab(`
				metadata:
					name: example
					namespace: default
				spec:
					logLevel: debug
					prometheus:
						scrapeInterval: 15s
						scrapeTimeout: 10s
						externalLabels:
							cluster: prod
							foo: bar
						remoteWrite:
						- name: rw-1
							url: http://localhost:9090/api/v1/write
			`),
			expect: util.Untab(`
				server:
					http_listen_port: 8080
					log_level: debug

				prometheus:
					wal_directory: /var/lib/grafana-agent/data
					global:
						scrape_interval: 15s
						scrape_timeout: 10s
						external_labels:
							cluster: prod
							foo: bar
							__replica__: replica-$(STATEFULSET_ORDINAL_NUMBER)
						remote_write:
						- name: rw-1
							url: http://localhost:9090/api/v1/write
			`),
		},
		{
			input: util.Untab(`
					metadata:
						name: example
						namespace: default
					spec:
						logLevel: debug
						prometheus:
							scrapeInterval: 15s
							scrapeTimeout: 10s
							externalLabels:
								cluster: prod
								foo: bar
							remoteWrite:
							- url: http://localhost:9090/api/v1/write
								basicAuth:
									username:
										name: example-secret
										key: key
									password:
										name: example-secret
										key: key
								tlsConfig:
									ca:
										configMap:
											name:	example-cm
											key: key
									cert:
										secret:
											name: example-secret
											key: key
									keySecret:
										name: example-secret
										key: key
				`),
			expect: util.Untab(`
					server:
						http_listen_port: 8080
						log_level: debug

					prometheus:
						wal_directory: /var/lib/grafana-agent/data
						global:
							scrape_interval: 15s
							scrape_timeout: 10s
							external_labels:
								cluster: prod
								foo: bar
								__replica__: replica-$(STATEFULSET_ORDINAL_NUMBER)
							remote_write:
							- url: http://localhost:9090/api/v1/write
								basic_auth:
									username: somesecret
									password: somesecret
								tls_config:
									ca_file: /var/lib/grafana-agent/secrets/_configMaps_default_example_cm_key
									cert_file: /var/lib/grafana-agent/secrets/_secrets_default_example_secret_key
									key_file: /var/lib/grafana-agent/secrets/_secrets_default_example_secret_key
				`),
		},
	}

	for i, tc := range tt {
		t.Run(fmt.Sprintf("index_%d", i), func(t *testing.T) {
			var spec grafana.GrafanaAgent
			err := k8s_yaml.Unmarshal([]byte(tc.input), &spec)
			require.NoError(t, err)

			d := Deployment{Agent: &spec}
			result, err := d.BuildConfig(store)
			require.NoError(t, err)

			if !assert.YAMLEq(t, tc.expect, result) {
				fmt.Println(result)
			}
		})
	}
}

// TestFullConfig tests generation of a config that may be used to collect
// metrics from Kubernetes.
func TestFullConfig(t *testing.T) {
	var store = make(assets.SecretStore)

	input := Deployment{
		Agent: &grafana.GrafanaAgent{
			TypeMeta: v1.TypeMeta{
				APIVersion: "monitoring.grafana.com/v1alpha1",
				Kind:       "GrafanaAgent",
			},
			ObjectMeta: v1.ObjectMeta{
				Name:      "agent",
				Namespace: "operator",
				Labels:    map[string]string{"app": "grafana-agent"},
			},
			Spec: grafana.GrafanaAgentSpec{
				Image:              strPointer("grafana/agent:latest"),
				ServiceAccountName: "agent",
				Prometheus: grafana.PrometheusSubsystemSpec{
					InstanceSelector: &v1.LabelSelector{
						MatchLabels: map[string]string{"agent": "agent"},
					},
				},
			},
		},
		Prometheis: []PrometheusInstance{{
			Instance: &grafana.PrometheusInstance{
				TypeMeta: v1.TypeMeta{
					APIVersion: "monitoring.grafana.com/v1alpha1",
					Kind:       "PrometheusInstance",
				},
				ObjectMeta: v1.ObjectMeta{
					Name:      "primary",
					Namespace: "operator",
					Labels: map[string]string{
						"agent": "agent",
						"app":   "grafana-agent",
					},
				},
				Spec: grafana.PrometheusInstanceSpec{
					RemoteWrite: []grafana.RemoteWriteSpec{{
						URL: "http://cortex:80/api/prom/push",
					}},
					ServiceMonitorSelector: &v1.LabelSelector{
						MatchLabels: map[string]string{"instance": "primary"},
					},
					PodMonitorSelector: &v1.LabelSelector{
						MatchLabels: map[string]string{"instance": "primary"},
					},
					ProbeSelector: &v1.LabelSelector{
						MatchLabels: map[string]string{"instance": "primary"},
					},
				},
			},

			ServiceMonitors: []*prom_v1.ServiceMonitor{{
				TypeMeta: v1.TypeMeta{
					APIVersion: "monitoring.coreos.com/v1",
					Kind:       "ServiceMonitor",
				},
				ObjectMeta: v1.ObjectMeta{
					Name:      "kubernetes",
					Namespace: "default",
					Labels: map[string]string{
						"instance": "primary",
						"app":      "grafana-agent",
					},
				},
				Spec: prom_v1.ServiceMonitorSpec{
					Selector: v1.LabelSelector{
						MatchLabels: map[string]string{"component": "apiserver"},
					},
					Endpoints: []prom_v1.Endpoint{{
						Port:   "https",
						Scheme: "https",
						TLSConfig: &prom_v1.TLSConfig{
							SafeTLSConfig: prom_v1.SafeTLSConfig{
								ServerName: "kubernetes",
							},
						},
						MetricRelabelConfigs: []*prom_v1.RelabelConfig{{
							SourceLabels: []string{"__name__"},
							Regex:        "workqueue_queue_duration_seconds_bucket|process_cpu_seconds_total|process_resident_memory_bytes|workqueue_depth|rest_client_request_duration_seconds_bucket|workqueue_adds_total|up|rest_client_requests_total|apiserver_request_total|go_goroutines",
							Action:       "keep",
						}},
					}},
				},
			}},

			PodMonitors: []*prom_v1.PodMonitor{{
				TypeMeta: v1.TypeMeta{
					APIVersion: "monitoring.coreos.com/v1",
					Kind:       "PodMonitor",
				},
				ObjectMeta: v1.ObjectMeta{
					Name:      "kubernetes-pods",
					Namespace: "operator",
					Labels: map[string]string{
						"instance": "primary",
						"app":      "grafana-agent",
					},
				},
				Spec: prom_v1.PodMonitorSpec{
					NamespaceSelector: prom_v1.NamespaceSelector{Any: true},
					Selector: v1.LabelSelector{
						MatchExpressions: []v1.LabelSelectorRequirement{{
							Key:      "name",
							Operator: v1.LabelSelectorOpExists,
						}},
					},
					PodMetricsEndpoints: []prom_v1.PodMetricsEndpoint{{
						Port: ".*-metrics",
						RelabelConfigs: []*prom_v1.RelabelConfig{
							{
								SourceLabels: []string{
									"__meta_kubernetes_namespace",
									"__meta_kubernetes_pod_label_name",
								},
								Action:      "replace",
								Separator:   "/",
								TargetLabel: "job",
								Replacement: "$1",
							},
							{
								SourceLabels: []string{
									"__meta_kubernetes_pod_name",
									"__meta_kubernetes_pod_container_name",
									"__meta_kubernetes_pod_container_port_name",
								},
								Action:      "replace",
								Separator:   ":",
								TargetLabel: "instance",
							},
						},
					}},
				},
			}},
		}},
	}

	expect := util.Untab(`
server:
  http_listen_port: 8080

prometheus:
  wal_directory: /var/lib/grafana-agent/data
  global:
    external_labels:
      __replica__: replica-$(STATEFULSET_ORDINAL_NUMBER)
      cluster: operator/agent
  configs:
  - name: operator/primary
    remote_write:
    - url: http://cortex:80/api/prom/push
    scrape_configs:
    - honor_labels: false
      job_name: serviceMonitor/default/kubernetes/0
      kubernetes_sd_configs:
      - namespaces:
          names:
          - default
        role: endpoints
      metric_relabel_configs:
      - action: keep
        regex: workqueue_queue_duration_seconds_bucket|process_cpu_seconds_total|process_resident_memory_bytes|workqueue_depth|rest_client_request_duration_seconds_bucket|workqueue_adds_total|up|rest_client_requests_total|apiserver_request_total|go_goroutines
        source_labels:
        - __name__
      relabel_configs:
      - source_labels:
        - job
        target_label: __tmp_prometheus_job_name
      - action: keep
        regex: apiserver
        source_labels:
        - __meta_kubernetes_service_label_component
      - action: keep
        regex: https
        source_labels:
        - __meta_kubernetes_endpoint_port_name
      - regex: Node;(.*)
        replacement: $1
        separator: ;
        source_labels:
        - __meta_kubernetes_endpoint_address_target_kind
        - __meta_kubernetes_endpoint_address_target_name
        target_label: node
      - regex: Pod;(.*)
        replacement: $1
        separator: ;
        source_labels:
        - __meta_kubernetes_endpoint_address_target_kind
        - __meta_kubernetes_endpoint_address_target_name
        target_label: pod
      - source_labels:
        - __meta_kubernetes_namespace
        target_label: namespace
      - source_labels:
        - __meta_kubernetes_service_name
        target_label: service
      - source_labels:
        - __meta_kubernetes_pod_name
        target_label: pod
      - source_labels:
        - __meta_kubernetes_pod_container_name
        target_label: container
      - replacement: $1
        source_labels:
        - __meta_kubernetes_service_name
        target_label: job
      - replacement: https
        target_label: endpoint
      - action: hashmod
        modulus: 1
        source_labels:
        - __address__
        target_label: __tmp_hash
      - action: keep
        regex: $(SHARD)
        source_labels:
        - __tmp_hash
      scheme: https
      tls_config:
        server_name: kubernetes
    - honor_labels: false
      job_name: podMonitor/operator/kubernetes-pods/0
      kubernetes_sd_configs:
      - role: pod
      relabel_configs:
      - source_labels:
        - job
        target_label: __tmp_prometheus_job_name
      - action: keep
        regex: "true"
        source_labels:
        - __meta_kubernetes_pod_labelpresent_name
      - action: keep
        regex: .*-metrics
        source_labels:
        - __meta_kubernetes_pod_container_port_name
      - source_labels:
        - __meta_kubernetes_namespace
        target_label: namespace
      - source_labels:
        - __meta_kubernetes_service_name
        target_label: service
      - source_labels:
        - __meta_kubernetes_pod_name
        target_label: pod
      - source_labels:
        - __meta_kubernetes_pod_container_name
        target_label: container
      - replacement: operator/kubernetes-pods
        target_label: job
      - replacement: .*-metrics
        target_label: endpoint
      - action: replace
        replacement: $1
        separator: /
        source_labels:
        - __meta_kubernetes_namespace
        - __meta_kubernetes_pod_label_name
        target_label: job
      - action: replace
        separator: ':'
        source_labels:
        - __meta_kubernetes_pod_name
        - __meta_kubernetes_pod_container_name
        - __meta_kubernetes_pod_container_port_name
        target_label: instance
      - action: hashmod
        modulus: 1
        source_labels:
        - __address__
        target_label: __tmp_hash
      - action: keep
        regex: $(SHARD)
        source_labels:
        - __tmp_hash
	`)

	result, err := input.BuildConfig(store)
	require.NoError(t, err)

	if !assert.YAMLEq(t, expect, result) {
		fmt.Println(result)
	}
}

func strPointer(s string) *string { return &s }
