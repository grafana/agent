package config_gen

import (
	"net/url"
	"testing"

	"github.com/grafana/agent/component/common/config"
	k8sConfig "github.com/grafana/agent/component/common/config"
	v1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	promConfig "github.com/prometheus/common/config"
	promk8s "github.com/prometheus/prometheus/discovery/kubernetes"
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
		client         *k8sConfig.ClientArguments
		attachMetadata *v1.AttachMetadata
		expected       *promk8s.SDConfig
	}{
		{
			name: "empty",
			client: &k8sConfig.ClientArguments{
				APIServer: k8sConfig.URL{},
			},
			attachMetadata: nil,
			expected: &promk8s.SDConfig{
				Role:               promk8s.RoleEndpoint,
				NamespaceDiscovery: nsDiscovery,
			},
		},
		{
			name: "kubeconfig",
			client: &k8sConfig.ClientArguments{
				KubeConfig: "kubeconfig",
				APIServer:  k8sConfig.URL{},
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
			client: &k8sConfig.ClientArguments{
				APIServer: k8sConfig.URL{},
			},
			attachMetadata: &v1.AttachMetadata{
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
			client: &k8sConfig.ClientArguments{
				APIServer: k8sConfig.URL{
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
			got := cg.generateK8SSDConfig(v1.NamespaceSelector{}, "", promk8s.RoleEndpoint, tt.attachMetadata)
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestGenerateSafeTLSConfig(t *testing.T) {
	tests := []struct {
		name       string
		tlsConfig  v1.SafeTLSConfig
		hasErr     bool
		serverName string
	}{
		{
			name: "empty",
			tlsConfig: v1.SafeTLSConfig{
				InsecureSkipVerify: true,
				ServerName:         "test",
			},
			hasErr:     false,
			serverName: "test",
		},
		{
			name: "ca_file",
			tlsConfig: v1.SafeTLSConfig{
				InsecureSkipVerify: true,
				CA:                 v1.SecretOrConfigMap{Secret: &k8sv1.SecretKeySelector{Key: "ca_file"}},
			},
			hasErr:     true,
			serverName: "",
		},
		{
			name: "ca_file",
			tlsConfig: v1.SafeTLSConfig{
				InsecureSkipVerify: true,
				CA:                 v1.SecretOrConfigMap{ConfigMap: &k8sv1.ConfigMapKeySelector{Key: "ca_file"}},
			},
			hasErr:     true,
			serverName: "",
		},
		{
			name: "cert_file",
			tlsConfig: v1.SafeTLSConfig{
				InsecureSkipVerify: true,
				Cert:               v1.SecretOrConfigMap{Secret: &k8sv1.SecretKeySelector{Key: "cert_file"}},
			},
			hasErr:     true,
			serverName: "",
		},
		{
			name: "cert_file",
			tlsConfig: v1.SafeTLSConfig{
				InsecureSkipVerify: true,
				Cert:               v1.SecretOrConfigMap{ConfigMap: &k8sv1.ConfigMapKeySelector{Key: "cert_file"}},
			},
			hasErr:     true,
			serverName: "",
		},
		{
			name: "key_file",
			tlsConfig: v1.SafeTLSConfig{
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
