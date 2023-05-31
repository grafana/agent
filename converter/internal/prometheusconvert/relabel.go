package prometheusconvert

import (
	"fmt"

	flow_relabel "github.com/grafana/agent/component/common/relabel"
	"github.com/grafana/agent/component/prometheus/relabel"
	"github.com/grafana/agent/converter/internal/common"
	"github.com/grafana/agent/pkg/river/token/builder"
	promrelabel "github.com/prometheus/prometheus/model/relabel"
	"github.com/prometheus/prometheus/storage"
)

func appendRelabel(f *builder.File, relabelConfigs []*promrelabel.Config, forwardTo []storage.Appendable, label string) *relabel.Exports {
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

	return &relabel.Arguments{
		ForwardTo:            forwardTo,
		MetricRelabelConfigs: metricRelabelConfigs,
	}
}
