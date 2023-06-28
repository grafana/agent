package config

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-kit/log/level"
	"github.com/grafana/agent/pkg/config/instrumentation"
	"github.com/grafana/agent/pkg/server"
	"github.com/prometheus/common/config"
	"gopkg.in/yaml.v2"
)

const cacheFilename = "remote-config-cache.yaml"
const apiPath = "/agent-management/api/agent/v2"

type remoteConfigProvider interface {
	GetCachedRemoteConfig() ([]byte, error)
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
	InitialConfigHash string `json:"initial_config_hash"`
	Config            string `json:"config"`
}

func hashInitialConfig(am AgentManagementConfig) (string, error) {
	marshalled, err := yaml.Marshal(am)
	if err != nil {
		return "", fmt.Errorf("could not marshal initial config: %w", err)
	}
	hashed := sha256.Sum256(marshalled)
	return hex.EncodeToString(hashed[:]), nil
}

// initialConfigHashCheck checks if the hash of initialConfig matches what is stored in configCache.InitialConfigHash.
// If an error is encountered while hashing initialConfig or the hashes do not match, initialConfigHashCheck
// returns an error. Otherwise, it returns nil.
func initialConfigHashCheck(initialConfig AgentManagementConfig, configCache remoteConfigCache) error {
	initialConfigHash, err := hashInitialConfig(initialConfig)
	if err != nil {
		return err
	}

	if !(configCache.InitialConfigHash == initialConfigHash) {
		return errors.New("invalid remote config cache: initial config hashes don't match")
	}
	return nil
}

// GetCachedRemoteConfig retrieves the cached remote config from the location specified
// in r.AgentManagement.CacheLocation
func (r remoteConfigHTTPProvider) GetCachedRemoteConfig() ([]byte, error) {
	cachePath := filepath.Join(r.InitialConfig.RemoteConfiguration.CacheLocation, cacheFilename)

	var configCache remoteConfigCache
	buf, err := os.ReadFile(cachePath)

	if err != nil {
		return nil, fmt.Errorf("error reading remote config cache: %w", err)
	}

	if err := json.Unmarshal(buf, &configCache); err != nil {
		return nil, fmt.Errorf("error trying to load cached remote config from file: %w", err)
	}

	if err = initialConfigHashCheck(*r.InitialConfig, configCache); err != nil {
		return nil, err
	}

	return []byte(configCache.Config), nil
}

// CacheRemoteConfig caches the remote config to the location specified in
// r.AgentManagement.CacheLocation
func (r remoteConfigHTTPProvider) CacheRemoteConfig(remoteConfigBytes []byte) error {
	cachePath := filepath.Join(r.InitialConfig.RemoteConfiguration.CacheLocation, cacheFilename)
	initialConfigHash, err := hashInitialConfig(*r.InitialConfig)
	if err != nil {
		return err
	}
	configCache := remoteConfigCache{
		InitialConfigHash: initialConfigHash,
		Config:            string(remoteConfigBytes),
	}
	marshalled, err := json.Marshal(configCache)
	if err != nil {
		return fmt.Errorf("could not marshal remote config cache: %w", err)
	}
	return os.WriteFile(cachePath, marshalled, 0666)
}

// readPasswordFile reads the specified password file, trims any whitespace, and returns the password as a string.
func readPasswordFile(passwordFile string) (string, error) {
	password, err := os.ReadFile(passwordFile)
	if err != nil {
		return "", errors.New("error reading password file")
	}
	cleanedPassword := strings.TrimSpace(string(password))
	return cleanedPassword, nil
}

