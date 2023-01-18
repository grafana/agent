package config

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"github.com/go-kit/log/level"
	"github.com/grafana/agent/pkg/server"
	"github.com/prometheus/common/config"
)

const cacheFilename = "remote-config-cache.yaml"

type labelMap map[string]string

type RemoteConfiguration struct {
	Labels    labelMap `yaml:"labels"`
	Namespace string   `yaml:"namespace"`
}

type AgentManagement struct {
	Enabled         bool             `yaml:"-"` // Derived from enable-features=agent-management
	Url             string           `yaml:"api_url"`
	BasicAuth       config.BasicAuth `yaml:"basic_auth"`
	Protocol        string           `yaml:"protocol"`
	PollingInterval string           `yaml:"polling_interval"`
	CacheLocation   string           `yaml:"remote_config_cache_location"`

	RemoteConfiguration RemoteConfiguration `yaml:"remote_configuration"`
}

// getRemoteConfig gets the remote config specified in the initial config, falling back to a local, cached copy
// of the remote config if the request to the remote fails. If both fail, an empty config and an
// error will be returned.
func getRemoteConfig(expandEnvVars bool, initialConfig *Config, log *server.Logger) (*Config, error) {
	if err := initialConfig.AgentManagement.Validate(); err != nil {
		return nil, fmt.Errorf("invalid initial config: %w", err)
	}
	remoteConfigBytes, err := fetchFromApi(initialConfig)
	if err != nil {
		level.Error(log).Log("msg", "could not fetch from API, falling back to cache", "err", err)
		return getCachedRemoteConfig(initialConfig.AgentManagement.CacheLocation, expandEnvVars)
	}
	var remoteConfig Config

	err = LoadBytes(remoteConfigBytes, expandEnvVars, &remoteConfig)
	if err != nil {
		level.Error(log).Log("msg", "could not load the response from the API, falling back to cache", "err", err)
		return getCachedRemoteConfig(initialConfig.AgentManagement.CacheLocation, expandEnvVars)
	}
	level.Info(log).Log("msg", "fetched and loaded remote config from API")

	if err = cacheRemoteConfig(initialConfig.AgentManagement.CacheLocation, remoteConfigBytes); err != nil {
		level.Error(log).Log("err", fmt.Errorf("could not cache config locally: %w", err))
	}
	return &remoteConfig, nil
}

func getCachedRemoteConfig(cachePath string, expandEnvVars bool) (*Config, error) {
	cachePath = filepath.Join(cachePath, cacheFilename)
	var cachedConfig Config
	if err := LoadFile(cachePath, expandEnvVars, &cachedConfig); err != nil {
		return nil, fmt.Errorf("error trying to load cached remote config from file: %w", err)
	}
	return &cachedConfig, nil
}

func cacheRemoteConfig(cachePath string, remoteConfigBytes []byte) error {
	cachePath = filepath.Join(cachePath, cacheFilename)
	return os.WriteFile(cachePath, remoteConfigBytes, 0666)
}

// fetchFromApi fetches the raw bytes from the API based on the protocol specified in c.
func fetchFromApi(c *Config) ([]byte, error) {
	switch p := c.AgentManagement.Protocol; {
	case p == "http":
		return fetchConfig(c)
	default:
		return nil, fmt.Errorf("unsupported protocol for agent management api: %s", p)
	}
}

// fetchConfig fetches the raw bytes of the config from the API specified in c.
func fetchConfig(c *Config) ([]byte, error) {
	httpClientConfig := &config.HTTPClientConfig{
		BasicAuth: &c.AgentManagement.BasicAuth,
	}

	dir, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get current working directory: %w", err)
	}
	httpClientConfig.SetDirectory(dir)

	remoteOpts := &remoteOpts{
		HTTPClientConfig: httpClientConfig,
	}

	url, err := c.AgentManagement.fullUrl()
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

// fullUrl creates and returns the URL that should be used when querying the Agent Management API,
// including the namespace, base config id, and any labels that have been specified.
func (am *AgentManagement) fullUrl() (string, error) {
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
func (am *AgentManagement) SleepTime() (time.Duration, error) {
	return time.ParseDuration(am.PollingInterval)
}

// Validate checks that necessary portions of the config have been set.
func (am *AgentManagement) Validate() error {
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
