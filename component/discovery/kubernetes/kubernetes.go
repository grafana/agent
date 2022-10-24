package kubernetes

import (
	"reflect"

	"github.com/grafana/agent/component"
	"github.com/grafana/agent/component/common/config"
	"github.com/grafana/agent/component/discovery"
	promk8s "github.com/prometheus/prometheus/discovery/kubernetes"
)

func init() {
	component.Register(component.Registration{
		Name:    "discovery.kubernetes",
		Args:    SDConfig{},
		Exports: discovery.Exports{},

		Build: func(opts component.Options, args component.Arguments) (component.Component, error) {
			return discovery.New(opts, args, newK8s)
		},
		Type: reflect.TypeOf(discovery.Component{}),
	})
}

var newK8s discovery.Creator = func(args component.Arguments, opts component.Options) (discovery.Discoverer, error) {
	newArgs := args.(SDConfig)
	return promk8s.New(opts.Logger, newArgs.Convert())
}

// SDConfig is a conversion of discover/kubernetes/SDConfig to be compatible with flow
type SDConfig struct {
	APIServer          config.URL              `river:"api_server,attr,optional"`
	Role               string                  `river:"role,attr"`
	KubeConfig         string                  `river:"kubeconfig_file,attr,optional"`
	HTTPClientConfig   config.HTTPClientConfig `river:"http_client_config,block,optional"`
	NamespaceDiscovery NamespaceDiscovery      `river:"namespaces,block,optional"`
	Selectors          []SelectorConfig        `river:"selectors,block,optional"`
}

// DefaultConfig holds defaults for SDConfig. (copied from prometheus)
var DefaultConfig = SDConfig{
	HTTPClientConfig: config.DefaultHTTPClientConfig,
}

// UnmarshalRiver simply applies defaults then unmarshals regularly
func (sd *SDConfig) UnmarshalRiver(f func(interface{}) error) error {
	*sd = DefaultConfig
	type arguments SDConfig
	return f((*arguments)(sd))
}

// Convert to prometheus config type
func (sd *SDConfig) Convert() *promk8s.SDConfig {
	selectors := make([]promk8s.SelectorConfig, len(sd.Selectors))
	for i, s := range sd.Selectors {
		selectors[i] = *s.convert()
	}
	return &promk8s.SDConfig{
		APIServer:          sd.APIServer.Convert(),
		Role:               promk8s.Role(sd.Role),
		KubeConfig:         sd.KubeConfig,
		HTTPClientConfig:   *sd.HTTPClientConfig.Convert(),
		NamespaceDiscovery: *sd.NamespaceDiscovery.convert(),
		Selectors:          selectors,
	}
}

// NamespaceDiscovery mirroring prometheus type
type NamespaceDiscovery struct {
	IncludeOwnNamespace bool     `river:"own_namespace,attr,optional"`
	Names               []string `river:"names,attr,optional"`
}

func (nd *NamespaceDiscovery) convert() *promk8s.NamespaceDiscovery {
	return &promk8s.NamespaceDiscovery{
		IncludeOwnNamespace: nd.IncludeOwnNamespace,
		Names:               nd.Names,
	}
}

// SelectorConfig mirroring prometheus type
type SelectorConfig struct {
	Role  string `river:"role,attr"`
	Label string `river:"label,attr,optional"`
	Field string `river:"field,attr,optional"`
}

func (sc *SelectorConfig) convert() *promk8s.SelectorConfig {
	return &promk8s.SelectorConfig{
		Role:  promk8s.Role(sc.Role),
		Label: sc.Label,
		Field: sc.Field,
	}
}
