package config

import (
	"github.com/alecthomas/units"
	"github.com/grafana/agent/component/common/config"
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/discovery"
)

// // ScrapeConfig configures a scraping unit for Prometheus.
// type ScrapeConfig struct {
// 	// The job name to which the job label is set by default.
// 	JobName string `yaml:"job_name"`
// 	// Indicator whether the scraped metrics should remain unmodified.
// 	HonorLabels bool `yaml:"honor_labels,omitempty"`
// 	// Indicator whether the scraped timestamps should be respected.
// 	HonorTimestamps bool `yaml:"honor_timestamps"`
// 	// A set of query parameters with which the target is scraped.
// 	Params url.Values `yaml:"params,omitempty"`
// 	// How frequently to scrape the targets of this scrape config.
// 	ScrapeInterval model.Duration `yaml:"scrape_interval,omitempty"`
// 	// The timeout for scraping targets of this config.
// 	ScrapeTimeout model.Duration `yaml:"scrape_timeout,omitempty"`
// 	// Whether to scrape a classic histogram that is also exposed as a native histogram.
// 	ScrapeClassicHistograms bool `yaml:"scrape_classic_histograms,omitempty"`
// 	// The HTTP resource path on which to fetch metrics from targets.
// 	MetricsPath string `yaml:"metrics_path,omitempty"`
// 	// The URL scheme with which to fetch metrics from targets.
// 	Scheme string `yaml:"scheme,omitempty"`
// 	// An uncompressed response body larger than this many bytes will cause the
// 	// scrape to fail. 0 means no limit.
// 	BodySizeLimit units.Base2Bytes `yaml:"body_size_limit,omitempty"`
// 	// More than this many samples post metric-relabeling will cause the scrape to
// 	// fail. 0 means no limit.
// 	SampleLimit uint `yaml:"sample_limit,omitempty"`
// 	// More than this many targets after the target relabeling will cause the
// 	// scrapes to fail. 0 means no limit.
// 	TargetLimit uint `yaml:"target_limit,omitempty"`
// 	// More than this many labels post metric-relabeling will cause the scrape to
// 	// fail. 0 means no limit.
// 	LabelLimit uint `yaml:"label_limit,omitempty"`
// 	// More than this label name length post metric-relabeling will cause the
// 	// scrape to fail. 0 means no limit.
// 	LabelNameLengthLimit uint `yaml:"label_name_length_limit,omitempty"`
// 	// More than this label value length post metric-relabeling will cause the
// 	// scrape to fail. 0 means no limit.
// 	LabelValueLengthLimit uint `yaml:"label_value_length_limit,omitempty"`
// 	// More than this many buckets in a native histogram will cause the scrape to
// 	// fail.
// 	NativeHistogramBucketLimit uint `yaml:"native_histogram_bucket_limit,omitempty"`
//
// 	// We cannot do proper Go type embedding below as the parser will then parse
// 	// values arbitrarily into the overflow maps of further-down types.
//
// 	ServiceDiscoveryConfigs discovery.Configs       `yaml:"-"`
// 	HTTPClientConfig        config.HTTPClientConfig `yaml:",inline"`
//
// 	// List of target relabel configurations.
// 	RelabelConfigs []*relabel.Config `yaml:"relabel_configs,omitempty"`
// 	// List of metric relabel configurations.
// 	MetricRelabelConfigs []*relabel.Config `yaml:"metric_relabel_configs,omitempty"`
// }
// This config is DTO for above ScrapeConfig
type ScrapeConfig struct {
	JobName                 string            `river:"job_name,string,optional"`
	HonorLabels             bool              `river:"honor_labels,bool,optional"`
	HonorTimestamps         bool              `river:"honor_timestamps,bool,optional"`
	Params                  config.URLValues  `river:"params,attr,optional"`
	ScrapeInterval          model.Duration    `river:"scrape_interval,attr,optional"`
	ScrapeTimeout           model.Duration    `river:"scrape_timeout,attr,optional"`
	MetricsPath             string            `river:"metrics_path,string,optional"`
	Scheme                  string            `river:"scheme,string,optional"`
	BodySizeLimit           units.Base2Bytes  `river:"body_size_limit,string,optional"`
	SampleLimit             uint              `river:"sample_limit,number,optional"`
	TargetLimit             uint              `river:"target_limit,number,optional"`
	LabelLimit              uint              `river:"label_limit,number,optional"`
	LabelNameLength         uint              `river:"label_name_length_limit,number,optional"`
	LabelValueLengthLimit   uint              `river:"label_value_length_limit,number,optional"`
	NativeHistogramBucket   uint              `river:"native_histogram_bucket_limit,number,optional"`
	ServiceDiscoveryConfigs discovery.Configs `river:"service_discovery_configs,block,optional"`
}
