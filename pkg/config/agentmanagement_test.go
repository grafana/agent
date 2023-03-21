package config

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"flag"
	"testing"
	"time"

	"github.com/grafana/agent/pkg/config/features"
	"github.com/grafana/agent/pkg/server"
	"github.com/grafana/agent/pkg/util"
	"github.com/prometheus/common/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
)

// testRemoteConfigProvider is an implementation of remoteConfigProvider that can be
// used for testing. It allows setting the values to return for both fetching the
// remote config bytes & errors as well as the cached config & errors.
type testRemoteConfigProvider struct {
	InitialConfig *AgentManagementConfig

	fetchedConfigBytesToReturn []byte
	fetchedConfigErrorToReturn error

	cachedConfigToReturn      *Config
	cachedConfigErrorToReturn error
	didCacheRemoteConfig      bool
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

var validAgentManagementConfig = AgentManagementConfig{
	Enabled: true,
	Url:     "https://localhost:1234/example/api",
	BasicAuth: config.BasicAuth{
		Username:     "test",
		PasswordFile: "/test/path",
	},
	Protocol:        "https",
	PollingInterval: time.Minute,
	CacheLocation:   "/test/path/",
	RemoteConfiguration: RemoteConfiguration{
		Labels:    labelMap{"b": "B", "a": "A"},
		Namespace: "test_namespace",
	},
}

func TestValidateValidConfig(t *testing.T) {
	assert.NoError(t, validAgentManagementConfig.Validate())
}

func TestValidateInvalidBasicAuth(t *testing.T) {
	invalidConfig := &AgentManagementConfig{
		Enabled:         true,
		Url:             "https://localhost:1234",
		BasicAuth:       config.BasicAuth{},
		Protocol:        "https",
		PollingInterval: time.Minute,
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

func TestMissingCacheLocation(t *testing.T) {
	invalidConfig := &AgentManagementConfig{
		Enabled: true,
		Url:     "https://localhost:1234",
		BasicAuth: config.BasicAuth{
			Username:     "test",
			PasswordFile: "/test/path",
		},
		Protocol:        "https",
		PollingInterval: 1 * time.Minute,
		RemoteConfiguration: RemoteConfiguration{
			Namespace: "test_namespace",
		},
	}
	assert.Error(t, invalidConfig.Validate())
}

func TestSleepTime(t *testing.T) {
	cfg := `
api_url: "http://localhost"
basic_auth:
  username: "initial_user"
protocol: "http"
polling_interval: "1m"
remote_config_cache_location: "/etc"
remote_configuration:
  namespace: "new_namespace"`

	var am AgentManagementConfig
	yaml.Unmarshal([]byte(cfg), &am)
	assert.Equal(t, time.Minute, am.SleepTime())
}

func TestFuzzJitterTime(t *testing.T) {
	am := validAgentManagementConfig
	pollingInterval := 2 * time.Minute
	am.PollingInterval = pollingInterval

	zero := time.Duration(0)

	for i := 0; i < 10_000; i++ {
		j := am.JitterTime()
		assert.GreaterOrEqual(t, j, zero)
		assert.Less(t, j, pollingInterval)
	}
}

func TestFullUrl(t *testing.T) {
	c := validAgentManagementConfig
	actual, err := c.fullUrl()
	assert.NoError(t, err)
	assert.Equal(t, "https://localhost:1234/example/api/namespace/test_namespace/remote_config?a=A&b=B", actual)
}

func TestRemoteConfigHashCheck(t *testing.T) {
	// not a truly valid Agent Management config, but used for testing against
	// precomputed sha256 hash
	ic := AgentManagementConfig{
		Protocol: "http",
	}
	marshalled, err := yaml.Marshal(ic)
	require.NoError(t, err)
	icHashBytes := sha256.Sum256(marshalled)
	icHash := hex.EncodeToString(icHashBytes[:])

	rcCache := remoteConfigCache{
		InitialConfigHash: icHash,
		Config:            "server:\\n log_level: debug",
	}

	require.NoError(t, initialConfigHashCheck(ic, rcCache))
	rcCache.InitialConfigHash = "abc"
	require.Error(t, initialConfigHashCheck(ic, rcCache))

	differentIc := validAgentManagementConfig
	require.Error(t, initialConfigHashCheck(differentIc, rcCache))
}

func TestNewRemoteConfigHTTPProvider_InvalidInitialConfig(t *testing.T) {
	// this is invalid because it is missing the password file
	invalidAgentManagementConfig := &AgentManagementConfig{
		Enabled: true,
		Url:     "https://localhost:1234/example/api",
		BasicAuth: config.BasicAuth{
			Username: "test",
		},
		Protocol:        "https",
		PollingInterval: time.Minute,
		CacheLocation:   "/test/path/",
		RemoteConfiguration: RemoteConfiguration{
			Labels:    labelMap{"b": "B", "a": "A"},
			Namespace: "test_namespace",
		},
	}

	cfg := Config{
		AgentManagement: *invalidAgentManagementConfig,
	}
	_, err := newRemoteConfigHTTPProvider(&cfg)
	assert.Error(t, err)
}

func TestGetRemoteConfig_UnmarshallableRemoteConfig(t *testing.T) {
	defaultCfg := DefaultConfig()
	brokenCfg := `completely invalid config (maybe it got corrupted, maybe it was somehow set this way)`

	invalidCfgBytes := []byte(brokenCfg)

	am := validAgentManagementConfig
	logger := server.NewLogger(defaultCfg.Server)
	testProvider := testRemoteConfigProvider{InitialConfig: &am}
	testProvider.fetchedConfigBytesToReturn = invalidCfgBytes
	testProvider.cachedConfigToReturn = &defaultCfg

	// a nil flagset is being used for testing because it should not reach flag validation
	cfg, err := getRemoteConfig(true, &testProvider, logger, nil, []string{}, "test")
	assert.NoError(t, err)
	assert.False(t, testProvider.didCacheRemoteConfig)

	// check that the returned config is the cached one
	assert.True(t, util.CompareYAML(*cfg, defaultCfg))
}

func TestGetRemoteConfig_RemoteFetchFails(t *testing.T) {
	defaultCfg := DefaultConfig()

	am := validAgentManagementConfig
	logger := server.NewLogger(defaultCfg.Server)
	testProvider := testRemoteConfigProvider{InitialConfig: &am}
	testProvider.fetchedConfigErrorToReturn = errors.New("connection refused")
	testProvider.cachedConfigToReturn = &defaultCfg

	// a nil flagset is being used for testing because it should not reach flag validation
	cfg, err := getRemoteConfig(true, &testProvider, logger, nil, []string{}, "test")
	assert.NoError(t, err)
	assert.False(t, testProvider.didCacheRemoteConfig)

	// check that the returned config is the cached one
	assert.True(t, util.CompareYAML(*cfg, defaultCfg))
}

func TestGetRemoteConfig_InvalidRemoteConfig(t *testing.T) {
	defaultCfg := DefaultConfig()

	// this is invalid because it has two scrape_configs with
	// the same job_name
	invalidConfig := `
metrics:
    configs:
    - name: Metrics Snippets
      scrape_configs:
      - job_name: agent-metrics
        honor_timestamps: true
        scrape_interval: 15s
        metrics_path: /metrics
        scheme: http
        follow_redirects: true
        enable_http2: true
        static_configs:
        - targets:
          - localhost:12345
      - job_name: agent-metrics
        honor_timestamps: true
        scrape_interval: 15s
        metrics_path: /metrics
        scheme: http
        follow_redirects: true
        enable_http2: true
        static_configs:
        - targets:
          - localhost:12345`
	invalidCfgBytes := []byte(invalidConfig)

	am := validAgentManagementConfig
	logger := server.NewLogger(defaultCfg.Server)
	testProvider := testRemoteConfigProvider{InitialConfig: &am}
	testProvider.fetchedConfigBytesToReturn = invalidCfgBytes
	testProvider.cachedConfigToReturn = &defaultCfg
	fs := flag.NewFlagSet("test", flag.ExitOnError)
	features.Register(fs, allFeatures)

	cfg, err := getRemoteConfig(true, &testProvider, logger, fs, []string{}, "test")
	assert.NoError(t, err)
	assert.False(t, testProvider.didCacheRemoteConfig)

	// check that the returned config is the cached one
	assert.True(t, util.CompareYAML(*cfg, defaultCfg))
}
