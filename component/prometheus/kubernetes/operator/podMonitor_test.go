package operator

import (
	"fmt"
	"os"
	"testing"

	"github.com/grafana/agent/pkg/util"
	v1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestGeneratePodMonitorConfig(t *testing.T) {
	var falseVal = false
	suite := []struct {
		name   string
		m      *v1.PodMonitor
		ep     v1.PodMetricsEndpoint
		args   Arguments
		expect string
	}{
		{
			name: "default",
			m: &v1.PodMonitor{
				ObjectMeta: meta_v1.ObjectMeta{
					Namespace: "operator",
					Name:      "podmonitor",
				},
			},
			ep: v1.PodMetricsEndpoint{
				Port: "metrics",
			},
			expect: util.Untab(`
			job_name: podMonitor/operator/podmonitor/0
			honor_timestamps: true
			scrape_interval: 1m
			scrape_timeout: 10s
			metrics_path: /metrics
			scheme: http
			follow_redirects: true
			enable_http2: true
			relabel_configs:
				- source_labels: [job]
				  separator: ;
				  regex: (.*)
				  target_label: __tmp_prometheus_job_name
				  replacement: $1
				  action: replace
				- source_labels: [__meta_kubernetes_pod_phase]
				  separator: ;
				  regex: (Failed|Succeeded)
				  replacement: $1
				  action: drop
				- source_labels: [__meta_kubernetes_pod_container_port_name]
				  separator: ;
				  regex: metrics
				  replacement: $1
				  action: keep
				- source_labels: [__meta_kubernetes_namespace]
				  separator: ;
				  regex: (.*)
				  target_label: namespace
				  replacement: $1
				  action: replace
				- source_labels: [__meta_kubernetes_pod_container_name]
				  separator: ;
				  regex: (.*)
				  target_label: container
				  replacement: $1
				  action: replace
				- source_labels: [__meta_kubernetes_pod_name]
				  separator: ;
				  regex: (.*)
				  target_label: pod
				  replacement: $1
				  action: replace
				- separator: ;
				  regex: (.*)
				  target_label: job
				  replacement: operator/podmonitor
				  action: replace
				- separator: ;
				  regex: (.*)
				  target_label: endpoint
				  replacement: metrics
				  action: replace
			kubernetes_sd_configs:
				- role: pod
				  kubeconfig_file: ""
				  follow_redirects: false
				  enable_http2: false
				  namespaces:
				    own_namespace: false
				    names: [operator]
			`),
		},
		{
			name: "everything",
			m: &v1.PodMonitor{
				ObjectMeta: meta_v1.ObjectMeta{
					Namespace: "operator",
					Name:      "podmonitor",
				},
				Spec: v1.PodMonitorSpec{
					JobLabel:        "abc",
					PodTargetLabels: []string{"label_a", "label_b"},
					Selector: meta_v1.LabelSelector{
						MatchLabels: map[string]string{"foo": "bar"},
						// TODO: test a variety of matchexpressions
					},
					NamespaceSelector: v1.NamespaceSelector{Any: false, MatchNames: []string{"ns_a", "ns_b"}},
				},
			},
			ep: v1.PodMetricsEndpoint{
				Port:        "metrics",
				EnableHttp2: &falseVal,
			},
			expect: util.Untab(`
			job_name: podMonitor/operator/podmonitor/1
			honor_timestamps: true
			scrape_interval: 1m
			scrape_timeout: 10s
			metrics_path: /metrics
			scheme: http
			follow_redirects: true
			enable_http2: false
			relabel_configs:
				- source_labels: [job]
				  separator: ;
				  regex: (.*)
				  target_label: __tmp_prometheus_job_name
				  replacement: $1
				  action: replace
				- source_labels: [__meta_kubernetes_pod_phase]
				  separator: ;
				  regex: (Failed|Succeeded)
				  replacement: $1
				  action: drop
				- source_labels: [__meta_kubernetes_pod_label_foo, __meta_kubernetes_pod_labelpresent_foo]
				  separator: ;
				  regex: (bar);true
				  replacement: $1
				  action: keep
				- source_labels: [__meta_kubernetes_pod_container_port_name]
				  separator: ;
				  regex: metrics
				  replacement: $1
				  action: keep
				- source_labels: [__meta_kubernetes_namespace]
				  separator: ;
				  regex: (.*)
				  target_label: namespace
				  replacement: $1
				  action: replace
				- source_labels: [__meta_kubernetes_pod_container_name]
				  separator: ;
				  regex: (.*)
				  target_label: container
				  replacement: $1
				  action: replace
				- source_labels: [__meta_kubernetes_pod_name]
				  separator: ;
				  regex: (.*)
				  target_label: pod
				  replacement: $1
				  action: replace
				- source_labels: [__meta_kubernetes_pod_label_label_a]
				  separator: ;
				  regex: (.+)
				  target_label: label_a
				  replacement: ${1}
				  action: replace
				- source_labels: [__meta_kubernetes_pod_label_label_b]
				  separator: ;
				  regex: (.+)
				  target_label: label_b
				  replacement: ${1}
				  action: replace
				- separator: ;
				  regex: (.*)
				  target_label: job
				  replacement: operator/podmonitor
				  action: replace
				- source_labels: [__meta_kubernetes_pod_label_abc]
				  separator: ;
				  regex: (.+)
				  target_label: job
				  replacement: ${1}
				  action: replace
				- separator: ;
				  regex: (.*)
				  target_label: endpoint
				  replacement: metrics
				  action: replace
			kubernetes_sd_configs:
				- role: pod
				  kubeconfig_file: ""
				  follow_redirects: false
				  enable_http2: false
				  namespaces:
				    own_namespace: false
				    names: [ns_a,ns_b]
			`),
		},
	}
	for i, tc := range suite {
		t.Run(tc.name, func(t *testing.T) {

			cg := &configGenerator{
				config: &tc.args,
			}
			cfg, err := cg.generatePodMonitorConfig(tc.m, tc.ep, i)
			require.NoError(t, err)
			actual, err := yaml.Marshal(cfg)
			require.NoError(t, err)
			if !assert.YAMLEq(t, tc.expect, string(actual)) {
				fmt.Fprintln(os.Stderr, string(actual))
			}

		})
	}
}
