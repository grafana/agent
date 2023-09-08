package prometheusconvert

import (
	"time"

	"github.com/grafana/agent/component/discovery"
	"github.com/grafana/agent/component/discovery/nerve"
	"github.com/grafana/agent/converter/diag"
	"github.com/grafana/agent/converter/internal/common"
	prom_nerve "github.com/prometheus/prometheus/discovery/zookeeper"
)

func appendDiscoveryNerve(pb *prometheusBlocks, label string, sdConfig *prom_nerve.NerveSDConfig) discovery.Exports {
	discoveryNerveArgs := toDiscoveryNerve(sdConfig)
	name := []string{"discovery", "nerve"}
	block := common.NewBlockWithOverride(name, label, discoveryNerveArgs)
	pb.discoveryBlocks = append(pb.discoveryBlocks, newPrometheusBlock(block, name, label, "", ""))
	return NewDiscoveryExports("discovery.nerve." + label + ".targets")
}

func validateDiscoveryNerve(sdConfig *prom_nerve.NerveSDConfig) diag.Diagnostics {
	return nil
}

func toDiscoveryNerve(sdConfig *prom_nerve.NerveSDConfig) *nerve.Arguments {
	if sdConfig == nil {
		return nil
	}

	return &nerve.Arguments{
		Servers: sdConfig.Servers,
		Paths:   sdConfig.Paths,
		Timeout: time.Duration(sdConfig.Timeout),
	}
}
