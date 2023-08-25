package build

import (
	"fmt"

	"github.com/grafana/agent/component/discovery"
	"github.com/grafana/agent/component/prometheus/exporter/snmp"
	"github.com/grafana/agent/converter/internal/common"
	"github.com/grafana/agent/converter/internal/prometheusconvert"
	"github.com/grafana/agent/pkg/integrations/snmp_exporter"
	"github.com/grafana/river/rivertypes"
	snmp_config "github.com/prometheus/snmp_exporter/config"
)

func (b *IntegrationsV1ConfigBuilder) appendSnmpExporter(config *snmp_exporter.Config) discovery.Exports {
	args := toSnmpExporter(config)
	compLabel := common.LabelForParts(b.globalCtx.LabelPrefix, config.Name())
	b.f.Body().AppendBlock(common.NewBlockWithOverride(
		[]string{"prometheus", "exporter", "snmp"},
		compLabel,
		args,
	))

	return prometheusconvert.NewDiscoveryExports(fmt.Sprintf("prometheus.exporter.snmp.%s.targets", compLabel))
}

func toSnmpExporter(config *snmp_exporter.Config) *snmp.Arguments {
	targets := make([]snmp.SNMPTarget, len(config.SnmpTargets))
	for i, t := range config.SnmpTargets {
		targets[i] = snmp.SNMPTarget{
			Name:       t.Name,
			Target:     t.Target,
			Module:     t.Module,
			Auth:       t.Auth,
			WalkParams: t.WalkParams,
		}
	}

	walkParams := make([]snmp.WalkParam, len(config.WalkParams))
	index := 0
	for name, p := range config.WalkParams {
		walkParams[index] = snmp.WalkParam{
			Name:                    name,
			MaxRepetitions:          p.MaxRepetitions,
			Retries:                 *p.Retries,
			Timeout:                 p.Timeout,
			UseUnconnectedUDPSocket: p.UseUnconnectedUDPSocket,
		}
		index++
	}

	return &snmp.Arguments{
		ConfigFile: config.SnmpConfigFile,
		Config:     rivertypes.OptionalSecret{},
		Targets:    targets,
		WalkParams: walkParams,
		ConfigStruct: snmp_config.Config{
			Auths:   config.SnmpConfig.Auths,
			Modules: config.SnmpConfig.Modules,
			Version: config.SnmpConfig.Version,
		},
	}
}
