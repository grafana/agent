package ebpf

import (
	"time"

	"github.com/grafana/agent/component/discovery"
	"github.com/grafana/agent/component/pyroscope"
)

type Arguments struct {
	ForwardTo            []pyroscope.Appendable `river:"forward_to,attr"`
	Targets              []discovery.Target     `river:"targets,attr,optional"`
	DefaultTarget        discovery.Target       `river:"default_target,attr,optional"` // undocumented, keeping it until we have other sd
	TargetsOnly          bool                   `river:"targets_only,attr,optional"`   // undocumented, keeping it until we have other sd
	CollectInterval      time.Duration          `river:"collect_interval,attr,optional"`
	SampleRate           int                    `river:"sample_rate,attr,optional"`
	PidCacheSize         int                    `river:"pid_cache_size,attr,optional"`
	BuildIDCacheSize     int                    `river:"build_id_cache_size,attr,optional"`
	SameFileCacheSize    int                    `river:"same_file_cache_size,attr,optional"`
	ContainerIDCacheSize int                    `river:"container_id_cache_size,attr,optional"`
	CacheRounds          int                    `river:"cache_rounds,attr,optional"`
	CollectUserProfile   bool                   `river:"collect_user_profile,attr,optional"`
	CollectKernelProfile bool                   `river:"collect_kernel_profile,attr,optional"`
	Demangle             string                 `river:"demangle,attr,optional"`
}
