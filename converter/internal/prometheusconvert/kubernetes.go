package prometheusconvert

import (
	"github.com/grafana/agent/component/common/config"
	"github.com/grafana/agent/component/discovery"
	"github.com/grafana/agent/component/discovery/kubernetes"
	"github.com/grafana/agent/converter/diag"
	"github.com/grafana/agent/converter/internal/common"
	"github.com/grafana/agent/pkg/river/token/builder"
	promkubernetes "github.com/prometheus/prometheus/discovery/kubernetes"
)

func appendDiscoveryKubernetes(f *builder.File, label string, sdConfig *promkubernetes.SDConfig) (discovery.Exports, diag.Diagnostics) {
	discoveryKubernetesArgs, diags := toDiscoveryKubernetes(sdConfig)
	common.AppendBlockWithOverride(f, []string{"discovery", "kubernetes"}, label, discoveryKubernetesArgs)
	return newDiscoverExports("discovery.kubernetes." + label + ".targets"), diags
}

func toDiscoveryKubernetes(sdConfig *promkubernetes.SDConfig) (*kubernetes.Arguments, diag.Diagnostics) {
	if sdConfig == nil {
		return nil, nil
	}

	return &kubernetes.Arguments{
		APIServer:          config.URL(sdConfig.APIServer),
		Role:               string(sdConfig.Role),
		KubeConfig:         sdConfig.KubeConfig,
		HTTPClientConfig:   *toHttpClientConfig(&sdConfig.HTTPClientConfig),
		NamespaceDiscovery: *toNamespaceDiscovery(&sdConfig.NamespaceDiscovery),
		Selectors:          *toSelectorConfig(&sdConfig.Selectors),
		AttachMetadata:     *toAttachMetadata(&sdConfig.AttachMetadata),
	}, validateDiscoveryKubernetes(sdConfig)
}

func validateDiscoveryKubernetes(sdConfig *promkubernetes.SDConfig) diag.Diagnostics {
	return validateHttpClientConfig(&sdConfig.HTTPClientConfig)
}

func toNamespaceDiscovery(ndConfig *promkubernetes.NamespaceDiscovery) *kubernetes.NamespaceDiscovery {
	return &kubernetes.NamespaceDiscovery{
		IncludeOwnNamespace: ndConfig.IncludeOwnNamespace,
		Names:               ndConfig.Names,
	}
}

func toSelectorConfig(selectors *[]promkubernetes.SelectorConfig) *[]kubernetes.SelectorConfig {
	var result []kubernetes.SelectorConfig

	for _, selector := range *selectors {
		result = append(result, kubernetes.SelectorConfig{
			Role:  string(selector.Role),
			Label: selector.Label,
			Field: selector.Field,
		})
	}

	return &result
}

func toAttachMetadata(amConfig *promkubernetes.AttachMetadataConfig) *kubernetes.AttachMetadataConfig {
	return &kubernetes.AttachMetadataConfig{
		Node: amConfig.Node,
	}
}
