package ebpf

import (
	"errors"
	"time"

	"github.com/grafana/agent/internal/component/discovery"
	"github.com/grafana/agent/internal/component/pyroscope"
)

type Arguments struct {
	ForwardTo            []pyroscope.Appendable `river:"forward_to,attr"`
	Targets              []discovery.Target     `river:"targets,attr,optional"`
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
	PythonEnabled        bool                   `river:"python_enabled,attr,optional"`
	SymbolsMapSize       int                    `river:"symbols_map_size,attr,optional"`
	PIDMapSize           int                    `river:"pid_map_size,attr,optional"`
}

// Validate implements syntax.Validator.
func (arg *Arguments) Validate() error {
	var errs []error
	if arg.SymbolsMapSize <= 0 {
		errs = append(errs, errors.New("symbols_map_size must be greater than 0"))
	}
	if arg.PIDMapSize <= 0 {
		errs = append(errs, errors.New("pid_map_size must be greater than 0"))
	}
	return errors.Join(errs...)
}
