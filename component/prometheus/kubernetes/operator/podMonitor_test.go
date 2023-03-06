package operator

import (
	"fmt"
	"io/fs"
	"os"
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
	corev1 "k8s.io/api/core/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type fakeSecretStore struct{}

func (f fakeSecretStore) Open(name string) (fs.File, error) {
	panic("no need to open files in this test suite")
}

func (f fakeSecretStore) ReadSecret(ns, name, key string) (string, error) {
	return fmt.Sprintf("content of secret:%s/%s:%s", ns, name, key), nil
}

func (f fakeSecretStore) ReadConfigMap(ns, name, key string) (string, error) {
	return fmt.Sprintf("content of configmap:%s/%s:%s", ns, name, key), nil
}

func TestGeneratePodMonitorConfig(t *testing.T) {
	var falseVal = false
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
					NamespaceSelector:     v1.NamespaceSelector{Any: false, MatchNames: []string{"ns_a", "ns_b"}},
					SampleLimit:           101,
					TargetLimit:           102,
					LabelLimit:            103,
					LabelNameLengthLimit:  104,
					LabelValueLengthLimit: 105,
					AttachMetadata:        &v1.AttachMetadata{Node: true},
				},
			},
			ep: v1.PodMetricsEndpoint{
				Port:            "metrics",
				EnableHttp2:     &falseVal,
				Path:            "/foo",
				Scheme:          "https",
				ScrapeTimeout:   "17m",
				Interval:        "1s",
				HonorLabels:     true,
				HonorTimestamps: &falseVal,
				FilterRunning:   &falseVal,
				BearerTokenSecret: corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{Name: "verysecret"},
					Key:                  "bts",
				},
				BasicAuth: &v1.BasicAuth{
					Username: corev1.SecretKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{Name: "verysecret"},
						Key:                  "username",
					},
					Password: corev1.SecretKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{Name: "verysecret"},
						Key:                  "pass",
					},
				},
				OAuth2: &v1.OAuth2{
					ClientID: v1.SecretOrConfigMap{
						ConfigMap: &corev1.ConfigMapKeySelector{
							LocalObjectReference: corev1.LocalObjectReference{Name: "notsosecret"},
							Key:                  "oauthclientid",
						},
					},
					ClientSecret: corev1.SecretKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{Name: "verysecret"},
						Key:                  "oathsecret",
					},
					TokenURL:       "https://token.example.com",
					Scopes:         []string{"some", "scope"},
					EndpointParams: map[string]string{"w": "ut"},
				},
				TLSConfig: &v1.PodMetricsEndpointTLSConfig{
					SafeTLSConfig: v1.SafeTLSConfig{
						CA: v1.SecretOrConfigMap{
							ConfigMap: &corev1.ConfigMapKeySelector{
								LocalObjectReference: corev1.LocalObjectReference{Name: "notsosecret"},
								Key:                  "ca",
							},
						},
						Cert: v1.SecretOrConfigMap{
							ConfigMap: &corev1.ConfigMapKeySelector{
								LocalObjectReference: corev1.LocalObjectReference{Name: "notsosecret"},
								Key:                  "cert",
							},
						},
						KeySecret: &corev1.SecretKeySelector{
							LocalObjectReference: corev1.LocalObjectReference{Name: "verysecret"},
							Key:                  "tls_key",
						},
						ServerName:         "foo.com",
						InsecureSkipVerify: true,
					},
				},
				Authorization: &v1.SafeAuthorization{
					Credentials: &corev1.SecretKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{Name: "verysecret"},
						Key:                  "creds",
					},
				},
			},
			expectedRelabels: util.Untab(`
				- source_labels: [job]
				  target_label: __tmp_prometheus_job_name
				- action: keep
				  regex: (bar);true
				  source_labels: [__meta_kubernetes_pod_label_foo,__meta_kubernetes_pod_labelpresent_foo]		
				- source_labels: [__meta_kubernetes_pod_container_port_name]
				  regex: metrics
				  action: keep
				- source_labels: [__meta_kubernetes_namespace]
				  target_label: namespace
				- source_labels: [__meta_kubernetes_pod_container_name]
				  target_label: container
				- source_labels: [__meta_kubernetes_pod_name]
				  target_label: pod
				- source_labels: [__meta_kubernetes_pod_label_label_a]
				  target_label: label_a
				  replacement: "${1}"
				  regex: "(.+)"
				- source_labels: [__meta_kubernetes_pod_label_label_b]
				  target_label: label_b
				  replacement: "${1}"
				  regex: "(.+)"
				- target_label: job
				  replacement: operator/podmonitor
				- source_labels: [__meta_kubernetes_pod_label_abc]
				  replacement: "${1}"
				  regex: "(.+)"
				  target_label: job
				
				- target_label: endpoint
				  replacement: metrics
			`),
			expected: &config.ScrapeConfig{
				JobName:         "podMonitor/operator/podmonitor/1",
				HonorTimestamps: false,
				HonorLabels:     true,
				ScrapeInterval:  model.Duration(time.Second),
				ScrapeTimeout:   model.Duration(17 * time.Minute),
				MetricsPath:     "/foo",
				Scheme:          "https",
				HTTPClientConfig: commonConfig.HTTPClientConfig{
					FollowRedirects: true,
					EnableHTTP2:     false,
					BearerTokenFile: "secret/operator/verysecret/bts",
					BasicAuth: &commonConfig.BasicAuth{
						Username:     "content of secret:operator/verysecret:username",
						PasswordFile: "secret/operator/verysecret/pass",
					},
					OAuth2: &commonConfig.OAuth2{
						ClientID:         "content of configmap:operator/notsosecret:oauthclientid",
						ClientSecretFile: "secret/operator/verysecret/oathsecret",
						TokenURL:         "https://token.example.com",
						Scopes:           []string{"some", "scope"},
						EndpointParams:   map[string]string{"w": "ut"},
					},
					TLSConfig: commonConfig.TLSConfig{
						CAFile:             "configmap/operator/notsosecret/ca",
						CertFile:           "configmap/operator/notsosecret/cert",
						KeyFile:            "secret/operator/verysecret/tls_key",
						ServerName:         "foo.com",
						InsecureSkipVerify: true,
					},
					Authorization: &commonConfig.Authorization{
						Type:            "Bearer",
						CredentialsFile: "secret/operator/verysecret/creds",
					},
				},
				ServiceDiscoveryConfigs: discovery.Configs{
					&promk8s.SDConfig{
						Role:           "pod",
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
	for i, tc := range suite {
		t.Run(tc.name, func(t *testing.T) {
			cg := &configGenerator{
				config:   &tc.args,
				secretfs: fakeSecretStore{},
			}
			cfg, err := cg.generatePodMonitorConfig(tc.m, tc.ep, i)
			// check relabel configs seperately
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
				}
			}
			checkRelabels(rlcs, tc.expectedRelabels)
			checkRelabels(mrlcs, tc.expectedMetricRelabels)
		})
	}
}
