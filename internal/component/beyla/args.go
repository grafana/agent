package beyla

import (
	"github.com/grafana/agent/internal/component/discovery"
	"github.com/grafana/agent/internal/component/otelcol"
)

// Arguments configures the Beyla component.
type Arguments struct {
	Port           string                     `river:"open_port,attr,optional"`
	ExecutableName string                     `river:"executable_name,attr,optional"`
	Routes         Routes                     `river:"routes,block,optional"`
	Attributes     Attributes                 `river:"attributes,block,optional"`
	Discovery      Discovery                  `river:"discovery,block,optional"`
	Output         *otelcol.ConsumerArguments `river:"output,block,optional"`
}

type Exports struct {
	Targets []discovery.Target `river:"targets,attr"`
}

type Routes struct {
	Unmatch        string   `river:"unmatched,attr,optional"`
	Patterns       []string `river:"patterns,attr,optional"`
	IgnorePatterns []string `river:"ignored_patterns,attr,optional"`
	IgnoredEvents  string   `river:"ignore_mode,attr,optional"`
}

type Attributes struct {
	Kubernetes KubernetesDecorator `river:"kubernetes,block"`
}

type KubernetesDecorator struct {
	Enable string `river:"enable,attr"`
}

type Services []Service

type Service struct {
	Name      string `river:"name,attr,optional"`
	Namespace string `river:"namespace,attr,optional"`
	OpenPorts string `river:"open_ports,attr,optional"`
	Path      string `river:"exe_path,attr,optional"`
}

type Discovery struct {
	Services Services `river:"services,block"`
}
