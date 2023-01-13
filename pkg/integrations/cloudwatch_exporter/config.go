package cloudwatch_exporter

import "time"

type Config struct {
	STSRegion string          `yaml:"stsRegion"`
	Discovery DiscoveryConfig `yaml:"discovery"`
}

// TagsPerNamespace represents for each namespace, a list of tags that will be exported as labels in each metric.
type TagsPerNamespace map[string][]string

type DiscoveryConfig struct {
	ExportedTags TagsPerNamespace `yaml:"exportedTags"`
	Jobs         []*DiscoveryJob  `yaml:"jobs"`
}

type DiscoveryJob struct {
	RegionAndRoles `yaml:",inline"`
	Type           string        `yaml:"type"`
	Metrics        []Metric      `yaml:"metrics"`
	ScrapeInterval time.Duration `yaml:"scrapeInterval"`
}

type StaticJob struct {
	RegionAndRoles `yaml:",inline"`
	Namespace      string      `yaml:"namespace"`
	Dimensions     []Dimension `yaml:"dimensions"`
	Metrics        []Metric    `yaml:"metrics"`
}

type Dimension struct {
	Name  string `yaml:"name"`
	Value string `yaml:"value"`
}

type Role struct {
	RoleArn    string `yaml:"roleArn"`
	ExternalID string `yaml:"externalID"`
}

type Metric struct {
	Name       string
	Statistics []string
}

// RegionAndRoles exposes for each supported job, the AWS regions and IAM roles in which the agent should perform the
// scrape.
type RegionAndRoles struct {
	Regions []string `yaml:"regions"`
	Roles   []Role   `yaml:"roles"`
}
