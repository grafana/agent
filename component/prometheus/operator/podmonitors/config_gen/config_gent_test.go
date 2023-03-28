package config_gen

import (
	"testing"

	v1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/stretchr/testify/assert"
	k8sv1 "k8s.io/api/core/v1"
)

var (
	configGen = &ConfigGenerator{}
)

func TestGenerateSafeTLSConfig(t *testing.T) {
	serverName := "test"
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
				ServerName:         serverName,
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
