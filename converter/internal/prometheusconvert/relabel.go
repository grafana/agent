package prometheusconvert

import (
	"fmt"

	flow_relabel "github.com/grafana/agent/component/common/relabel"
	"github.com/grafana/agent/component/discovery"
	disc_relabel "github.com/grafana/agent/component/discovery/relabel"
	"github.com/grafana/agent/component/prometheus/relabel"
	"github.com/grafana/agent/converter/internal/common"
	"github.com/grafana/agent/pkg/river/token/builder"
	promrelabel "github.com/prometheus/prometheus/model/relabel"
	"github.com/prometheus/prometheus/storage"
)

func appendPrometheusRelabel(f *builder.File, relabelConfigs []*promrelabel.Config, forwardTo []storage.Appendable, label string) *relabel.Exports {
	if len(relabelConfigs) == 0 {
		return nil
	}

	relabelArgs := toRelabelArguments(relabelConfigs, forwardTo)
	common.AppendBlockWithOverride(f, []string{"prometheus", "relabel"}, label, relabelArgs)

	return &relabel.Exports{
		Receiver: common.ConvertAppendable{Expr: fmt.Sprintf("prometheus.relabel.%s.receiver", label)},
	}
}

func toRelabelArguments(relabelConfigs []*promrelabel.Config, forwardTo []storage.Appendable) *relabel.Arguments {
	if len(relabelConfigs) == 0 {
		return nil
	}

	return &relabel.Arguments{
		ForwardTo:            forwardTo,
		MetricRelabelConfigs: toRelabelConfigs(relabelConfigs),
	}
}

func appendDiscoveryRelabel(f *builder.File, relabelConfigs []*promrelabel.Config, label string, targets []discovery.Target) *disc_relabel.Exports {
	if len(relabelConfigs) == 0 {
		return nil
	}

	relabelArgs := toDiscoveryRelabelArguments(relabelConfigs, targets)
	common.AppendBlockWithOverride(f, []string{"discovery", "relabel"}, label, relabelArgs)

	return &disc_relabel.Exports{
		Output: newDiscoveryTargets(fmt.Sprintf("discovery.relabel.%s.targets", label)),
	}
}

func toDiscoveryRelabelArguments(relabelConfigs []*promrelabel.Config, targets []discovery.Target) *disc_relabel.Arguments {
	return &disc_relabel.Arguments{
		Targets:        targets,
		RelabelConfigs: toRelabelConfigs(relabelConfigs),
	}
}

func toRelabelConfigs(relabelConfigs []*promrelabel.Config) []*flow_relabel.Config {
	if len(relabelConfigs) == 0 {
		return nil
	}

	var metricRelabelConfigs []*flow_relabel.Config
	for _, relabelConfig := range relabelConfigs {
		sourceLabels := make([]string, len(relabelConfig.SourceLabels))
		for i, sourceLabel := range relabelConfig.SourceLabels {
			sourceLabels[i] = string(sourceLabel)
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