// FetchRemoteConfig fetches the raw bytes of the config from a remote API using
// the values in r.AgentManagement.
func (r remoteConfigHTTPProvider) FetchRemoteConfig() ([]byte, error) {
	// Create the full url, possibly including labels as query params
	url, err := r.InitialConfig.fullUrl()
	if err != nil {
		return nil, fmt.Errorf("error trying to create full url: %w", err)
	}

	// Create the request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	// Add basic auth
	password, err := readPasswordFile(r.InitialConfig.BasicAuth.PasswordFile)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(r.InitialConfig.BasicAuth.Username, password)

	// Set headers for label management, if enabled
	if r.InitialConfig.RemoteConfiguration.LabelManagementEnabled && r.InitialConfig.RemoteConfiguration.AgentID != "" {
		req.Header.Add("X-LabelManagementEnabled", "1")
		req.Header.Add("X-AgentID", r.InitialConfig.RemoteConfiguration.AgentID)
	}

	client := &http.Client{
		Timeout: 5 * time.Second,
	}
	response, err := client.Do(req)
	if err != nil {
		instrumentation.InstrumentRemoteConfigFetchError()
		return nil, fmt.Errorf("request failed: %w", err)
	}
	instrumentation.InstrumentRemoteConfigFetch(response.StatusCode)

	if response.StatusCode == http.StatusTooManyRequests {
		retryAfter := response.Header.Get("Retry-After")
		if retryAfter == "" {
			return nil, fmt.Errorf("server indicated to retry, but no Retry-After header was provided")
		}
		retryAfterDuration, err := time.ParseDuration(retryAfter)
		if err != nil {
			return nil, fmt.Errorf("server indicated to retry, but Retry-After header was not a valid duration: %w", err)
		}
		return nil, retryAfterError{retryAfter: retryAfterDuration}
	}

	if err != nil {
		return nil, fmt.Errorf("error retrieving remote config: %w", err)
	}
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}
	return body, nil
}

type labelMap map[string]string

type RemoteConfiguration struct {
	Labels                 labelMap `yaml:"labels"`
	LabelManagementEnabled bool     `yaml:"label_management_enabled"`
	AgentID                string   `yaml:"agent_id"`
	Namespace              string   `yaml:"namespace"`
	CacheLocation          string   `yaml:"cache_location"`
}

type AgentManagementConfig struct {
	Enabled         bool             `yaml:"-"` // Derived from enable-features=agent-management
	Host            string           `yaml:"host"`
	BasicAuth       config.BasicAuth `yaml:"basic_auth"`
	Protocol        string           `yaml:"protocol"`
	PollingInterval time.Duration    `yaml:"polling_interval"`

	RemoteConfiguration RemoteConfiguration `yaml:"remote_configuration"`
}

// getRemoteConfig gets the remote config specified in the initial config, falling back to a local, cached copy
// of the remote config if the request to the remote fails. If both fail, an empty config and an
// error will be returned.
func getRemoteConfig(expandEnvVars bool, configProvider remoteConfigProvider, log *server.Logger, fs *flag.FlagSet, retry bool) (*Config, error) {
	remoteConfigBytes, err := configProvider.FetchRemoteConfig()
	if err != nil {
		var retryAfterErr retryAfterError
		if errors.As(err, &retryAfterErr) && retry {
			level.Error(log).Log("msg", "received retry-after from API, sleeping and falling back to cache", "retry-after", retryAfterErr.retryAfter)
			time.Sleep(retryAfterErr.retryAfter)
			return getRemoteConfig(expandEnvVars, configProvider, log, fs, false)
		}
		level.Error(log).Log("msg", "could not fetch from API, falling back to cache", "err", err)
		return getCachedRemoteConfig(expandEnvVars, configProvider, fs, log)
	}

	config, err := loadRemoteConfig(remoteConfigBytes, expandEnvVars, fs)
	if err != nil {
		level.Error(log).Log("msg", "could not load remote config, falling back to cache", "err", err)
		return getCachedRemoteConfig(expandEnvVars, configProvider, fs, log)
	}

	level.Info(log).Log("msg", "fetched and loaded remote config from API")

	if err = configProvider.CacheRemoteConfig(remoteConfigBytes); err != nil {
		level.Error(log).Log("err", fmt.Errorf("could not cache config locally: %w", err))
	}
	return config, nil
}

