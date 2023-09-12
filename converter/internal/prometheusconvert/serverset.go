package prometheusconvert

import (
	"time"

	"github.com/grafana/agent/component/discovery"
	"github.com/grafana/agent/component/discovery/serverset"
	"github.com/grafana/agent/converter/diag"
	"github.com/grafana/agent/converter/internal/common"
	prom_zk "github.com/prometheus/prometheus/discovery/zookeeper"
)

func appendDiscoveryServerset(pb *prometheusBlocks, label string, sdc *prom_zk.ServersetSDConfig) discovery.Exports {
	discoveryServersetArgs := ToDiscoveryServerset(sdc)
	name := []string{"discovery", "serverset"}
	block := common.NewBlockWithOverride(name, label, discoveryServersetArgs)
	pb.discoveryBlocks = append(pb.discoveryBlocks, newPrometheusBlock(block, name, label, "", ""))
	return NewDiscoveryExports("discovery.serverset." + label + ".targets")
}

func ToDiscoveryServerset(sdc *prom_zk.ServersetSDConfig) *serverset.Arguments {
	if sdc == nil {
		return nil
	}

	return &serverset.Arguments{
		Servers: sdc.Servers,
		Paths:   sdc.Paths,
		Timeout: time.Duration(sdc.Timeout),
	}
}

func validateDiscoveryServerset(_ *prom_zk.ServersetSDConfig) diag.Diagnostics {
	return nil
}
