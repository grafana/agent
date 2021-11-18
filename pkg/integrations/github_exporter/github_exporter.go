package github_exporter //nolint:golint

import (
	"fmt"
	"net/url"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/grafana/agent/pkg/integrations"
	"github.com/grafana/agent/pkg/integrations/config"
	gh_config "github.com/infinityworks/github-exporter/config"
	"github.com/infinityworks/github-exporter/exporter"
)

// DefaultConfig holds the default settings for the github_exporter integration
var DefaultConfig Config = Config{
	APIURL: "https://api.github.com",
}

// Config controls github_exporter
type Config struct {
	Common config.Common `yaml:",inline"`

	// URL for the github API
	APIURL string `yaml:"api_url,omitempty"`

	// A list of github repositories for which to collect metrics.
	Repositories []string `yaml:"repositories,omitempty"`

	// A list of github organizations for which to collect metrics.
	Organizations []string `yaml:"organizations,omitempty"`

	// A list of github users for which to collect metrics.
	Users []string `yaml:"users,omitempty"`

	// A github authentication token that allows the API to be queried more often.
	APIToken string `yaml:"api_token,omitempty"`

	// A path to a file containing a github authentication token that allows the API to be queried more often. If supplied, this supercedes `api_token`
	APITokenFile string `yaml:"api_token_file,omitempty"`
}

// UnmarshalYAML implements yaml.Unmarshaler for Config
func (c *Config) UnmarshalYAML(unmarshal func(interface{}) error) error {
	*c = DefaultConfig

	type plain Config
	return unmarshal((*plain)(c))
}

// Name returns the name of the integration that this config represents.
func (c *Config) Name() string {
	return "github_exporter"
}

// CommonConfig returns the common settings shared across all configs for
// integrations.
func (c *Config) CommonConfig() config.Common {
	return c.Common
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
func (c *Config) NewIntegration(logger log.Logger) (integrations.Integration, error) {
	return New(logger, c)
}

func init() {
	integrations.RegisterIntegration(&Config{})
}

// New creates a new github_exporter integration.
func New(logger log.Logger, c *Config) (integrations.Integration, error) {

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
		conf.SetAPIToken(c.APIToken)
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

	return integrations.NewCollectorIntegration(
		c.Name(),
		integrations.WithCollectors(&ghExporter),
	), nil
}
