package build

import (
	"fmt"

	"github.com/grafana/agent/component/discovery"
	"github.com/grafana/agent/component/prometheus/exporter/process"
	"github.com/grafana/agent/converter/internal/common"
	"github.com/grafana/agent/converter/internal/prometheusconvert"
	"github.com/grafana/agent/pkg/integrations/process_exporter"
)

func (b *IntegrationsV1ConfigBuilder) appendProcessExporter(config *process_exporter.Config) discovery.Exports {
	args := toProcessExporter(config)
	compLabel := common.LabelForParts(b.globalCtx.LabelPrefix, config.Name())
	b.f.Body().AppendBlock(common.NewBlockWithOverride(
		[]string{"prometheus", "exporter", "process"},
		compLabel,
		args,
	))

	return prometheusconvert.NewDiscoveryExports(fmt.Sprintf("prometheus.exporter.process.%s.targets", compLabel))
}

func toProcessExporter(config *process_exporter.Config) *process.Arguments {
	matcherGroups := make([]process.MatcherGroup, 0)
	for _, matcherGroup := range config.ProcessExporter {
		matcherGroups = append(matcherGroups, process.MatcherGroup{
			Name:         matcherGroup.Name,
			CommRules:    matcherGroup.CommRules,
			ExeRules:     matcherGroup.ExeRules,
			CmdlineRules: matcherGroup.CmdlineRules,
		})
	}

	return &process.Arguments{
		ProcessExporter: matcherGroups,
		ProcFSPath:      config.ProcFSPath,
		Children:        config.Children,
		Threads:         config.Threads,
		SMaps:           config.SMaps,
		Recheck:         config.Recheck,
	}
}
