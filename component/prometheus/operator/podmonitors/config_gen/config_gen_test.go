package config_gen

import (
	"net/url"
	"testing"

	"github.com/grafana/agent/component/common/config"
	promopv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	promConfig "github.com/prometheus/common/config"
	"github.com/prometheus/common/model"
	promk8s "github.com/prometheus/prometheus/discovery/kubernetes"
	"github.com/prometheus/prometheus/model/relabel"
	"github.com/stretchr/testify/assert"
	k8sv1 "k8s.io/api/core/v1"
)

var (
	configGen = &ConfigGenerator{}
)

func TestGenerateK8SSDConfig(t *testing.T) {
	nsDiscovery := promk8s.NamespaceDiscovery{
		Names: []string{""},
	}

	tests := []struct {
		name           string
		client         *config.ClientArguments
		attachMetadata *promopv1.AttachMetadata
		expected       *promk8s.SDConfig
	}{
		{
			name: "empty",
			client: &config.ClientArguments{
				APIServer: config.URL{},
			},
			attachMetadata: nil,
			expected: &promk8s.SDConfig{
				Role:               promk8s.RoleEndpoint,
				NamespaceDiscovery: nsDiscovery,
			},
		},
		{
			name: "kubeconfig",
			client: &config.ClientArguments{
				KubeConfig: "kubeconfig",
				APIServer:  config.URL{},
			},
			attachMetadata: nil,
			expected: &promk8s.SDConfig{
				Role:               promk8s.RoleEndpoint,
				NamespaceDiscovery: nsDiscovery,
				KubeConfig:         "kubeconfig",
			},
		},
		{
			name: "attach metadata",
			client: &config.ClientArguments{
				APIServer: config.URL{},
			},
			attachMetadata: &promopv1.AttachMetadata{
				Node: true,
			},
			expected: &promk8s.SDConfig{
				Role:               promk8s.RoleEndpoint,
				NamespaceDiscovery: nsDiscovery,
				AttachMetadata:     promk8s.AttachMetadataConfig{Node: true},
			},
		},
		{
			name: "http client config",
			client: &config.ClientArguments{
				APIServer: config.URL{
					URL: &url.URL{
						Scheme: "https",
						Host:   "localhost:8080",
					},
				},
				HTTPClientConfig: config.HTTPClientConfig{
					BasicAuth: &config.BasicAuth{
						Username: "username",
						Password: "password",
					},
					BearerToken:     "bearer",
					BearerTokenFile: "bearer_file",
					TLSConfig: config.TLSConfig{
						CAFile:   "ca_file",
						CertFile: "cert_file",
					},
					Authorization: &config.Authorization{
						Credentials: "credentials",
					},
				},
			},
			attachMetadata: nil,
			expected: &promk8s.SDConfig{
				Role:               promk8s.RoleEndpoint,
				NamespaceDiscovery: nsDiscovery,
				APIServer: promConfig.URL{
					URL: &url.URL{
						Scheme: "https",
						Host:   "localhost:8080",
					},
				},
				HTTPClientConfig: promConfig.HTTPClientConfig{
					BasicAuth: &promConfig.BasicAuth{
						Username: "username",
						Password: "password",
					},
					BearerToken:     "bearer",
					BearerTokenFile: "bearer_file",
					TLSConfig: promConfig.TLSConfig{
						CAFile:   "ca_file",
						CertFile: "cert_file",
					},
					Authorization: &promConfig.Authorization{
						Type:        "Bearer",
						Credentials: "credentials",
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cg := &ConfigGenerator{
				Client: tt.client,
			}
			got := cg.generateK8SSDConfig(promopv1.NamespaceSelector{}, "", promk8s.RoleEndpoint, tt.attachMetadata)
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestGenerateSafeTLSConfig(t *testing.T) {
	tests := []struct {
		name       string
		tlsConfig  promopv1.SafeTLSConfig
		hasErr     bool
		serverName string
	}{
		{
			name: "empty",
			tlsConfig: promopv1.SafeTLSConfig{
				InsecureSkipVerify: true,
				ServerName:         "test",
			},
			hasErr:     false,
			serverName: "test",
		},
		{
			name: "ca_file",
			tlsConfig: promopv1.SafeTLSConfig{
				InsecureSkipVerify: true,
				CA:                 promopv1.SecretOrConfigMap{Secret: &k8sv1.SecretKeySelector{Key: "ca_file"}},
			},
			hasErr:     true,
			serverName: "",
		},
		{
			name: "ca_file",
			tlsConfig: promopv1.SafeTLSConfig{
				InsecureSkipVerify: true,
				CA:                 promopv1.SecretOrConfigMap{ConfigMap: &k8sv1.ConfigMapKeySelector{Key: "ca_file"}},
			},
			hasErr:     true,
			serverName: "",
		},
		{
			name: "cert_file",
			tlsConfig: promopv1.SafeTLSConfig{
				InsecureSkipVerify: true,
				Cert:               promopv1.SecretOrConfigMap{Secret: &k8sv1.SecretKeySelector{Key: "cert_file"}},
			},
			hasErr:     true,
			serverName: "",
		},
		{
			name: "cert_file",
			tlsConfig: promopv1.SafeTLSConfig{
				InsecureSkipVerify: true,
				Cert:               promopv1.SecretOrConfigMap{ConfigMap: &k8sv1.ConfigMapKeySelector{Key: "cert_file"}},
			},
			hasErr:     true,
			serverName: "",
		},
		{
			name: "key_file",
			tlsConfig: promopv1.SafeTLSConfig{
				InsecureSkipVerify: true,
				KeySecret:          &k8sv1.SecretKeySelector{Key: "key_file"},
			},
			hasErr:     true,
			serverName: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := configGen.generateSafeTLS(tt.tlsConfig)
			assert.Equal(t, tt.hasErr, err != nil)
			assert.True(t, got.InsecureSkipVerify)
			assert.Equal(t, tt.serverName, got.ServerName)
		})
	}
}

func TestRelabelerAdd(t *testing.T) {
	relabeler := &relabeler{}

	cfgs := []*relabel.Config{
		{
			Action:      "",
			Separator:   "",
			Regex:       relabel.Regexp{},
			Replacement: "",
		},
	}
	expected := relabel.Config{
		Action:      relabel.DefaultRelabelConfig.Action,
		Separator:   relabel.DefaultRelabelConfig.Separator,
		Regex:       relabel.DefaultRelabelConfig.Regex,
		Replacement: relabel.DefaultRelabelConfig.Replacement,
	}
	relabeler.add(cfgs...)
	cfg := cfgs[0]

	assert.Equal(t, expected.Action, cfg.Action)
	assert.Equal(t, expected.Separator, cfg.Separator)
	assert.Equal(t, expected.Regex, cfg.Regex)
	assert.Equal(t, expected.Replacement, cfg.Replacement)
}

func TestRelabelerAddFromV1(t *testing.T) {
	relabeler := &relabeler{}

	cfgs := []*promopv1.RelabelConfig{
		{
			SourceLabels: []promopv1.LabelName{"__meta_kubernetes_pod_label_app"},
			Separator:    ";",
			TargetLabel:  "app",
			Regex:        "(.*)",
			Modulus:      1,
			Replacement:  "$1",
			Action:       "replace",
		},
	}
	expected := relabel.Config{
		SourceLabels: model.LabelNames{"__meta_kubernetes_pod_label_app"},
		Separator:    ";",
		TargetLabel:  "app",
		Regex:        relabel.MustNewRegexp("(.*)"),
		Modulus:      1,
		Replacement:  "$1",
		Action:       relabel.Replace,
	}
	relabeler.addFromV1(cfgs...)
	cfg := relabeler.configs[0]

	assert.Equal(t, expected.SourceLabels, cfg.SourceLabels)
	assert.Equal(t, expected.Separator, cfg.Separator)
	assert.Equal(t, expected.TargetLabel, cfg.TargetLabel)
	assert.Equal(t, expected.Regex, cfg.Regex)
	assert.Equal(t, expected.Modulus, cfg.Modulus)
	assert.Equal(t, expected.Replacement, cfg.Replacement)
	assert.Equal(t, expected.Action, cfg.Action)
}
