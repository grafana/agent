package prometheusconvert

import (
	"time"

	"github.com/grafana/agent/component/discovery"
	"github.com/grafana/agent/component/discovery/docker"
	"github.com/grafana/agent/converter/diag"
	"github.com/grafana/agent/converter/internal/common"
	"github.com/grafana/agent/pkg/river/token/builder"
	prom_docker "github.com/prometheus/prometheus/discovery/moby"
)

func appendDiscoveryDocker(f *builder.File, label string, sdConfig *prom_docker.DockerSDConfig) (discovery.Exports, diag.Diagnostics) {
	discoveryDockerArgs, diags := toDiscoveryDocker(sdConfig)
	common.AppendBlockWithOverride(f, []string{"discovery", "docker"}, label, discoveryDockerArgs)
	return newDiscoverExports("discovery.docker." + label + ".targets"), diags
}

func toDiscoveryDocker(sdConfig *prom_docker.DockerSDConfig) (*docker.Arguments, diag.Diagnostics) {
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

func validateDiscoveryDocker(sdConfig *prom_docker.DockerSDConfig) diag.Diagnostics {
	return validateHttpClientConfig(&sdConfig.HTTPClientConfig)
}

func toDockerFilters(filtersConfig []prom_docker.Filter) []docker.Filter {
	filters := make([]docker.Filter, 0)

	for _, filter := range filtersConfig {
		filters = append(filters, docker.Filter{
			Name:   filter.Name,
			Values: filter.Values,
		})
	}

	return filters
}
