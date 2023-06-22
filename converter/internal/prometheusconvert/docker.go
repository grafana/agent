package prometheusconvert

import (
	"time"

	"github.com/grafana/agent/component/discovery"
	"github.com/grafana/agent/component/discovery/docker"
	"github.com/grafana/agent/converter/diag"
	"github.com/grafana/agent/converter/internal/common"
	"github.com/grafana/agent/pkg/river/token/builder"
	promdocker "github.com/prometheus/prometheus/discovery/moby"
)

func appendDiscoveryDocker(f *builder.File, label string, sdConfig *promdocker.DockerSDConfig) (discovery.Exports, diag.Diagnostics) {
	discoveryDockerArgs, diags := toDiscoveryDocker(sdConfig)
	common.AppendBlockWithOverride(f, []string{"discovery", "docker"}, label, discoveryDockerArgs)
	return newDiscoverExports("discovery.docker." + label + ".targets"), diags
}

func toDiscoveryDocker(sdConfig *promdocker.DockerSDConfig) (*docker.Arguments, diag.Diagnostics) {
	if sdConfig == nil {
		return nil, nil
	}

	return &docker.Arguments{
		Host:               sdConfig.Host,
		Port:               sdConfig.Port,
		HostNetworkingHost: sdConfig.HostNetworkingHost,
		RefreshInterval:    time.Duration(sdConfig.RefreshInterval),
		Filters:            toDockerFilters(sdConfig.Filters),
		HTTPClientConfig:   *toHttpClientConfig(&sdConfig.HTTPClientConfig),
	}, validateDiscoveryDocker(sdConfig)
}

func validateDiscoveryDocker(sdConfig *promdocker.DockerSDConfig) diag.Diagnostics {
	return validateHttpClientConfig(&sdConfig.HTTPClientConfig)
}

func toDockerFilters(filtersConfig []promdocker.Filter) []docker.Filter {
	filters := make([]docker.Filter, 0)

	for _, filter := range filtersConfig {
		filters = append(filters, docker.Filter{
			Name:   filter.Name,
			Values: filter.Values,
		})
	}

	return filters
}