// getCachedRemoteConfig gets the cached remote config, falling back to the default config if the cache is invalid or not found.
func getCachedRemoteConfig(expandEnvVars bool, configProvider remoteConfigProvider, fs *flag.FlagSet, log *server.Logger) (*Config, error) {
	rc, err := configProvider.GetCachedRemoteConfig()
	if err != nil {
		level.Error(log).Log("msg", "could not get cached remote config, falling back to default (empty) config", "err", err)
		d := DefaultConfig()
		instrumentation.InstrumentAgentManagementConfigFallback("empty_config")
		return &d, nil
	}
	instrumentation.InstrumentAgentManagementConfigFallback("cache")
	return loadRemoteConfig(rc, expandEnvVars, fs)
}

// loadRemoteConfig parses and validates the remote config, both syntactically and semantically.
func loadRemoteConfig(remoteConfigBytes []byte, expandEnvVars bool, fs *flag.FlagSet) (*Config, error) {
	expandedRemoteConfigBytes, err := performEnvVarExpansion(remoteConfigBytes, expandEnvVars)
	if err != nil {
		instrumentation.InstrumentInvalidRemoteConfig("env_var_expansion")
		return nil, fmt.Errorf("could not expand env vars for remote config: %w", err)
	}

	remoteConfig, err := NewRemoteConfig(expandedRemoteConfigBytes)
	if err != nil {
		instrumentation.InstrumentInvalidRemoteConfig("invalid_yaml")
		return nil, fmt.Errorf("could not unmarshal remote config: %w", err)
	}

	config, err := remoteConfig.BuildAgentConfig()
	if err != nil {
		instrumentation.InstrumentInvalidRemoteConfig("invalid_remote_config")
		return nil, fmt.Errorf("could not build agent config: %w", err)
	}

	if err = config.Validate(fs); err != nil {
		instrumentation.InstrumentInvalidRemoteConfig("semantically_invalid_agent_config")
		return nil, fmt.Errorf("semantically invalid config received from the API: %w", err)
	}
	return config, nil
}

// newRemoteConfigProvider creates a remoteConfigProvider based on the protocol
// specified in c.AgentManagement
func newRemoteConfigProvider(c *Config) (*remoteConfigHTTPProvider, error) {
	switch p := c.AgentManagement.Protocol; {
	case p == "https" || p == "http":
		return newRemoteConfigHTTPProvider(c)
	default:
		return nil, fmt.Errorf("unsupported protocol for agent management api: %s", p)
	}
}

// fullUrl creates and returns the URL that should be used when querying the Agent Management API,
// including the namespace, base config id, and any labels that have been specified.
func (am *AgentManagementConfig) fullUrl() (string, error) {
	fullPath, err := url.JoinPath(am.Protocol+"://", am.Host, apiPath, "namespace", am.RemoteConfiguration.Namespace, "remote_config")
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

// SleepTime returns the duration in between config fetches.
func (am *AgentManagementConfig) SleepTime() time.Duration {
	return am.PollingInterval
}

// jitterTime returns a random duration in the range [0, am.PollingInterval).
func (am *AgentManagementConfig) JitterTime() time.Duration {
	return time.Duration(rand.Int63n(int64(am.PollingInterval)))
}

// Validate checks that necessary portions of the config have been set.
func (am *AgentManagementConfig) Validate() error {
	if am.BasicAuth.Username == "" || am.BasicAuth.PasswordFile == "" {
		return errors.New("both username and password_file fields must be specified")
	}

	if am.PollingInterval <= 0 {
		return fmt.Errorf("polling interval must be >0")
	}

	if am.RemoteConfiguration.Namespace == "" {
		return errors.New("namespace must be specified in 'remote_configuration' block of the config")
	}

	if am.RemoteConfiguration.CacheLocation == "" {
		return errors.New("path to cache must be specified in 'agent_management.remote_configuration.cache_location'")
	}

	return nil
}
