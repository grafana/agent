package config

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"github.com/go-kit/log/level"
	"github.com/grafana/agent/pkg/config/instrumentation"
	"github.com/grafana/agent/pkg/server"
	"github.com/prometheus/common/config"
)

const cacheFilename = "remote-config-cache.yaml"

type remoteConfigProvider interface {
	GetCachedRemoteConfig(expandEnvVars bool) (*Config, error)
	CacheRemoteConfig(remoteConfigBytes []byte) error
	FetchRemoteConfig() ([]byte, error)
}

type remoteConfigHTTPProvider struct {
	InitialConfig *AgentManagementConfig
}

func newRemoteConfigHTTPProvider(c *Config) (*remoteConfigHTTPProvider, error) {
	err := c.AgentManagement.Validate()
	if err != nil {
		return nil, err
	}
	return &remoteConfigHTTPProvider{
		InitialConfig: &c.AgentManagement,
	}, nil
}

type remoteConfigCache struct {
	UrlHash string `json:"url_hash"`
	Config  string `json:"config"`
}

func hashUrl(u string) string {
	hashed := sha256.Sum256([]byte(u))
	return hex.EncodeToString(hashed[:])
}

// GetCachedRemoteConfig retrieves the cached remote config from the location specified
// in r.AgentManagement.CacheLocation
func (r remoteConfigHTTPProvider) GetCachedRemoteConfig(expandEnvVars bool) (*Config, error) {
	cachePath := filepath.Join(r.InitialConfig.CacheLocation, cacheFilename)
	curUrl, err := r.InitialConfig.fullUrl()
	if err != nil {
		return nil, fmt.Errorf("unable to create full url: %w", err)
	}
	var configCache remoteConfigCache

	buf, err := os.ReadFile(cachePath)

	if err != nil {
		return nil, fmt.Errorf("error reading remote config cache: %w", err)
	}

	if err := json.Unmarshal(buf, &configCache); err != nil {
		return nil, fmt.Errorf("error trying to load cached remote config from file: %w", err)
	}

	// If a different url was used when caching the config, it is no longer valid
	if r.InitialConfig.InvalidateCacheOnUrlChange && !(configCache.UrlHash == hashUrl(curUrl)) {
		return nil, errors.New("invalid remote config cache: url hashes don't match")
	}

	var cachedConfig Config

	if err = LoadBytes([]byte(configCache.Config), expandEnvVars, &cachedConfig); err != nil {
		return nil, fmt.Errorf("unable to load cached config: %w", err)
	}

	return &cachedConfig, nil
}

// CacheRemoteConfig caches the remote config to the location specified in
// r.AgentManagement.CacheLocation
func (r remoteConfigHTTPProvider) CacheRemoteConfig(remoteConfigBytes []byte) error {
	cachePath := filepath.Join(r.InitialConfig.CacheLocation, cacheFilename)
	u, err := r.InitialConfig.fullUrl()
	if err != nil {
		return fmt.Errorf("unable to create full url: %w", err)
	}
	configCache := remoteConfigCache{
		UrlHash: hashUrl(u),
		Config:  string(remoteConfigBytes),
	}
	marshalled, err := json.Marshal(configCache)
	if err != nil {
		return fmt.Errorf("could not marshal remote config cache: %w", err)
	}
	return os.WriteFile(cachePath, marshalled, 0666)
}

// FetchRemoteConfig fetches the raw bytes of the config from a remote API using
// the values in r.AgentManagement.
func (r remoteConfigHTTPProvider) FetchRemoteConfig() ([]byte, error) {
	httpClientConfig := &config.HTTPClientConfig{
		BasicAuth: &r.InitialConfig.BasicAuth,
	}

	dir, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get current working directory: %w", err)
	}
	httpClientConfig.SetDirectory(dir)

	remoteOpts := &remoteOpts{
		HTTPClientConfig: httpClientConfig,
	}

	url, err := r.InitialConfig.fullUrl()
	if err != nil {
		return nil, fmt.Errorf("error trying to create full url: %w", err)
	}
	rc, err := newRemoteConfig(url, remoteOpts)
	if err != nil {
		return nil, fmt.Errorf("error reading remote config: %w", err)
	}

	bb, err := rc.retrieve()
	if err != nil {
		return nil, fmt.Errorf("error retrieving remote config: %w", err)
	}
	return bb, nil
}

type labelMap map[string]string

type RemoteConfiguration struct {
	Labels    labelMap `yaml:"labels"`
	Namespace string   `yaml:"namespace"`
}

