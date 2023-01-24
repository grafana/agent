package config

import (
	"errors"
	"path/filepath"
	"testing"
	"time"

	"github.com/grafana/agent/pkg/server"
	"github.com/grafana/agent/pkg/util"
	"github.com/prometheus/common/config"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
)

// testRemoteConfigProvider is an implementation of remoteConfigProvider that can be
// used for testing. It allows setting the values to return for both fetching the
// remote config bytes & errors as well as the cached config & errors.
type testRemoteConfigProvider struct {
	AgentManagement *AgentManagement

	fetchedConfigBytesToReturn []byte
	fetchedConfigErrorToReturn error

	cachedConfigToReturn      *Config
	cachedConfigErrorToReturn error
	didCacheRemoteConfig      bool
}

func (t *testRemoteConfigProvider) AgentManagementConfig() *AgentManagement {
	return t.AgentManagement
}

func (t *testRemoteConfigProvider) GetCachedRemoteConfig(expandEnvVars bool) (*Config, error) {
	return t.cachedConfigToReturn, t.cachedConfigErrorToReturn
}

func (t *testRemoteConfigProvider) FetchRemoteConfig() ([]byte, error) {
	return t.fetchedConfigBytesToReturn, t.fetchedConfigErrorToReturn
}

func (t *testRemoteConfigProvider) CacheRemoteConfig(r []byte) error {
	t.didCacheRemoteConfig = true
	return nil
}

var validAgentManagement = AgentManagement{
	Enabled: true,
	Url:     "https://localhost:1234/example/api",
	BasicAuth: config.BasicAuth{
		Username:     "test",
		PasswordFile: "/test/path",
	},
	Protocol:        "https",
	PollingInterval: "1m",
	CacheLocation:   "/test/path/",
	RemoteConfiguration: RemoteConfiguration{
		Labels:    labelMap{"b": "B", "a": "A"},
		Namespace: "test_namespace",
	},
}

func TestValidateValidConfig(t *testing.T) {
	assert.NoError(t, validAgentManagement.Validate())
}

func TestValidateInvalidBasicAuth(t *testing.T) {
	invalidConfig := &AgentManagement{
		Enabled:         true,
		Url:             "https://localhost:1234",
		BasicAuth:       config.BasicAuth{},
		Protocol:        "https",
		PollingInterval: "1m",
		CacheLocation:   "/test/path/",
		RemoteConfiguration: RemoteConfiguration{
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
		RemoteConfiguration: RemoteConfiguration{
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
		RemoteConfiguration: RemoteConfiguration{
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
	c := validAgentManagement
	st, err := c.SleepTime()
	assert.NoError(t, err)
	assert.Equal(t, time.Minute*1, st)

	c.PollingInterval = "15s"
	st, err = c.SleepTime()
	assert.NoError(t, err)
	assert.Equal(t, time.Second*15, st)
}

func TestFullUrl(t *testing.T) {
	c := validAgentManagement
	actual, err := c.fullUrl()
	assert.NoError(t, err)
	assert.Equal(t, "https://localhost:1234/example/api/namespace/test_namespace/remote_config?a=A&b=B", actual)
}

func TestGetRemoteConfig_InvalidInitialConfig(t *testing.T) {
	// this is invalid because it is missing the password file
	invalidAgentManagement := &AgentManagement{
		Enabled: true,
		Url:     "https://localhost:1234/example/api",
		BasicAuth: config.BasicAuth{
			Username: "test",
		},
		Protocol:        "https",
		PollingInterval: "1m",
		CacheLocation:   "/test/path/",
		RemoteConfiguration: RemoteConfiguration{
			Labels:    labelMap{"b": "B", "a": "A"},
			Namespace: "test_namespace",
		},
	}

	logger := server.NewLogger(&server.DefaultConfig)
	testProvider := testRemoteConfigProvider{AgentManagement: invalidAgentManagement}

	_, err := getRemoteConfig(true, &testProvider, logger)
	assert.Error(t, err)
	assert.False(t, testProvider.didCacheRemoteConfig)
}

func TestGetRemoteConfig_UnmarshallableRemoteConfig(t *testing.T) {
	brokenCfg := `completely invalid config (maybe it got corrupted, maybe it was somehow set this way)`

	invalidCfgBytes, err := yaml.Marshal(brokenCfg)
	assert.NoError(t, err)

	am := validAgentManagement
	logger := server.NewLogger(&server.DefaultConfig)
	testProvider := testRemoteConfigProvider{AgentManagement: &am}
	testProvider.fetchedConfigBytesToReturn = invalidCfgBytes
	testProvider.cachedConfigToReturn = &DefaultConfig

	cfg, err := getRemoteConfig(true, &testProvider, logger)
	assert.NoError(t, err)
	assert.False(t, testProvider.didCacheRemoteConfig)

	// check that the returned config is the cached one
	assert.True(t, util.CompareYAML(*cfg, DefaultConfig))
}

func TestGetRemoteConfig_RemoteFetchFails(t *testing.T) {
	am := validAgentManagement
	logger := server.NewLogger(&server.DefaultConfig)
	testProvider := testRemoteConfigProvider{AgentManagement: &am}
	testProvider.fetchedConfigErrorToReturn = errors.New("connection refused")
	testProvider.cachedConfigToReturn = &DefaultConfig

	cfg, err := getRemoteConfig(true, &testProvider, logger)
	assert.NoError(t, err)
	assert.False(t, testProvider.didCacheRemoteConfig)

	// check that the returned config is the cached one
	assert.True(t, util.CompareYAML(*cfg, DefaultConfig))
}

func TestGetRemoteConfig_ValidRemoteConfig(t *testing.T) {
	validConfig := `server:
  log_level: info`

	validConfigBytes := []byte(validConfig)

	am := validAgentManagement
	logger := server.NewLogger(&server.DefaultConfig)
	testProvider := testRemoteConfigProvider{AgentManagement: &am}
	testProvider.fetchedConfigBytesToReturn = validConfigBytes

	_, err := getRemoteConfig(true, &testProvider, logger)
	assert.NoError(t, err)
	assert.True(t, testProvider.didCacheRemoteConfig)
}
