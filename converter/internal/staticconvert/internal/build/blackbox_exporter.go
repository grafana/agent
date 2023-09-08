package build

import (
	"fmt"
	"time"

	"github.com/grafana/agent/component/discovery"
	"github.com/grafana/agent/component/prometheus/exporter/blackbox"
	"github.com/grafana/agent/converter/internal/common"
	"github.com/grafana/agent/converter/internal/prometheusconvert"
	"github.com/grafana/agent/pkg/integrations/blackbox_exporter"
	"github.com/grafana/river/rivertypes"
)

func (b *IntegrationsV1ConfigBuilder) appendBlackboxExporter(config *blackbox_exporter.Config) discovery.Exports {
	args := toBlackboxExporter(config)
	compLabel := common.LabelForParts(b.globalCtx.LabelPrefix, config.Name())
	b.f.Body().AppendBlock(common.NewBlockWithOverride(
		[]string{"prometheus", "exporter", "blackbox"},
		compLabel,
		args,
	))

	return prometheusconvert.NewDiscoveryExports(fmt.Sprintf("prometheus.exporter.blackbox.%s.targets", compLabel))
}

func toBlackboxExporter(config *blackbox_exporter.Config) *blackbox.Arguments {
	return &blackbox.Arguments{
		ConfigFile:         config.BlackboxConfigFile,
		Config:             rivertypes.OptionalSecret{},
		Targets:            toBlackboxTargets(config.BlackboxTargets),
		ProbeTimeoutOffset: time.Duration(config.ProbeTimeoutOffset),
		ConfigStruct:       config.BlackboxConfig,
	}
}

func toBlackboxTargets(blackboxTargets []blackbox_exporter.BlackboxTarget) blackbox.TargetBlock {
	var targetBlock blackbox.TargetBlock

	for _, bt := range blackboxTargets {
		targetBlock = append(targetBlock, toBlackboxTarget(bt))
	}

	return targetBlock
}

func toBlackboxTarget(target blackbox_exporter.BlackboxTarget) blackbox.BlackboxTarget {
	return blackbox.BlackboxTarget{
		Name:   common.SanitizeIdentifierPanics(target.Name),
		Target: target.Target,
		Module: target.Module,
	}
}
