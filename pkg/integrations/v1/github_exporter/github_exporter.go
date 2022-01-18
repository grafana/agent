package github_exporter //nolint:golint

import (
	"fmt"
	"net/url"

	"github.com/grafana/agent/pkg/integrations/shared"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	gh_config "github.com/infinityworks/github-exporter/config"
	"github.com/infinityworks/github-exporter/exporter"
	config_util "github.com/prometheus/common/config"
)

// DefaultConfig holds the default settings for the github_exporter integration
var DefaultConfig Config = Config{
	APIURL: "https://api.github.com",
}

// Config controls github_exporter
type Config struct {
	// URL for the github API
	APIURL string `yaml:"api_url,omitempty"`

	// A list of github repositories for which to collect metrics.
	Repositories []string `yaml:"repositories,omitempty"`

	// A list of github organizations for which to collect metrics.
	Organizations []string `yaml:"organizations,omitempty"`

	// A list of github users for which to collect metrics.
	Users []string `yaml:"users,omitempty"`

	// A github authentication token that allows the API to be queried more often.
	APIToken config_util.Secret `yaml:"api_token,omitempty"`

	// A path to a file containing a github authentication token that allows the API to be queried more often. If supplied, this supercedes `api_token`
	APITokenFile string `yaml:"api_token_file,omitempty"`
}

// Name returns the name of the integration that this config represents.
func (c *Config) Name() string {
	return "github_exporter"
}

// InstanceKey returns the hostname:port of the GitHub API server.
func (c *Config) InstanceKey(agentKey string) (string, error) {
	u, err := url.Parse(c.APIURL)
	if err != nil {
		return "", fmt.Errorf("could not parse url: %w", err)
	}
	return u.Host, nil
}

// NewIntegration creates a new github_exporter
func (c *Config) NewIntegration(logger log.Logger) (shared.Integration, error) {
	return New(logger, c)
}

// New creates a new github_exporter integration.
func New(logger log.Logger, c *Config) (shared.Integration, error) {

	conf := gh_config.Config{}
	err := conf.SetAPIURL(c.APIURL)
	if err != nil {
		level.Error(logger).Log("msg", "api url is invalid", "err", err)
		return nil, err
	}
	conf.SetRepositories(c.Repositories)
	conf.SetOrganisations(c.Organizations)
	conf.SetUsers(c.Users)
	if c.APIToken != "" {
		conf.SetAPIToken(string(c.APIToken))
	}
	if c.APITokenFile != "" {
		err = conf.SetAPITokenFromFile(c.APITokenFile)
		if err != nil {
			level.Error(logger).Log("msg", "unable to load Github API token from file", "err", err)
			return nil, err
		}
	}

	ghExporter := exporter.Exporter{
		APIMetrics: exporter.AddMetrics(),
		Config:     conf,
	}

	return shared.NewCollectorIntegration(
		c.Name(),
		shared.WithCollectors(&ghExporter),
	), nil
}
