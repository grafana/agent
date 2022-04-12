package config

import "time"

type Config struct {
	Nodes []Node `yaml:"nodes,omitempty"`
}

type Node struct {
	Name     string   `yaml:"name,omitempty"`
	FilePath string   `yaml:"file_path,omitempty"`
	Outputs  []string `yaml:"outputs,omitempty"`

	MetricGenerator *MetricGenerator `yaml:"metric_generator,omitempty"`
	MetricFilter    *MetricFilter    `yaml:"metric_filter,omitempty"`

	AgentLogs     *AgentLogs     `yaml:"agent_logs,omitempty"`
	LogFileWriter *LogFileWriter `yaml:"log_file_writer,omitempty"`

	Github *Github `yaml:"github,omitempty"`

	FakeMetricRemoteWrite *FakeRemoteWrite       `yaml:"fake_metric_remote_write,omitempty"`
	SimpleRemoteWrite     *SimpleRemoteWrite     `yaml:"simple_metric_remote_write,omitempty"`
	PrometheusRemoteWrite *PrometheusRemoteWrite `yaml:"prometheus_remote_write,omitempty"`
}

type MetricGenerator struct {
	Format        string        `yaml:"format"`
	SpawnInterval time.Duration `yaml:"spawn_interval,omitempty"`
}

type MetricFilter struct {
	Filters []MetricFilterFilter `yaml:"filters,omitempty"`
}

type MetricFilterFilter struct {
	MatchField string `yaml:"match_field,omitempty"`
	Action     string `yaml:"action,omitempty"`
	Regex      string `yaml:"regex,omitempty"`
	AddValue   string `yaml:"add_value,omitempty"`
	AddLabel   string `yaml:"add_label,omitempty"`
}

type FakeRemoteWrite struct {
	Credential Credential `yaml:"credential,omitempty"`
}

type SimpleRemoteWrite struct {
	URL string `yaml:"url,omitempty"`
}

type Credential struct {
	URL      string `yaml:"url,omitempty"`
	Username string `yaml:"username,omitempty"`
	Password string `yaml:"password,omitempty"`
}

type AgentLogs struct {
}

type LogFileWriter struct {
	Path string `yaml:"path,omitempty"`
}

type Github struct {
	ApiURL       string   `yaml:"api_url,omitempty"`
	Repositories []string `yaml:"repositories,omitempty"`
}

type PrometheusRemoteWrite struct {
	WalDir   string `yaml:"wal_dir,omitempty"`
	URL      string `yaml:"url,omitempty"`
	Username string `yaml:"username,omitempty"`
	Password string `yaml:"password,omitempty"`
}
