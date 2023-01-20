package cloudwatch_exporter

import (
	"fmt"
	"github.com/grafana/agent/pkg/integrations"
	integrations_v2 "github.com/grafana/agent/pkg/integrations/v2"
	"github.com/grafana/agent/pkg/integrations/v2/metricsutils"
	"time"

	"github.com/go-kit/log"
)

func init() {
	integrations.RegisterIntegration(&Config{})
	integrations_v2.RegisterLegacy(&Config{}, integrations_v2.TypeMultiplex, metricsutils.NewNamedShim("cloudwatch_exporter"))
}

// Config is the configuration for the CloudWatch metrics integration
type Config struct {
	STSRegion string          `yaml:"sts_region"`
	Discovery DiscoveryConfig `yaml:"discovery"`
	Static    []StaticJob     `yaml:"static"`
}

// DiscoveryConfig configures scraping jobs that will auto-discover metrics dimensions for a given service.
type DiscoveryConfig struct {
	ExportedTags TagsPerNamespace `yaml:"exported_tags"`
	Jobs         []*DiscoveryJob  `yaml:"jobs"`
}

// TagsPerNamespace represents for each namespace, a list of tags that will be exported as labels in each metric.
type TagsPerNamespace map[string][]string

// DiscoveryJob configures a discovery job for a given service.
type DiscoveryJob struct {
	InlineRegionAndRoles `yaml:",inline"`
	InlineCustomTags     `yaml:",inline"`
	Type                 string   `yaml:"type"`
	Metrics              []Metric `yaml:"metrics"`
}

// StaticJob will scrape metrics that match all defined dimensions.
type StaticJob struct {
	InlineRegionAndRoles `yaml:",inline"`
	InlineCustomTags     `yaml:",inline"`
	Name                 string      `yaml:"name"`
	Namespace            string      `yaml:"namespace"`
	Dimensions           []Dimension `yaml:"dimensions"`
	Metrics              []Metric    `yaml:"metrics"`
}

// InlineRegionAndRoles exposes for each supported job, the AWS regions and IAM roles in which the agent should perform the
// scrape.
type InlineRegionAndRoles struct {
	Regions []string `yaml:"regions"`
	Roles   []Role   `yaml:"roles"`
}

type InlineCustomTags struct {
	CustomTags []Tag `yaml:"custom_tags"`
}

type Role struct {
	RoleArn    string `yaml:"role_arn"`
	ExternalID string `yaml:"external_id"`
}

type Dimension struct {
	Name  string `yaml:"name"`
	Value string `yaml:"value"`
}

type Tag struct {
	Key   string `yaml:"key"`
	Value string `yaml:"value"`
}

type Metric struct {
	Name       string        `yaml:"name"`
	Statistics []string      `yaml:"statistics"`
	Period     time.Duration `yaml:"period"`
}

// Name returns the name of the integration this config is for.
func (c *Config) Name() string {
	return "cloudwatch_exporter"
}

func (c *Config) InstanceKey(agentKey string) (string, error) {
	return c.Name(), nil
}

// NewIntegration creates a new integration from the config.
func (c *Config) NewIntegration(l log.Logger) (integrations.Integration, error) {
	exporterConfig, err := ToYACEConfig(c)
	if err != nil {
		return nil, fmt.Errorf("invalid cloudwatch exporter configuration: %w", err)
	}
	collector := newCloudwatchCollector(l, exporterConfig)

	return integrations.NewCollectorIntegration(
		c.Name(),
		integrations.WithCollectors(collector),
	), nil
}
