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

func TestDefaultConfig(t *testing.T) {
	empty := `agent_management:`
	var am AgentManagement
	yaml.Unmarshal([]byte(empty), &am)
	assert.Equal(t, "data-agent/", am.CacheLocation)
}
