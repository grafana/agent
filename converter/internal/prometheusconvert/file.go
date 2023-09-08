package prometheusconvert

import (
	"time"

	"github.com/grafana/agent/component/discovery"
	"github.com/grafana/agent/component/discovery/file"
	"github.com/grafana/agent/converter/diag"
	"github.com/grafana/agent/converter/internal/common"
	prom_file "github.com/prometheus/prometheus/discovery/file"
)

func appendDiscoveryFile(pb *prometheusBlocks, label string, sdConfig *prom_file.SDConfig) discovery.Exports {
	discoveryFileArgs := toDiscoveryFile(sdConfig)
	name := []string{"discovery", "file"}
	block := common.NewBlockWithOverride(name, label, discoveryFileArgs)
	pb.discoveryBlocks = append(pb.discoveryBlocks, newPrometheusBlock(block, name, label, "", ""))
	return NewDiscoveryExports("discovery.file." + label + ".targets")
}

func validateDiscoveryFile(sdConfig *prom_file.SDConfig) diag.Diagnostics {
	return make(diag.Diagnostics, 0)
}

func toDiscoveryFile(sdConfig *prom_file.SDConfig) *file.Arguments {
	if sdConfig == nil {
		return nil
	}

	return &file.Arguments{
		Files:           sdConfig.Files,
		RefreshInterval: time.Duration(sdConfig.RefreshInterval),
	}
}
