package configgen

import (
	"fmt"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/grafana/agent/component/common/kubernetes"
	flow_relabel "github.com/grafana/agent/component/common/relabel"
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
	"k8s.io/apimachinery/pkg/util/intstr"
)

func TestGenerateServiceMonitorConfig(t *testing.T) {
	var falseVal = false
	var proxyURL = "https://proxy:8080"
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
			ep: promopv1.Endpoint{},
			expectedRelabels: util.Untab(`
				- target_label: __meta_foo
				  replacement: bar
				- source_labels: [job]
				  target_label: __tmp_prometheus_job_name
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
		{
			name: "targetport_string",
			m: &promopv1.ServiceMonitor{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "operator",
					Name:      "svcmonitor",
				},
			},
			ep: promopv1.Endpoint{
				TargetPort: &intstr.IntOrString{StrVal: "http_metrics", Type: intstr.String},
			},
			expectedRelabels: util.Untab(`
				- target_label: __meta_foo
				  replacement: bar
				- source_labels: [job]
				  target_label: __tmp_prometheus_job_name
				- source_labels: ["__meta_kubernetes_pod_container_port_name"]
				  regex: "http_metrics"
				  action: "keep"
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
				  replacement: http_metrics
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
		{
			name: "targetport_int",
			m: &promopv1.ServiceMonitor{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "operator",
					Name:      "svcmonitor",
				},
			},
			ep: promopv1.Endpoint{
				TargetPort: &intstr.IntOrString{IntVal: 4242, Type: intstr.Int},
			},
			expectedRelabels: util.Untab(`
				- target_label: __meta_foo
				  replacement: bar
				- source_labels: [job]
				  target_label: __tmp_prometheus_job_name
				- source_labels: ["__meta_kubernetes_pod_container_port_number"]
				  regex: "4242"
				  action: "keep"
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
				  replacement: "4242"
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
					Selector: metav1.LabelSelector{
						MatchLabels: map[string]string{"foo": "bar"},
						MatchExpressions: []metav1.LabelSelectorRequirement{
							{
								Key:      "key",
								Operator: metav1.LabelSelectorOpIn,
								Values:   []string{"val0", "val1"},
							},
							{
								Key:      "key",
								Operator: metav1.LabelSelectorOpNotIn,
								Values:   []string{"val2", "val3"},
							},
							{
								Key:      "key",
								Operator: metav1.LabelSelectorOpExists,
							},
							{
								Key:      "key2",
								Operator: metav1.LabelSelectorOpDoesNotExist,
							},
						},
					},
					NamespaceSelector:     promopv1.NamespaceSelector{Any: false, MatchNames: []string{"ns_a", "ns_b"}},
					SampleLimit:           101,
					TargetLimit:           102,
					LabelLimit:            103,
					LabelNameLengthLimit:  104,
					LabelValueLengthLimit: 105,
					AttachMetadata:        &promopv1.AttachMetadata{Node: true},
				},
			},
			ep: promopv1.Endpoint{
				Port:            "metrics",
				EnableHttp2:     &falseVal,
				Path:            "/foo",
				Params:          map[string][]string{"a": {"b"}},
				FollowRedirects: &falseVal,
				ProxyURL:        &proxyURL,
				Scheme:          "https",
				ScrapeTimeout:   "17m",
				Interval:        "1s",
				HonorLabels:     true,
				HonorTimestamps: &falseVal,
				FilterRunning:   &falseVal,
				TLSConfig: &promopv1.TLSConfig{
					SafeTLSConfig: promopv1.SafeTLSConfig{
						ServerName:         "foo.com",
						InsecureSkipVerify: true,
					},
				},
				RelabelConfigs: []*promopv1.RelabelConfig{
					{
						SourceLabels: []promopv1.LabelName{"foo"},
						TargetLabel:  "bar",
					},
				},
			},
			expectedRelabels: util.Untab(`
				- target_label: __meta_foo
				  replacement: bar
				- source_labels: [job]
				  target_label: __tmp_prometheus_job_name
				- source_labels: [__meta_kubernetes_service_label_foo, __meta_kubernetes_service_labelpresent_foo]
				  action: keep
				  regex: (bar);true
				- source_labels: [__meta_kubernetes_service_label_key, __meta_kubernetes_service_labelpresent_key]
				  action: keep
				  regex: (val0|val1);true
				- source_labels: [__meta_kubernetes_service_label_key, __meta_kubernetes_service_labelpresent_key]
				  action: drop
				  regex: (val2|val3);true
				- source_labels: [__meta_kubernetes_service_labelpresent_key]
				  action: keep
				  regex: true
				- source_labels: [__meta_kubernetes_service_labelpresent_key2]
				  action: drop
				  regex: true
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
				- target_label: bar
				  source_labels: [foo]
			`),
			expected: &config.ScrapeConfig{
				JobName:         "serviceMonitor/operator/svcmonitor/1",
				HonorLabels:     true,
				HonorTimestamps: false,
				Params: url.Values{
					"a": []string{"b"},
				},
				ScrapeInterval: model.Duration(time.Second),
				ScrapeTimeout:  model.Duration(17 * time.Minute),
				MetricsPath:    "/foo",
				Scheme:         "https",
				HTTPClientConfig: commonConfig.HTTPClientConfig{
					FollowRedirects: falseVal,
					EnableHTTP2:     false,
					TLSConfig: commonConfig.TLSConfig{
						ServerName:         "foo.com",
						InsecureSkipVerify: true,
					},
					ProxyConfig: commonConfig.ProxyConfig{
						ProxyURL: commonConfig.URL{URL: &url.URL{Scheme: "https", Host: "proxy:8080"}},
					},
				},
				ServiceDiscoveryConfigs: discovery.Configs{
					&promk8s.SDConfig{
						Role:           "endpoints",
						AttachMetadata: promk8s.AttachMetadataConfig{Node: true},
						NamespaceDiscovery: promk8s.NamespaceDiscovery{
							IncludeOwnNamespace: false,
							Names:               []string{"ns_a", "ns_b"},
						},
					},
				},
				SampleLimit:           101,
				TargetLimit:           102,
				LabelLimit:            103,
				LabelNameLengthLimit:  104,
				LabelValueLengthLimit: 105,
			},
		},
	}
	for _, tc := range suite {
		t.Run(tc.name, func(t *testing.T) {
			cg := &ConfigGenerator{
				Client: &kubernetes.ClientArguments{},
				AdditionalRelabelConfigs: []*flow_relabel.Config{
					{TargetLabel: "__meta_foo", Replacement: "bar"},
				},
			}
			cfg, err := cg.GenerateServiceMonitorConfig(tc.m, tc.ep, 1)
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
