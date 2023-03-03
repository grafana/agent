package operator

import (
	"testing"
	"time"

	"github.com/grafana/agent/pkg/util"
	v1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	commonConfig "github.com/prometheus/common/config"
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/config"
	"github.com/prometheus/prometheus/discovery"
	promk8s "github.com/prometheus/prometheus/discovery/kubernetes"
	"github.com/prometheus/prometheus/model/relabel"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestGeneratePodMonitorConfig(t *testing.T) {
	//var falseVal = false
	suite := []struct {
		name                   string
		m                      *v1.PodMonitor
		ep                     v1.PodMetricsEndpoint
		args                   Arguments
		expectedRelabels       string
		expectedMetricRelabels string
		expected               *config.ScrapeConfig
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
			expectedRelabels: util.Untab(`
				- source_labels: [job]
				  target_label: __tmp_prometheus_job_name
				- source_labels: [__meta_kubernetes_pod_phase]
				  regex: (Failed|Succeeded)
				  action: drop
				- source_labels: [__meta_kubernetes_pod_container_port_name]
				  regex: metrics
				  action: keep
				- source_labels: [__meta_kubernetes_namespace]
				  target_label: namespace
				- source_labels: [__meta_kubernetes_pod_container_name]
				  target_label: container
				- source_labels: [__meta_kubernetes_pod_name]
				  target_label: pod
				- target_label: job
				  replacement: operator/podmonitor
				- target_label: endpoint
				  replacement: metrics
			`),
			expected: &config.ScrapeConfig{
				JobName:         "podMonitor/operator/podmonitor/0",
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
						Role: "pod",

						NamespaceDiscovery: promk8s.NamespaceDiscovery{
							IncludeOwnNamespace: false,
							Names:               []string{"operator"},
						},
					},
				},
			},
		},
		// {
		// 	name: "everything",
		// 	m: &v1.PodMonitor{
		// 		ObjectMeta: meta_v1.ObjectMeta{
		// 			Namespace: "operator",
		// 			Name:      "podmonitor",
		// 		},
		// 		Spec: v1.PodMonitorSpec{
		// 			JobLabel:        "abc",
		// 			PodTargetLabels: []string{"label_a", "label_b"},
		// 			Selector: meta_v1.LabelSelector{
		// 				MatchLabels: map[string]string{"foo": "bar"},
		// 				// TODO: test a variety of matchexpressions
		// 			},
		// 			NamespaceSelector: v1.NamespaceSelector{Any: false, MatchNames: []string{"ns_a", "ns_b"}},
		// 		},
		// 	},
		// 	ep: v1.PodMetricsEndpoint{
		// 		Port:        "metrics",
		// 		EnableHttp2: &falseVal,
		// 	},
		// },
	}
	for i, tc := range suite {
		t.Run(tc.name, func(t *testing.T) {
			cg := &configGenerator{
				config: &tc.args,
			}
			cfg, err := cg.generatePodMonitorConfig(tc.m, tc.ep, i)
			// check relabel configs seperately
			rlcs := cfg.RelabelConfigs
			mrlcs := cfg.MetricRelabelConfigs
			cfg.RelabelConfigs = nil
			cfg.MetricRelabelConfigs = nil
			require.NoError(t, err)

			require.Equal(t, tc.expected, cfg)

			// load expected relabel configs
			exRlcs := []*relabel.Config{}
			err = yaml.Unmarshal([]byte(tc.expectedRelabels), &exRlcs)
			require.NoError(t, err)
			assert.Equal(t, exRlcs, rlcs)

			if tc.expectedMetricRelabels == "" {
				assert.Nil(t, mrlcs)
			} else {
				exMrlcs := []*relabel.Config{}
				err = yaml.Unmarshal([]byte(tc.expectedMetricRelabels), &exMrlcs)
				require.NoError(t, err)
				assert.Equal(t, exMrlcs, mrlcs)
			}
		})
	}
}