type AgentManagementConfig struct {
	Enabled                    bool             `yaml:"-"` // Derived from enable-features=agent-management
	Url                        string           `yaml:"api_url"`
	BasicAuth                  config.BasicAuth `yaml:"basic_auth"`
	Protocol                   string           `yaml:"protocol"`
	PollingInterval            string           `yaml:"polling_interval"`
	CacheLocation              string           `yaml:"remote_config_cache_location"`
	InvalidateCacheOnUrlChange bool             `yaml:"invalidate_cache_on_url_change"` // Whether to invalidate the cached config if the url hash is different

	RemoteConfiguration RemoteConfiguration `yaml:"remote_configuration"`
}

// getRemoteConfig gets the remote config specified in the initial config, falling back to a local, cached copy
// of the remote config if the request to the remote fails. If both fail, an empty config and an
// error will be returned.
//
// It also validates that the response from the API is a semantically correct config by calling config.Validate().
func getRemoteConfig(expandEnvVars bool, configProvider remoteConfigProvider, log *server.Logger, fs *flag.FlagSet, args []string, configPath string) (*Config, error) {
	remoteConfigBytes, err := configProvider.FetchRemoteConfig()
	if err != nil {
		level.Error(log).Log("msg", "could not fetch from API, falling back to cache", "err", err)
		return configProvider.GetCachedRemoteConfig(expandEnvVars)
	}
	var remoteConfig Config

	err = LoadBytes(remoteConfigBytes, expandEnvVars, &remoteConfig)
	if err != nil {
		level.Error(log).Log("msg", "could not load the response from the API, falling back to cache", "err", err)
		instrumentation.InstrumentInvalidRemoteConfig("invalid_yaml")
		return configProvider.GetCachedRemoteConfig(expandEnvVars)
	}

	// this is done in order to validate the config semantically
	if err = applyIntegrationValuesFromFlagset(fs, args, configPath, &remoteConfig); err != nil {
		level.Error(log).Log("msg", "could not load integrations from config, falling back to cache", "err", err)
		instrumentation.InstrumentInvalidRemoteConfig("invalid_integrations_config")
		return configProvider.GetCachedRemoteConfig(expandEnvVars)
	}
	if err = remoteConfig.Validate(fs); err != nil {
		level.Error(log).Log("msg", "invalid config received from the API, falling back to cache", "err", err)
		instrumentation.InstrumentInvalidRemoteConfig("invalid_agent_config")
		return configProvider.GetCachedRemoteConfig(expandEnvVars)
	}

	level.Info(log).Log("msg", "fetched and loaded remote config from API")

	if err = configProvider.CacheRemoteConfig(remoteConfigBytes); err != nil {
		level.Error(log).Log("err", fmt.Errorf("could not cache config locally: %w", err))
	}
	return &remoteConfig, nil
}

// newRemoteConfigProvider creates a remoteConfigProvider based on the protocol
// specified in c.AgentManagement
func newRemoteConfigProvider(c *Config) (*remoteConfigHTTPProvider, error) {
	switch p := c.AgentManagement.Protocol; {
	case p == "http":
		return newRemoteConfigHTTPProvider(c)
	default:
		return nil, fmt.Errorf("unsupported protocol for agent management api: %s", p)
	}
}

// fullUrl creates and returns the URL that should be used when querying the Agent Management API,
// including the namespace, base config id, and any labels that have been specified.
func (am *AgentManagementConfig) fullUrl() (string, error) {
	fullPath, err := url.JoinPath(am.Url, "namespace", am.RemoteConfiguration.Namespace, "remote_config")
	if err != nil {
		return "", fmt.Errorf("error trying to join url: %w", err)
	}
	u, err := url.Parse(fullPath)
	if err != nil {
		return "", fmt.Errorf("error trying to parse url: %w", err)
	}
	q := u.Query()
	for label, value := range am.RemoteConfiguration.Labels {
		q.Add(label, value)
	}
	u.RawQuery = q.Encode()
	return u.String(), nil
}

// SleepTime returns the parsed duration in between config fetches.
func (am *AgentManagementConfig) SleepTime() (time.Duration, error) {
	return time.ParseDuration(am.PollingInterval)
}

// Validate checks that necessary portions of the config have been set.
func (am *AgentManagementConfig) Validate() error {
	if am.BasicAuth.Username == "" || am.BasicAuth.PasswordFile == "" {
		return errors.New("both username and password_file fields must be specified")
	}

	if _, err := time.ParseDuration(am.PollingInterval); err != nil {
		return fmt.Errorf("error trying to parse polling interval: %w", err)
	}

	if am.RemoteConfiguration.Namespace == "" {
		return errors.New("namespace must be specified in 'remote_configuration' block of the config")
	}

	if am.CacheLocation == "" {
		return errors.New("path to cache must be specified in 'agent_management.remote_config_cache_location'")
	}

	return nil
}
