package config

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestValidateValidConfig(t *testing.T) {
	validConfigPolling := &AgentManagement{
		Enabled: true,
		Url:     "https://localhost:1234",
		BasicAuth: BasicAuth{
			Username:     "test",
			PasswordFile: "/test/path",
		},
		Protocol:        "https",
		PollingInterval: "1m",
		RemoteConfiguration: RemoteConfiguration{
			BaseConfigId: "test_config",
			Namespace:    "test_namespace",
		},
	}
	assert.NoError(t, validConfigPolling.Validate())
}

func TestValidateInvalidBasicAuth(t *testing.T) {
	invalidConfig := &AgentManagement{
		Enabled:         true,
		Url:             "https://localhost:1234",
		BasicAuth:       BasicAuth{},
		Protocol:        "https",
		PollingInterval: "1m",
		RemoteConfiguration: RemoteConfiguration{
			BaseConfigId: "test_config",
			Namespace:    "test_namespace",
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
		BasicAuth: BasicAuth{
			Username:     "test",
			PasswordFile: "/test/path",
		},
		Protocol:        "https",
		PollingInterval: "1?",
		RemoteConfiguration: RemoteConfiguration{
			BaseConfigId: "test_config",
			Namespace:    "test_namespace",
		},
	}
	assert.Error(t, invalidConfig.Validate())

	invalidConfig.PollingInterval = ""
	assert.Error(t, invalidConfig.Validate())
}

func TestGetCachedRemoteConfig(t *testing.T) {
	cwd := filepath.Clean("./testdata/")
	_, err := GetCachedRemoteConfig(cwd, false)
	assert.NoError(t, err)
}

func TestSleepTime(t *testing.T) {
	c := &AgentManagement{
		Enabled: true,
		Url:     "https://localhost:1234",
		BasicAuth: BasicAuth{
			Username:     "test",
			PasswordFile: "/test/path",
		},
		Protocol:        "https",
		PollingInterval: "1m",
		RemoteConfiguration: RemoteConfiguration{
			BaseConfigId: "test_config",
			Namespace:    "test_namespace",
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

func TestGetLabelMap(t *testing.T) {
	labels := []string{"key1:value1"}

	c := &AgentManagement{
		RemoteConfiguration: RemoteConfiguration{
			Labels:       labels,
			BaseConfigId: "test_config",
			Namespace:    "test_namespace",
		},
	}
	m := c.LabelMap()
	assert.Contains(t, m, "key1")
	assert.Equal(t, "value1", m["key1"])

	c.RemoteConfiguration.Labels = []string{"key2:value2", "key3:value3"}
	m = c.LabelMap()
	assert.Contains(t, m, "key2")
	assert.Contains(t, m, "key3")
	assert.Equal(t, "value2", m["key2"])
	assert.Equal(t, "value3", m["key3"])

	c.RemoteConfiguration.Labels = []string{}
	assert.Equal(t, 0, len(c.LabelMap()))
}

func TestFullUrl(t *testing.T) {
	c := &AgentManagement{
		Enabled: true,
		Url:     "https://localhost:1234/example/api",
		BasicAuth: BasicAuth{
			Username:     "test",
			PasswordFile: "/test/path",
		},
		Protocol:        "https",
		PollingInterval: "1m",
		RemoteConfiguration: RemoteConfiguration{
			Labels:       []string{"b:B", "a:A"},
			Namespace:    "test_namespace",
			BaseConfigId: "test_config",
		},
	}
	actual, err := c.FullUrl()
	assert.NoError(t, err)
	assert.Equal(t, "https://localhost:1234/example/api/test_namespace/test_config/remote_config?a=A&b=B", actual)
}
