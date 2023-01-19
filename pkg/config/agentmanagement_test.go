package config

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/prometheus/common/config"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
)

func TestValidateValidConfig(t *testing.T) {
	validConfigPolling := &AgentManagement{
		Enabled: true,
		Url:     "https://localhost:1234",
		BasicAuth: config.BasicAuth{
			Username:     "test",
			PasswordFile: "/test/path",
		},
		Protocol:        "https",
		PollingInterval: "1m",
		CacheLocation:   "/test/path/",
		RemoteConfiguration: remoteConfiguration{
			Namespace: "test_namespace",
		},
	}
	assert.NoError(t, validConfigPolling.Validate())
}

func TestValidateInvalidBasicAuth(t *testing.T) {
	invalidConfig := &AgentManagement{
		Enabled:         true,
		Url:             "https://localhost:1234",
		BasicAuth:       config.BasicAuth{},
		Protocol:        "https",
		PollingInterval: "1m",
		CacheLocation:   "/test/path/",
		RemoteConfiguration: remoteConfiguration{
			Namespace: "test_namespace",
		},
	}
	assert.Error(t, invalidConfig.Validate())

	invalidConfig.BasicAuth.Username = "test"
	assert.Error(t, invalidConfig.Validate()) // Should still error as there is no password file set

	invalidConfig.BasicAuth.Username = ""
	invalidConfig.BasicAuth.PasswordFile = "/test/path"
	assert.Error(t, invalidConfig.Validate()) // Should still error as there is no username set
}

func TestValidateInvalidPollingInterval(t *testing.T) {
	invalidConfig := &AgentManagement{
		Enabled: true,
		Url:     "https://localhost:1234",
		BasicAuth: config.BasicAuth{
			Username:     "test",
			PasswordFile: "/test/path",
		},
		Protocol:        "https",
		PollingInterval: "1?",
		CacheLocation:   "/test/path/",
		RemoteConfiguration: remoteConfiguration{
			Namespace: "test_namespace",
		},
	}
	assert.Error(t, invalidConfig.Validate())

	invalidConfig.PollingInterval = ""
	assert.Error(t, invalidConfig.Validate())
}

func TestMissingCacheLocation(t *testing.T) {
	invalidConfig := &AgentManagement{
		Enabled: true,
		Url:     "https://localhost:1234",
		BasicAuth: config.BasicAuth{
			Username:     "test",
			PasswordFile: "/test/path",
		},
		Protocol:        "https",
		PollingInterval: "1?",
		RemoteConfiguration: remoteConfiguration{
			Namespace: "test_namespace",
		},
	}
	assert.Error(t, invalidConfig.Validate())
}

func TestGetCachedRemoteConfig(t *testing.T) {
	cwd := filepath.Clean("./testdata/")
	_, err := getCachedRemoteConfig(cwd, false)
	assert.NoError(t, err)
}

func TestSleepTime(t *testing.T) {
	c := &AgentManagement{
		Enabled: true,
		Url:     "https://localhost:1234",
		BasicAuth: config.BasicAuth{
			Username:     "test",
			PasswordFile: "/test/path",
		},
		Protocol:        "https",
		PollingInterval: "1m",
		CacheLocation:   "/test/path/",
		RemoteConfiguration: remoteConfiguration{
			Namespace: "test_namespace",
		},
	}
	st, err := c.SleepTime()
	assert.NoError(t, err)
	assert.Equal(t, time.Minute*1, st)

	c.PollingInterval = "15s"
	st, err = c.SleepTime()
	assert.NoError(t, err)
	assert.Equal(t, time.Second*15, st)
}

func TestFullUrl(t *testing.T) {
	c := &AgentManagement{
		Enabled: true,
		Url:     "https://localhost:1234/example/api",
		BasicAuth: config.BasicAuth{
			Username:     "test",
			PasswordFile: "/test/path",
		},
		Protocol:        "https",
		PollingInterval: "1m",
		CacheLocation:   "/test/path/",
		RemoteConfiguration: remoteConfiguration{
			Labels:    labelMap{"b": "B", "a": "A"},
			Namespace: "test_namespace",
		},
	}
	actual, err := c.fullUrl()
	assert.NoError(t, err)
	assert.Equal(t, "https://localhost:1234/example/api/namespace/test_namespace/remote_config?a=A&b=B", actual)
}

func TestBareConfig(t *testing.T) {
	cfg := `
protocol: http
`
	var am AgentManagement
	err := yaml.UnmarshalStrict([]byte(cfg), &am)
	assert.NoError(t, err)
	assert.Equal(t, defaultConfig.CacheLocation, am.CacheLocation)
}

func TestUnmarshal(t *testing.T) {
	cfg := `
api_url: http://localhost:8080
basic_auth:
  username: test_user
  password_file: /tmp/test/passfile
protocol: http
polling_interval: 1m
remote_configuration:
  namespace: test_namespace
  labels:
    l1: label1`

	var am AgentManagement
	err := yaml.UnmarshalStrict([]byte(cfg), &am)
	assert.NoError(t, err)
	expected := AgentManagement{
		Url: "http://localhost:8080",
		BasicAuth: config.BasicAuth{
			Username:     "test_user",
			PasswordFile: "/tmp/test/passfile",
		},
		Protocol:        "http",
		PollingInterval: "1m",
		CacheLocation:   defaultConfig.CacheLocation,
		RemoteConfiguration: remoteConfiguration{
			Labels:    labelMap{"l1": "label1"},
			Namespace: "test_namespace",
		},
	}
	assert.Equal(t, expected, am)
}
