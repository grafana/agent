package github_exporter //nolint:golint

import (
	"os"
	"strings"

	"github.com/go-kit/kit/log"
	"github.com/grafana/agent/pkg/integrations"
	"github.com/grafana/agent/pkg/integrations/config"
	gh_config "github.com/infinityworks/github-exporter/config"
	"github.com/infinityworks/github-exporter/exporter"
)

var DefaultConfig Config = Config{
	ApiUrl: "https://api.github.com",
}

// Config controls github_exporter
type Config struct {
	Common config.Common `yaml:",inline"`

  // URL for the github API
	ApiUrl string `yaml:"api_url,omitempty"`

  // A list of github repositories for which to collect metrics.
	Repositories []string `yaml:"repositories,omitempty"`

  // A list of github organizations for which to collect metrics.
	Organizations []string `yaml:"organizations,omitempty"`

  // A list of github users for which to collect metrics.
	Users []string `yaml:"users,omitempty"`

  // A github authentication token that allows the API to be queried more often.
  ApiToken string `yaml:"api_token,omitempty"`

  // A path to a file containing a github authentication token that allows the API to be queried more often. If supplied, this supercedes `api_token`
	ApiTokenFile string `yaml:"api_token_file,omitempty"`
}

func (c *Config) UnmarshalYAML(unmarshal func(interface{}) error) error {
	*c = DefaultConfig

	type plain Config
	return unmarshal((*plain)(c))
}

func (c *Config) Name() string {
	return "github_exporter"
}

func (c *Config) CommonConfig() config.Common {
	return c.Common
}

func (c *Config) NewIntegration(logger log.Logger) (integrations.Integration, error) {
	return New(logger, c)
}

func init() {
	integrations.RegisterIntegration(&Config{})
}

func New(logger log.Logger, c *Config) (integrations.Integration, error) {
	// It's not very pretty, but this exporter is configured entirely by environment
	// variables, and uses some private helper methods in it's config package to
	// assemble other key pieces of the config. Thus, we can't (easily) access the
	// config directly, and must assign environment variables.
	//
	// In an effort to avoid conflicts with other integrations, the environment vars
	// are unset immediately after being consumed.

	os.Setenv("API_URL", c.ApiUrl)
	os.Setenv("REPOS", strings.Join(c.Repositories, ", "))
	os.Setenv("ORGS", strings.Join(c.Organizations, ", "))
	os.Setenv("USERS", strings.Join(c.Users, ", "))
	if c.ApiToken != "" {
		os.Setenv("GITHUB_TOKEN", c.ApiToken)
	}

	if c.ApiTokenFile != "" {
		os.Setenv("GITHUB_TOKEN_FILE", c.ApiTokenFile)
	}

	conf := gh_config.Init()

	os.Unsetenv("API_URL")
	os.Unsetenv("REPOS")
	os.Unsetenv("ORGS")
	os.Unsetenv("USERS")
	os.Unsetenv("GITHUB_TOKEN")
	os.Unsetenv("GITHUB_TOKEN_FILE")

	gh_exporter := exporter.Exporter{
		APIMetrics: exporter.AddMetrics(),
		Config:     conf,
	}

	return integrations.NewCollectorIntegration(
		c.Name(),
		integrations.WithCollectors(&gh_exporter),
	), nil
}
