package component

import (
	"fmt"

	flow_relabel "github.com/grafana/agent/component/common/relabel"
	"github.com/grafana/agent/component/discovery"
	disc_relabel "github.com/grafana/agent/component/discovery/relabel"
	"github.com/grafana/agent/component/prometheus/relabel"
	"github.com/grafana/agent/converter/internal/common"
	"github.com/grafana/agent/converter/internal/prometheusconvert/build"
	prom_relabel "github.com/prometheus/prometheus/model/relabel"
	"github.com/prometheus/prometheus/storage"
)

func AppendPrometheusRelabel(pb *build.PrometheusBlocks, relabelConfigs []*prom_relabel.Config, forwardTo []storage.Appendable, label string) *relabel.Exports {
	if len(relabelConfigs) == 0 {
		return nil
	}

	relabelArgs := toRelabelArguments(relabelConfigs, forwardTo)
	name := []string{"prometheus", "relabel"}
	block := common.NewBlockWithOverride(name, label, relabelArgs)
	pb.PrometheusRelabelBlocks = append(pb.PrometheusRelabelBlocks, build.NewPrometheusBlock(block, name, label, "", ""))

	return &relabel.Exports{
		Receiver: common.ConvertAppendable{Expr: fmt.Sprintf("prometheus.relabel.%s.receiver", label)},
	}
}

func toRelabelArguments(relabelConfigs []*prom_relabel.Config, forwardTo []storage.Appendable) *relabel.Arguments {
	if len(relabelConfigs) == 0 {
		return nil
	}

	return &relabel.Arguments{
		ForwardTo:            forwardTo,
		MetricRelabelConfigs: ToFlowRelabelConfigs(relabelConfigs),
		CacheSize:            100_000,
	}
}

func AppendDiscoveryRelabel(pb *build.PrometheusBlocks, relabelConfigs []*prom_relabel.Config, targets []discovery.Target, label string) *disc_relabel.Exports {
	if len(relabelConfigs) == 0 {
		return nil
	}

	relabelArgs := toDiscoveryRelabelArguments(relabelConfigs, targets)
	name := []string{"discovery", "relabel"}
	block := common.NewBlockWithOverride(name, label, relabelArgs)
	pb.DiscoveryRelabelBlocks = append(pb.DiscoveryRelabelBlocks, build.NewPrometheusBlock(block, name, label, "", ""))

	return &disc_relabel.Exports{
		Output: common.NewDiscoveryTargets(fmt.Sprintf("discovery.relabel.%s.output", label)),
	}
}

func toDiscoveryRelabelArguments(relabelConfigs []*prom_relabel.Config, targets []discovery.Target) *disc_relabel.Arguments {
	return &disc_relabel.Arguments{
		Targets:        targets,
		RelabelConfigs: ToFlowRelabelConfigs(relabelConfigs),
	}
}

func ToFlowRelabelConfigs(relabelConfigs []*prom_relabel.Config) []*flow_relabel.Config {
	if len(relabelConfigs) == 0 {
		return nil
	}

	var metricRelabelConfigs []*flow_relabel.Config
	for _, relabelConfig := range relabelConfigs {
		var sourceLabels []string
		if len(relabelConfig.SourceLabels) > 0 {
			sourceLabels = make([]string, len(relabelConfig.SourceLabels))
			for i, sourceLabel := range relabelConfig.SourceLabels {
				sourceLabels[i] = string(sourceLabel)
			}
		}

		metricRelabelConfigs = append(metricRelabelConfigs, &flow_relabel.Config{
			SourceLabels: sourceLabels,
			Separator:    relabelConfig.Separator,
			Regex:        flow_relabel.Regexp(relabelConfig.Regex),
			Modulus:      relabelConfig.Modulus,
			TargetLabel:  relabelConfig.TargetLabel,
			Replacement:  relabelConfig.Replacement,
			Action:       flow_relabel.Action(relabelConfig.Action),
		})
	}

	return metricRelabelConfigs
}
