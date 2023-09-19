package instance

import (
	"github.com/grafana/agent/component/common/prometheus/config"
	"github.com/grafana/agent/component/common/relabel"
)

// type Config struct {
// 	Name                     string                      `yaml:"name,omitempty"`
// 	HostFilter               bool                        `yaml:"host_filter,omitempty"`
// 	HostFilterRelabelConfigs []*relabel.Config           `yaml:"host_filter_relabel_configs,omitempty"`
// 	ScrapeConfigs            []*config.ScrapeConfig      `yaml:"scrape_configs,omitempty"`
// 	RemoteWrite              []*config.RemoteWriteConfig `yaml:"remote_write,omitempty"`
//
// 	// How frequently the WAL should be truncated.
// 	WALTruncateFrequency time.Duration `yaml:"wal_truncate_frequency,omitempty"`
//
// 	// Minimum and maximum time series should exist in the WAL for.
// 	MinWALTime time.Duration `yaml:"min_wal_time,omitempty"`
// 	MaxWALTime time.Duration `yaml:"max_wal_time,omitempty"`
//
// 	RemoteFlushDeadline  time.Duration `yaml:"remote_flush_deadline,omitempty"`
// 	WriteStaleOnShutdown bool          `yaml:"write_stale_on_shutdown,omitempty"`
//
// 	global GlobalConfig `yaml:"-"`
// }
// This config is DTO for above Config
type Config struct {
	name                     string                 `river:"name,string,optional"`
	hostFilter               bool                   `river:"host_filter,bool,optional"`
	hostFilterRelabelConfigs []*relabel.Config      `river:"host_filter_relabel_configs,block,optional"`
	scrapeConfigs            []*config.ScrapeConfig `river:"scrape_configs,block,optional"`
}
