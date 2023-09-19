package metrics

import (
	"time"

	"github.com/grafana/agent/component/common/loki/client"
	"github.com/grafana/agent/component/common/metrics/cluster"
	"github.com/grafana/agent/component/common/metrics/instance"
)

// type Config struct {
// 	Global                 instance.GlobalConfig `yaml:"global,omitempty"`
// 	WALDir                 string                `yaml:"wal_directory,omitempty"`
// 	WALCleanupAge          time.Duration         `yaml:"wal_cleanup_age,omitempty"`
// 	WALCleanupPeriod       time.Duration         `yaml:"wal_cleanup_period,omitempty"`
// 	ServiceConfig          cluster.Config        `yaml:"scraping_service,omitempty"`
// 	ServiceClientConfig    client.Config         `yaml:"scraping_service_client,omitempty"`
// 	Configs                []instance.Config     `yaml:"configs,omitempty"`
// 	InstanceRestartBackoff time.Duration         `yaml:"instance_restart_backoff,omitempty"`
// 	InstanceMode           instance.Mode         `yaml:"instance_mode,omitempty"`
// 	DisableKeepAlives      bool                  `yaml:"http_disable_keepalives,omitempty"`
// 	IdleConnTimeout        time.Duration         `yaml:"http_idle_conn_timeout,omitempty"`
//
// 	// Unmarshaled is true when the Config was unmarshaled from YAML.
// 	Unmarshaled bool `yaml:"-"`
// }
// This config is DTO for above Config
type Config struct {
	global              instance.GlobalConfig `river:"global,block,optional"`
	walDir              string                `river:"wal_directory,string,optional"`
	walCleanupAge       time.Duration         `river:"wal_cleanup_age,attr,optional"`
	walCleanupPeriod    time.Duration         `river:"wal_cleanup_period,attr,optional"`
	serviceConfig       cluster.Config        `river:"scraping_service,block,optional"`
	serviceClientConfig client.Config         `river:"scraping_service_client,block,optional"`
	configs             []instance.Config     `river:"configs,block,optional"`
}
