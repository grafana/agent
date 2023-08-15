package prometheusconvert

import (
	"time"

	"github.com/grafana/agent/component/discovery"
	"github.com/grafana/agent/component/discovery/docker"
	"github.com/grafana/agent/converter/diag"
	"github.com/grafana/agent/converter/internal/common"
	prom_docker "github.com/prometheus/prometheus/discovery/moby"
)

func appendDiscoveryDocker(pb *prometheusBlocks, label string, sdConfig *prom_docker.DockerSDConfig) discovery.Exports {
	discoveryDockerArgs := ToDiscoveryDocker(sdConfig)
	name := []string{"discovery", "docker"}
	block := common.NewBlockWithOverride(name, label, discoveryDockerArgs)
	pb.discoveryBlocks = append(pb.discoveryBlocks, newPrometheusBlock(block, name, label, "", ""))
	return NewDiscoverExports("discovery.docker." + label + ".targets")
}

func validateDiscoveryDocker(sdConfig *prom_docker.DockerSDConfig) diag.Diagnostics {
	return ValidateHttpClientConfig(&sdConfig.HTTPClientConfig)
}

func ToDiscoveryDocker(sdConfig *prom_docker.DockerSDConfig) *docker.Arguments {
	if sdConfig == nil {
		return nil
	}

	return &docker.Arguments{
		Host:               sdConfig.Host,
		Port:               sdConfig.Port,
		HostNetworkingHost: sdConfig.HostNetworkingHost,
		RefreshInterval:    time.Duration(sdConfig.RefreshInterval),
		Filters:            toDockerFilters(sdConfig.Filters),
		HTTPClientConfig:   *ToHttpClientConfig(&sdConfig.HTTPClientConfig),
	}
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
