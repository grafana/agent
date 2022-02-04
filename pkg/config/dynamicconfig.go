package config

// LoaderConfig is used by dynamic configuration
type LoaderConfig struct {
	// Sources is used to define sources for variables using gomplate
	Sources []Datasource `yaml:"datasources"`

	// TemplatePaths is the "directory" to look for templates in, they will be found and matched to configs but various
	// naming conventions. They can be S3/gcp, or file based resources. The directory structure is NOT walked.
	TemplatePaths []string `yaml:"template_paths"`

	AgentFilter           string `yaml:"agent_filter,omitempty"`
	ServerFilter          string `yaml:"sever_filter,omitempty"`
	MetricsFilter         string `yaml:"metrics_filter,omitempty"`
	MetricsInstanceFilter string `yaml:"metrics_instance_filter,omitempty"`
	IntegrationsFilter    string `yaml:"integrations_filter,omitempty"`
	LogsFilter            string `yaml:"logs_filter,omitempty"`
	TracesFilter          string `yaml:"traces_filter,omitempty"`
}

// Datasource is used for gomplate and can be used for a variety of resources.
type Datasource struct {
	Name string `yaml:"name"`
	URL  string `yaml:"url"`
}
