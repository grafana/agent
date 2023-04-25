package configgen

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/grafana/agent/component/common/kubernetes"
	"github.com/grafana/agent/pkg/util"
	promopv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	commonConfig "github.com/prometheus/common/config"
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/config"
	"github.com/prometheus/prometheus/discovery"
	promk8s "github.com/prometheus/prometheus/discovery/kubernetes"
	"github.com/prometheus/prometheus/model/relabel"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestGenerateServiceMonitorConfig(t *testing.T) {
	//var falseVal = false
	//var proxyURL = "https://proxy:8080"
	suite := []struct {
		name                   string
		m                      *promopv1.ServiceMonitor
		ep                     promopv1.Endpoint
		expectedRelabels       string
		expectedMetricRelabels string
		expected               *config.ScrapeConfig
	}{
		{
			name: "default",
			m: &promopv1.ServiceMonitor{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "operator",
					Name:      "svcmonitor",
				},
			},
			ep: promopv1.Endpoint{
				Port: "metrics",
			},
			expectedRelabels: util.Untab(`
				- source_labels: [job]
				  target_label: __tmp_prometheus_job_name
				- source_labels: [__meta_kubernetes_endpoint_port_name]
				  regex: metrics
				  action: keep
				- source_labels: [__meta_kubernetes_endpoint_address_target_kind, __meta_kubernetes_endpoint_address_target_name]
				  regex: Node;(.*)
				  target_label: node
				  replacement: ${1}
				- source_labels: [__meta_kubernetes_endpoint_address_target_kind, __meta_kubernetes_endpoint_address_target_name]
				  regex: Pod;(.*)
				  target_label: pod
				  action: replace
				  replacement: ${1}
				- source_labels: [__meta_kubernetes_namespace]
				  target_label: namespace
				- source_labels: [__meta_kubernetes_service_name]
				  target_label: service
				- source_labels: [__meta_kubernetes_pod_container_name]
				  target_label: container
				- source_labels: [__meta_kubernetes_pod_name]
				  target_label: pod
				- source_labels: [__meta_kubernetes_pod_phase]
				  regex: (Failed|Succeeded)
				  action: drop
				- source_labels: [__meta_kubernetes_service_name]
				  target_label: job
				  replacement: ${1}
				- target_label: endpoint
				  replacement: metrics
				  action: replace
			`),
			expected: &config.ScrapeConfig{
				JobName:         "serviceMonitor/operator/svcmonitor/0",
				HonorTimestamps: true,
				ScrapeInterval:  model.Duration(time.Minute),
				ScrapeTimeout:   model.Duration(10 * time.Second),
				MetricsPath:     "/metrics",
				Scheme:          "http",
				HTTPClientConfig: commonConfig.HTTPClientConfig{
					FollowRedirects: true,
					EnableHTTP2:     true,
				},
				ServiceDiscoveryConfigs: discovery.Configs{
					&promk8s.SDConfig{
						Role: "endpoints",

						NamespaceDiscovery: promk8s.NamespaceDiscovery{
							IncludeOwnNamespace: false,
							Names:               []string{"operator"},
						},
					},
				},
			},
		},
		{
			name: "everything",
			m: &promopv1.ServiceMonitor{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "operator",
					Name:      "svcmonitor",
				},
				Spec: promopv1.ServiceMonitorSpec{
					JobLabel:        "joblabelispecify",
					TargetLabels:    []string{"a", "b"},
					PodTargetLabels: []string{"c", "d"},
				},
			},
			ep: promopv1.Endpoint{
				Port: "metrics",
			},
			expectedRelabels: util.Untab(`
				- source_labels: [job]
				  target_label: __tmp_prometheus_job_name
				- source_labels: [__meta_kubernetes_endpoint_port_name]
				  regex: metrics
				  action: keep
				- source_labels: [__meta_kubernetes_endpoint_address_target_kind, __meta_kubernetes_endpoint_address_target_name]
				  regex: Node;(.*)
				  target_label: node
				  replacement: ${1}
				- source_labels: [__meta_kubernetes_endpoint_address_target_kind, __meta_kubernetes_endpoint_address_target_name]
				  regex: Pod;(.*)
				  target_label: pod
				  action: replace
				  replacement: ${1}
				- source_labels: [__meta_kubernetes_namespace]
				  target_label: namespace
				- source_labels: [__meta_kubernetes_service_name]
				  target_label: service
				- source_labels: [__meta_kubernetes_pod_container_name]
				  target_label: container
				- source_labels: [__meta_kubernetes_pod_name]
				  target_label: pod
				- source_labels: [__meta_kubernetes_pod_phase]
				  regex: (Failed|Succeeded)
				  action: drop
				- regex: "(.+)"
				  replacement: ${1}
				  source_labels: [__meta_kubernetes_service_label_a]
				  target_label: a
				- regex: "(.+)"
				  replacement: ${1}
				  source_labels: [__meta_kubernetes_service_label_b]
				  target_label: b
				- regex: "(.+)"
				  replacement: ${1}
				  source_labels: [__meta_kubernetes_pod_label_c]
				  target_label: c
				- regex: "(.+)"
				  replacement: ${1}
				  source_labels: [__meta_kubernetes_pod_label_d]
				  target_label: d
				- source_labels: [__meta_kubernetes_service_name]
				  target_label: job
				  replacement: ${1}
				- source_labels: [__meta_kubernetes_service_label_joblabelispecify]
				  target_label: job
				  regex: "(.+)"
				  replacement: ${1}
				- target_label: endpoint
				  replacement: metrics
				  action: replace
			`),
			expected: &config.ScrapeConfig{
				JobName:         "serviceMonitor/operator/svcmonitor/1",
				HonorTimestamps: true,
				ScrapeInterval:  model.Duration(time.Minute),
				ScrapeTimeout:   model.Duration(10 * time.Second),
				MetricsPath:     "/metrics",
				Scheme:          "http",
				HTTPClientConfig: commonConfig.HTTPClientConfig{
					FollowRedirects: true,
					EnableHTTP2:     true,
				},
				ServiceDiscoveryConfigs: discovery.Configs{
					&promk8s.SDConfig{
						Role: "endpoints",

						NamespaceDiscovery: promk8s.NamespaceDiscovery{
							IncludeOwnNamespace: false,
							Names:               []string{"operator"},
						},
					},
				},
			},
		},
	}
	for i, tc := range suite {
		t.Run(tc.name, func(t *testing.T) {
			cg := &ConfigGenerator{Client: &kubernetes.ClientArguments{}}
			cfg, err := cg.GenerateServiceMonitorConfig(tc.m, tc.ep, i)
			require.NoError(t, err)
			// check relabel configs separately
			rlcs := cfg.RelabelConfigs
			mrlcs := cfg.MetricRelabelConfigs
			cfg.RelabelConfigs = nil
			cfg.MetricRelabelConfigs = nil
			require.NoError(t, err)

			assert.Equal(t, tc.expected, cfg)

			checkRelabels := func(actual []*relabel.Config, expected string) {
				// load the expected relabel rules as yaml so we get the defaults put in there.
				ex := []*relabel.Config{}
				err := yaml.Unmarshal([]byte(expected), &ex)
				require.NoError(t, err)
				y, err := yaml.Marshal(ex)
				require.NoError(t, err)
				expected = string(y)

				y, err = yaml.Marshal(actual)
				require.NoError(t, err)

				if !assert.YAMLEq(t, expected, string(y)) {
					fmt.Fprintln(os.Stderr, string(y))
					fmt.Fprintln(os.Stderr, expected)
				}
			}
			checkRelabels(rlcs, tc.expectedRelabels)
			checkRelabels(mrlcs, tc.expectedMetricRelabels)
		})
	}
}
