package prometheusconvert

import (
	"github.com/grafana/agent/component/common/config"
	"github.com/grafana/agent/component/discovery"
	"github.com/grafana/agent/component/discovery/kubernetes"
	"github.com/grafana/agent/converter/diag"
	"github.com/grafana/agent/converter/internal/common"
	prom_kubernetes "github.com/prometheus/prometheus/discovery/kubernetes"
)

func appendDiscoveryKubernetes(pb *prometheusBlocks, label string, sdConfig *prom_kubernetes.SDConfig) (discovery.Exports, diag.Diagnostics) {
	discoveryKubernetesArgs, diags := toDiscoveryKubernetes(sdConfig)
	block := common.NewBlockWithOverride([]string{"discovery", "kubernetes"}, label, discoveryKubernetesArgs)
	pb.discoveryBlocks = append(pb.discoveryBlocks, block)
	return newDiscoverExports("discovery.kubernetes." + label + ".targets"), diags
}

func toDiscoveryKubernetes(sdConfig *prom_kubernetes.SDConfig) (*kubernetes.Arguments, diag.Diagnostics) {
	if sdConfig == nil {
		return nil, nil
	}

	return &kubernetes.Arguments{
		APIServer:          config.URL(sdConfig.APIServer),
		Role:               string(sdConfig.Role),
		KubeConfig:         sdConfig.KubeConfig,
		HTTPClientConfig:   *toHttpClientConfig(&sdConfig.HTTPClientConfig),
		NamespaceDiscovery: toNamespaceDiscovery(&sdConfig.NamespaceDiscovery),
		Selectors:          toSelectorConfig(sdConfig.Selectors),
		AttachMetadata:     toAttachMetadata(&sdConfig.AttachMetadata),
	}, validateDiscoveryKubernetes(sdConfig)
}

func validateDiscoveryKubernetes(sdConfig *prom_kubernetes.SDConfig) diag.Diagnostics {
	return validateHttpClientConfig(&sdConfig.HTTPClientConfig)
}

func toNamespaceDiscovery(ndConfig *prom_kubernetes.NamespaceDiscovery) kubernetes.NamespaceDiscovery {
	return kubernetes.NamespaceDiscovery{
		IncludeOwnNamespace: ndConfig.IncludeOwnNamespace,
		Names:               ndConfig.Names,
	}
}

func toSelectorConfig(selectors []prom_kubernetes.SelectorConfig) []kubernetes.SelectorConfig {
	var result []kubernetes.SelectorConfig

	for _, selector := range selectors {
		result = append(result, kubernetes.SelectorConfig{
			Role:  string(selector.Role),
			Label: selector.Label,
			Field: selector.Field,
		})
	}

	return result
}

func toAttachMetadata(amConfig *prom_kubernetes.AttachMetadataConfig) kubernetes.AttachMetadataConfig {
	return kubernetes.AttachMetadataConfig{
		Node: amConfig.Node,
	}
}
