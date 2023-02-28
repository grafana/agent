package github

import (
	"github.com/grafana/agent/component"
	"github.com/grafana/agent/component/prometheus/exporter"
	"github.com/grafana/agent/pkg/flow/rivertypes"
	"github.com/grafana/agent/pkg/integrations"
	"github.com/grafana/agent/pkg/integrations/github_exporter"
	config_util "github.com/prometheus/common/config"
)

func init() {
	component.Register(component.Registration{
		Name:    "prometheus.exporter.github",
		Args:    Config{},
		Exports: exporter.Exports{},
		Build:   exporter.New(createExporter, "github"),
	})
}

func createExporter(opts component.Options, args component.Arguments) (integrations.Integration, error) {
	cfg := args.(Config)
	return cfg.Convert().NewIntegration(opts.Logger)
}

// DefaultConfig holds non-zero default options for the Config when it is
// unmarshaled from river.
var DefaultConfig = Config{
	APIURL: github_exporter.DefaultConfig.APIURL,
}

type Config struct {
	APIURL        string            `river:"api_url,attr,optional"`
	Repositories  []string          `river:"repositories,attr,optional"`
	Organizations []string          `river:"organizations,attr,optional"`
	Users         []string          `river:"users,attr,optional"`
	APIToken      rivertypes.Secret `river:"api_token,attr,optional"`
	APITokenFile  string            `river:"api_token_file,attr,optional"`
}

// UnmarshalRiver implements River unmarshalling for Config.
func (c *Config) UnmarshalRiver(f func(interface{}) error) error {
	*c = DefaultConfig

	type cfg Config
	return f((*cfg)(c))
}

func (c *Config) Convert() *github_exporter.Config {
	return &github_exporter.Config{
		APIURL:        c.APIURL,
		Repositories:  c.Repositories,
		Organizations: c.Organizations,
		Users:         c.Users,
		APIToken:      config_util.Secret(c.APIToken),
		APITokenFile:  c.APITokenFile,
	}
}
