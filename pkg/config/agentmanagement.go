package config

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/go-kit/log/level"
	"github.com/grafana/agent/pkg/config/instrumentation"
	"github.com/grafana/agent/pkg/server"
	"github.com/prometheus/common/config"
)

type BasicAuth struct {
	Username     string `yaml:"username"`
	PasswordFile string `yaml:"password_file"`
}

type RemoteConfiguration struct {
	Labels       []string `yaml:"labels"`
	Namespace    string   `yaml:"namespace"`
	BaseConfigId string   `yaml:"base_config_id"`
}

type AgentManagement struct {
	Enabled         bool      `yaml:"-"`
	Url             string    `yaml:"api_url"`
	BasicAuth       BasicAuth `yaml:"basic_auth"`
	Protocol        string    `yaml:"protocol"`
	PollingInterval string    `yaml:"polling_interval"`

	RemoteConfiguration RemoteConfiguration `yaml:"remote_configuration"`
}

func tryLog(log *server.Logger, lvl string, keyvals ...interface{}) {
	if log == nil {
		return
	}

	switch lvl {
	case "info":
		level.Info(log).Log(keyvals...)
	case "debug":
		level.Debug(log).Log(keyvals...)
	case "warn":
		level.Warn(log).Log(keyvals...)
	case "error":
		level.Error(log).Log(keyvals...)
	}
}

// Gets the remote config specified in the initial config, falling back to a local, cached copy
// of the remote config if the request to the remote fails. If both fail, an empty config and an
// error will be returned.
func GetRemoteConfig(dir string, expandEnvVars bool, initialConfig *Config, log *server.Logger) (*Config, error) {
	remoteConfigBytes, err := FetchFromApi(initialConfig)
	if err != nil {
		tryLog(log, "error", "msg", "could not fetch from API, falling back to cache", "err", err)
		return GetCachedRemoteConfig(dir, expandEnvVars)
	}
	var remoteConfig Config

	err = LoadBytes(remoteConfigBytes, expandEnvVars, &remoteConfig)
	if err != nil {
		tryLog(log, "error", "msg", "could not load the response from the API, falling back to cache", "err", err)
		return GetCachedRemoteConfig(dir, expandEnvVars)
	}
	tryLog(log, "info", "msg", "fetched and loaded new config from remote API")
	instrumentation.ConfigMetrics.InstrumentConfig(remoteConfigBytes)

	tryLog(log, "debug", "msg", "caching remote config")
	if err = cacheRemoteConfig(dir, remoteConfigBytes); err != nil {
		tryLog(log, "error", "err", fmt.Errorf("could not cache config locally: %w", err))
	}
	return &remoteConfig, nil
}

func GetCachedRemoteConfig(dir string, expandEnvVars bool) (*Config, error) {
	cachePath := filepath.Join(dir, "remote-config-cache.yaml")
	var cachedConfig Config
	if err := LoadFile(cachePath, expandEnvVars, &cachedConfig); err != nil {
		return nil, fmt.Errorf("error trying to load cached remote config from file: %w", err)
	}
	return &cachedConfig, nil
}

func cacheRemoteConfig(dir string, remoteConfigBytes []byte) error {
	cachePath := filepath.Join(dir, "remote-config-cache.yaml")
	return os.WriteFile(cachePath, remoteConfigBytes, 0666)
}

// Fetches the raw bytes from the API based on the protocol specified in c.
func FetchFromApi(c *Config) ([]byte, error) {
	switch p := c.AgentManagement.Protocol; {
	case p == "http":
		return FetchConfig(c)
	default:
		return nil, fmt.Errorf("unsupported procotol for agent management api: %s", p)
	}
}

// Fetches the raw bytes of the config from the API specified in c.
func FetchConfig(c *Config) ([]byte, error) {
	httpClientConfig := &config.HTTPClientConfig{
		BasicAuth: &config.BasicAuth{
			Username:     c.AgentManagement.BasicAuth.Username,
			PasswordFile: c.AgentManagement.BasicAuth.PasswordFile,
		},
	}

	dir, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get current working directory: %w", err)
	}
	httpClientConfig.SetDirectory(dir)

	remoteOpts := &remoteOpts{
		HTTPClientConfig: httpClientConfig,
	}

	url, err := c.AgentManagement.FullUrl()
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

// Fully creates and returns the URL that should be used when querying the Agent Management API,
// including the namespace, base config id, and any labels that have been specified.
func (am *AgentManagement) FullUrl() (string, error) {
	labelMap := am.LabelMap()
	fullPath, err := url.JoinPath(am.Url, am.RemoteConfiguration.Namespace, am.RemoteConfiguration.BaseConfigId, "remote_config")
	if err != nil {
		return "", fmt.Errorf("error trying to join url: %w", err)
	}
	u, err := url.Parse(fullPath)
	if err != nil {
		return "", fmt.Errorf("error trying to parse url: %w", err)
	}
	q := u.Query()
	for label, value := range labelMap {
		q.Add(label, value)
	}
	u.RawQuery = q.Encode()
	return u.String(), nil
}

// Parses the AgentManagement.RemoteConfiguration.Label string representing a list of comma-separated
// key:value pairs into a map.
//
// e.g. "key1:value1,key2:value2" -> {"key1": "value2", "key2": "value2"}
func (am *AgentManagement) LabelMap() map[string]string {
	labelMap := map[string]string{}

	if len(am.RemoteConfiguration.Labels) == 0 {
		return labelMap
	}
	pairs := am.RemoteConfiguration.Labels
	sort.Strings(pairs)
	for _, pair := range pairs {
		split := strings.Split(pair, ":")
		if len(split) != 2 {
			return nil
		}
		labelMap[split[0]] = split[1]
	}
	return labelMap
}

// Returns the duration in between config fetches.
func (am *AgentManagement) SleepTime() (time.Duration, error) {
	return time.ParseDuration(am.PollingInterval)
}

// Validates portions of the agent_management config
func (am *AgentManagement) Validate() error {
	if am.BasicAuth.Username == "" || am.BasicAuth.PasswordFile == "" {
		return errors.New("both username and password_file fields must be specified")
	}

	if _, err := time.ParseDuration(am.PollingInterval); err != nil {
		return fmt.Errorf("error trying to parse polling interval: %w", err)
	}

	if am.RemoteConfiguration.BaseConfigId == "" {
		return errors.New("base config id must be specified on the CLI with -agentmanagement.base_config_id=<id>")
	} else if am.RemoteConfiguration.Namespace == "" {
		return errors.New("namespace must be specified on the CLI with -agentmanagement.namespace=<namespace>")
	}

	return nil
}

// Fetches the config from the Agent Management API specified in c via an HTTP GET request.
func FetchConfigHTTPRaw(c *Config, tenantId string) ([]byte, error) {
	am := c.AgentManagement

	url, err := am.FullUrl()
	if err != nil {
		return nil, fmt.Errorf("error trying to create full url: %w", err)
	}

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("error trying to create http request: %w", err)
	}

	req.Header.Add("X-Scope-OrgID", tenantId)

	// Lifted from Prometheus code and updated to use os.ReadFile instead of ioutils
	bs, err := os.ReadFile(am.BasicAuth.PasswordFile)
	if err != nil {
		return nil, fmt.Errorf("unable to read basic auth password file %s: %s", am.BasicAuth.PasswordFile, err)
	}
	req.SetBasicAuth(am.BasicAuth.Username, strings.TrimSpace(string(bs)))

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("unable to complete http request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("non-200 response code from agent management API: %d", resp.StatusCode)
	}
	return io.ReadAll(resp.Body)
}
