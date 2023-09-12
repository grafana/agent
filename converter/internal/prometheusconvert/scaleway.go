package prometheusconvert

import (
	"time"

	"github.com/grafana/agent/component/common/config"
	"github.com/grafana/agent/component/discovery"
	"github.com/grafana/agent/component/discovery/scaleway"
	"github.com/grafana/agent/converter/diag"
	"github.com/grafana/agent/converter/internal/common"
	"github.com/grafana/river/rivertypes"
	prom_scaleway "github.com/prometheus/prometheus/discovery/scaleway"
)

func appendDiscoveryScaleway(pb *prometheusBlocks, label string, sdConfig *prom_scaleway.SDConfig) discovery.Exports {
	discoveryScalewayArgs := ToDiscoveryScaleway(sdConfig)
	name := []string{"discovery", "scaleway"}
	block := common.NewBlockWithOverride(name, label, discoveryScalewayArgs)
	pb.discoveryBlocks = append(pb.discoveryBlocks, newPrometheusBlock(block, name, label, "", ""))
	return NewDiscoveryExports("discovery.scaleway." + label + ".targets")
}

func validateDiscoveryScaleway(sdConfig *prom_scaleway.SDConfig) diag.Diagnostics {
	return nil
}

func ToDiscoveryScaleway(sdConfig *prom_scaleway.SDConfig) *scaleway.Arguments {
	if sdConfig == nil {
		return nil
	}

	return &scaleway.Arguments{
		Project:         sdConfig.Project,
		Role:            scaleway.Role(sdConfig.Role),
		APIURL:          sdConfig.APIURL,
		Zone:            sdConfig.Zone,
		AccessKey:       sdConfig.AccessKey,
		SecretKey:       rivertypes.Secret(sdConfig.SecretKey),
		SecretKeyFile:   sdConfig.SecretKeyFile,
		NameFilter:      sdConfig.NameFilter,
		TagsFilter:      sdConfig.TagsFilter,
		RefreshInterval: time.Duration(sdConfig.RefreshInterval),
		Port:            sdConfig.Port,
		ProxyURL:        config.URL(sdConfig.HTTPClientConfig.ProxyURL),
		TLSConfig:       *ToTLSConfig(&sdConfig.HTTPClientConfig.TLSConfig),
		FollowRedirects: sdConfig.HTTPClientConfig.FollowRedirects,
		EnableHTTP2:     sdConfig.HTTPClientConfig.EnableHTTP2,
	}
}
