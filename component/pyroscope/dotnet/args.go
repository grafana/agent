package dotnet

import (
	"time"

	"github.com/grafana/agent/component/discovery"
)

type Arguments struct {
	Targets []discovery.Target `river:"targets,attr"`

	TmpDir          string          `river:"tmp_dir,attr,optional"`
	ProfilingConfig ProfilingConfig `river:"profiling_config,block,optional"`
}

type ProfilingConfig struct {
	Interval time.Duration `river:"interval,attr,optional"`
}

func (rc *Arguments) UnmarshalRiver(f func(interface{}) error) error {
	*rc = defaultArguments()
	type config Arguments
	return f((*config)(rc))
}

func defaultArguments() Arguments {
	return Arguments{TmpDir: "/tmp", ProfilingConfig: ProfilingConfig{
		Interval: 60 * time.Second,
	},
	}
}
