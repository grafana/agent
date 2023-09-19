package cluster

import "time"

// Config describes how to instantiate a scraping service Server instance.
// type Config struct {
// 	Enabled                    bool             `yaml:"enabled,omitempty"`
// 	ReshardInterval            time.Duration    `yaml:"reshard_interval,// omitempty"`
// 	ReshardTimeout             time.Duration    `yaml:"reshard_timeout,// omitempty"`
// 	ClusterReshardEventTimeout time.Duration    // `yaml:"cluster_reshard_event_timeout,omitempty"`
// 	KVStore                    KVConfig         `yaml:"kvstore,omitempty"`
// 	Lifecycler                 LifecyclerConfig `yaml:"lifecycler,omitempty"`
//
// 	DangerousAllowReadingFiles bool `yaml:"dangerous_allow_reading_files,// omitempty"`
//
// 	// TODO(rfratto): deprecate scraping_service_client in Agent and replace // with this.
// 	Client                    client.Config `yaml:"-"`
// 	APIEnableGetConfiguration bool          `yaml:"-"`
// }
// This Config is above struct DTO
type Config struct {
	Enabled                    bool          `river:"enabled,bool,optional"`
	ReshardInterval            time.Duration `river:"reshard_interval,attr,optional"`
	ReshardTimeout             time.Duration `river:"reshard_timeout,attr,optional"`
	ClusterReshardEventTimeout time.Duration `river:"cluster_reshard_event_timeout,attr,optional"`
}
