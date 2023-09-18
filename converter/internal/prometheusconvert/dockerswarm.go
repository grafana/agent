package prometheusconvert

import (
	"time"

	"github.com/grafana/agent/component/discovery"
	"github.com/grafana/agent/component/discovery/dockerswarm"
	"github.com/grafana/agent/converter/diag"
	"github.com/grafana/agent/converter/internal/common"
	prom_moby "github.com/prometheus/prometheus/discovery/moby"
)

func appendDiscoveryDockerswarm(pb *prometheusBlocks, label string, sdConfig *prom_moby.DockerSwarmSDConfig) discovery.Exports {
	discoveryDockerswarmArgs := toDiscoveryDockerswarm(sdConfig)
	name := []string{"discovery", "dockerswarm"}
	block := common.NewBlockWithOverride(name, label, discoveryDockerswarmArgs)
	pb.discoveryBlocks = append(pb.discoveryBlocks, newPrometheusBlock(block, name, label, "", ""))
	return NewDiscoveryExports("discovery.dockerswarm." + label + ".targets")
}

func validateDiscoveryDockerswarm(sdConfig *prom_moby.DockerSwarmSDConfig) diag.Diagnostics {
	return ValidateHttpClientConfig(&sdConfig.HTTPClientConfig)
}

func toDiscoveryDockerswarm(sdConfig *prom_moby.DockerSwarmSDConfig) *dockerswarm.Arguments {
	if sdConfig == nil {
		return nil
	}

	return &dockerswarm.Arguments{
		Host:             sdConfig.Host,
		Role:             sdConfig.Role,
		Port:             sdConfig.Port,
		Filters:          convertFilters(sdConfig.Filters),
		RefreshInterval:  time.Duration(sdConfig.RefreshInterval),
		HTTPClientConfig: *ToHttpClientConfig(&sdConfig.HTTPClientConfig),
	}
}

func convertFilters(mobyFilters []prom_moby.Filter) []dockerswarm.Filter {
	riverFilters := make([]dockerswarm.Filter, len(mobyFilters))
	for i, mobyFilter := range mobyFilters {
		riverFilters[i] = convertFilter(&mobyFilter)
	}
	return riverFilters
}

func convertFilter(mobyFilter *prom_moby.Filter) dockerswarm.Filter {
	values := make([]string, len(mobyFilter.Values))
	copy(values, mobyFilter.Values)

	return dockerswarm.Filter{
		Name:   mobyFilter.Name,
		Values: values,
	}
}
