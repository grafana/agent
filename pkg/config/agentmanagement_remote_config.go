package config

import (
	"github.com/grafana/agent/pkg/integrations"
	"github.com/grafana/agent/pkg/logs"
	"github.com/grafana/agent/pkg/metrics/instance"
	"github.com/grafana/loki/clients/pkg/promtail/scrapeconfig"
	pc "github.com/prometheus/prometheus/config"
	"gopkg.in/yaml.v2"
)

type (
	RemoteConfig struct {
		BaseConfig BaseConfigContent `json:"base_config" yaml:"base_config"`
		Snippets   []Snippet         `json:"snippets" yaml:"snippets"`
	}

	// BaseConfigContent is the content of a base config
	BaseConfigContent string

	// Snippet is a snippet of configuration returned by the config API.
	Snippet struct {
		// Config is the snippet of config to be included.
		Config string `json:"config" yaml:"config"`
	}

	// SnippetContent defines the internal structure of a snippet configuration.
	SnippetContent struct {
		// MetricsScrapeConfigs is a YAML containing list of metrics scrape configs.
		MetricsScrapeConfigs []*pc.ScrapeConfig `yaml:"metrics_scrape_configs,omitempty"`

		// LogsScrapeConfigs is a YAML containing list of logs scrape configs.
		LogsScrapeConfigs []scrapeconfig.Config `yaml:"logs_scrape_configs,omitempty"`

		// IntegrationConfigs is a YAML containing list of integrations.
		IntegrationConfigs integrations.ManagerConfig `yaml:"integration_configs,omitempty"`
	}
)

func NewRemoteConfig(buf []byte) (*RemoteConfig, error) {
	rc := &RemoteConfig{}
	err := yaml.Unmarshal(buf, rc)
	if err != nil {
		return nil, err
	}
	return rc, nil
}

// BuildAgentConfig builds an agent configuration from a base config and a list of snippets
func (rc *RemoteConfig) BuildAgentConfig() (*Config, error) {
	c := DefaultConfig()
	err := yaml.Unmarshal([]byte(rc.BaseConfig), &c)
	if err != nil {
		return nil, err
	}

	// For now Agent Management only supports integrations v1
	if err := c.Integrations.setVersion(integrationsVersion1); err != nil {
		return nil, err
	}

	err = appendSnippets(&c, rc.Snippets)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func appendSnippets(c *Config, snippets []Snippet) error {
	metricsConfigs := instance.DefaultConfig
	metricsConfigs.Name = "snippets"
	logsConfigs := logs.InstanceConfig{
		Name:         "snippets",
		ScrapeConfig: []scrapeconfig.Config{},
	}
	logsConfigs.Initialize()
	integrationConfigs := integrations.DefaultManagerConfig()

	for _, snippet := range snippets {
		var snippetContent SnippetContent
		err := yaml.Unmarshal([]byte(snippet.Config), &snippetContent)
		if err != nil {
			return err
		}
		metricsConfigs.ScrapeConfigs = append(metricsConfigs.ScrapeConfigs, snippetContent.MetricsScrapeConfigs...)
		logsConfigs.ScrapeConfig = append(logsConfigs.ScrapeConfig, snippetContent.LogsScrapeConfigs...)
		integrationConfigs.Integrations = append(integrationConfigs.Integrations, snippetContent.IntegrationConfigs.Integrations...)
	}
	if len(metricsConfigs.ScrapeConfigs) > 0 {
		c.Metrics.Configs = append(c.Metrics.Configs, metricsConfigs)
	}

	if len(logsConfigs.ScrapeConfig) > 0 {
		// rc.Config.Logs is initialized as nil, so we need to check if it's nil before appending
		if c.Logs == nil {
			c.Logs = &logs.Config{
				Configs: []*logs.InstanceConfig{},
			}
		}
		c.Logs.Configs = append(c.Logs.Configs, &logsConfigs)
	}

	c.Integrations.configV1.Integrations = append(c.Integrations.configV1.Integrations, integrationConfigs.Integrations...)
	return nil
}
