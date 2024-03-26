package otelcolconvert

import (
	"fmt"

	"github.com/grafana/agent/internal/component/otelcol"
	"github.com/grafana/agent/internal/component/otelcol/receiver/vcenter"
	"github.com/grafana/agent/internal/converter/diag"
	"github.com/grafana/agent/internal/converter/internal/common"
	"github.com/grafana/river/rivertypes"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/vcenterreceiver"
	"go.opentelemetry.io/collector/component"
)

func init() {
	converters = append(converters, vcenterReceiverConverter{})
}

type vcenterReceiverConverter struct{}

func (vcenterReceiverConverter) Factory() component.Factory { return vcenterreceiver.NewFactory() }

func (vcenterReceiverConverter) InputComponentName() string { return "" }

func (vcenterReceiverConverter) ConvertAndAppend(state *State, id component.InstanceID, cfg component.Config) diag.Diagnostics {
	var diags diag.Diagnostics

	label := state.FlowComponentLabel()

	args := toVcenterReceiver(state, id, cfg.(*vcenterreceiver.Config))
	block := common.NewBlockWithOverride([]string{"otelcol", "receiver", "vcenter"}, label, args)

	diags.Add(
		diag.SeverityLevelInfo,
		fmt.Sprintf("Converted %s into %s", StringifyInstanceID(id), StringifyBlock(block)),
	)

	state.Body().AppendBlock(block)
	return diags
}

func toVcenterReceiver(state *State, id component.InstanceID, cfg *vcenterreceiver.Config) *vcenter.Arguments {
	var (
		nextMetrics = state.Next(id, component.DataTypeMetrics)
		nextTraces  = state.Next(id, component.DataTypeTraces)
	)

	return &vcenter.Arguments{
		Endpoint: cfg.Endpoint,
		Username: cfg.Username,
		Password: rivertypes.Secret(cfg.Password),

		DebugMetrics: common.DefaultValue[vcenter.Arguments]().DebugMetrics,

		MetricsBuilderConfig: toMetricsBuildConfig(encodeMapstruct(cfg.MetricsBuilderConfig)),

		ScraperControllerArguments: otelcol.ScraperControllerArguments{
			CollectionInterval: cfg.CollectionInterval,
			InitialDelay:       cfg.InitialDelay,
			Timeout:            cfg.Timeout,
		},

		TLS: toTLSClientArguments(cfg.TLSClientSetting),

		Output: &otelcol.ConsumerArguments{
			Metrics: ToTokenizedConsumers(nextMetrics),
			Traces:  ToTokenizedConsumers(nextTraces),
		},
	}
}

func toMetricsBuildConfig(cfg map[string]any) vcenter.MetricsBuilderConfig {
	return vcenter.MetricsBuilderConfig{
		Metrics:            toVcenterMetricsConfig(encodeMapstruct(cfg["metrics"])),
		ResourceAttributes: toVcenterResourceAttributesConfig(encodeMapstruct(cfg["resource_attributes"])),
	}
}

func toVcenterMetricConfig(cfg map[string]any) vcenter.MetricConfig {
	return vcenter.MetricConfig{
		Enabled: cfg["enabled"].(bool),
	}
}

func toVcenterMetricsConfig(cfg map[string]any) vcenter.MetricsConfig {
	return vcenter.MetricsConfig{
		VcenterClusterCPUEffective:      toVcenterMetricConfig(encodeMapstruct(cfg["vcenter.cluster.cpu.effective"])),
		VcenterClusterCPULimit:          toVcenterMetricConfig(encodeMapstruct(cfg["vcenter.cluster.cpu.limit"])),
		VcenterClusterHostCount:         toVcenterMetricConfig(encodeMapstruct(cfg["vcenter.cluster.host.count"])),
		VcenterClusterMemoryEffective:   toVcenterMetricConfig(encodeMapstruct(cfg["vcenter.cluster.memory.effective"])),
		VcenterClusterMemoryLimit:       toVcenterMetricConfig(encodeMapstruct(cfg["vcenter.cluster.memory.limit"])),
		VcenterClusterMemoryUsed:        toVcenterMetricConfig(encodeMapstruct(cfg["vcenter.cluster.memory.used"])),
		VcenterClusterVMCount:           toVcenterMetricConfig(encodeMapstruct(cfg["vcenter.cluster.vm.count"])),
		VcenterDatastoreDiskUsage:       toVcenterMetricConfig(encodeMapstruct(cfg["vcenter.datastore.disk.usage"])),
		VcenterDatastoreDiskUtilization: toVcenterMetricConfig(encodeMapstruct(cfg["vcenter.datastore.disk.utilization"])),
		VcenterHostCPUUsage:             toVcenterMetricConfig(encodeMapstruct(cfg["vcenter.host.cpu.usage"])),
		VcenterHostCPUUtilization:       toVcenterMetricConfig(encodeMapstruct(cfg["vcenter.host.cpu.utilization"])),
		VcenterHostDiskLatencyAvg:       toVcenterMetricConfig(encodeMapstruct(cfg["vcenter.host.disk.latency.avg"])),
		VcenterHostDiskLatencyMax:       toVcenterMetricConfig(encodeMapstruct(cfg["vcenter.host.disk.latency.max"])),
		VcenterHostDiskThroughput:       toVcenterMetricConfig(encodeMapstruct(cfg["vcenter.host.disk.throughput"])),
		VcenterHostMemoryUsage:          toVcenterMetricConfig(encodeMapstruct(cfg["vcenter.host.memory.usage"])),
		VcenterHostMemoryUtilization:    toVcenterMetricConfig(encodeMapstruct(cfg["vcenter.host.memory.utilization"])),
		VcenterHostNetworkPacketCount:   toVcenterMetricConfig(encodeMapstruct(cfg["vcenter.host.network.packet.count"])),
		VcenterHostNetworkPacketErrors:  toVcenterMetricConfig(encodeMapstruct(cfg["vcenter.host.network.packet.errors"])),
		VcenterHostNetworkThroughput:    toVcenterMetricConfig(encodeMapstruct(cfg["vcenter.host.network.throughput"])),
		VcenterHostNetworkUsage:         toVcenterMetricConfig(encodeMapstruct(cfg["vcenter.host.network.usage"])),
		VcenterResourcePoolCPUShares:    toVcenterMetricConfig(encodeMapstruct(cfg["vcenter.resource_pool.cpu.shares"])),
		VcenterResourcePoolCPUUsage:     toVcenterMetricConfig(encodeMapstruct(cfg["vcenter.resource_pool.cpu.usage"])),
		VcenterResourcePoolMemoryShares: toVcenterMetricConfig(encodeMapstruct(cfg["vcenter.resource_pool.memory.shares"])),
		VcenterResourcePoolMemoryUsage:  toVcenterMetricConfig(encodeMapstruct(cfg["vcenter.resource_pool.memory.usage"])),
		VcenterVMCPUUsage:               toVcenterMetricConfig(encodeMapstruct(cfg["vcenter.vm.cpu.usage"])),
		VcenterVMCPUUtilization:         toVcenterMetricConfig(encodeMapstruct(cfg["vcenter.vm.cpu.utilization"])),
		VcenterVMDiskLatencyAvg:         toVcenterMetricConfig(encodeMapstruct(cfg["vcenter.vm.disk.latency.avg"])),
		VcenterVMDiskLatencyMax:         toVcenterMetricConfig(encodeMapstruct(cfg["vcenter.vm.disk.latency.max"])),
		VcenterVMDiskThroughput:         toVcenterMetricConfig(encodeMapstruct(cfg["vcenter.vm.disk.throughput"])),
		VcenterVMDiskUsage:              toVcenterMetricConfig(encodeMapstruct(cfg["vcenter.vm.disk.usage"])),
		VcenterVMDiskUtilization:        toVcenterMetricConfig(encodeMapstruct(cfg["vcenter.vm.disk.utilization"])),
		VcenterVMMemoryBallooned:        toVcenterMetricConfig(encodeMapstruct(cfg["vcenter.vm.memory.ballooned"])),
		VcenterVMMemorySwapped:          toVcenterMetricConfig(encodeMapstruct(cfg["vcenter.vm.memory.swapped"])),
		VcenterVMMemorySwappedSsd:       toVcenterMetricConfig(encodeMapstruct(cfg["vcenter.vm.memory.swapped_ssd"])),
		VcenterVMMemoryUsage:            toVcenterMetricConfig(encodeMapstruct(cfg["vcenter.vm.memory.usage"])),
		VcenterVMMemoryUtilization:      toVcenterMetricConfig(encodeMapstruct(cfg["vcenter.vm.memory.utilization"])),
		VcenterVMNetworkPacketCount:     toVcenterMetricConfig(encodeMapstruct(cfg["vcenter.vm.network.packet.count"])),
		VcenterVMNetworkThroughput:      toVcenterMetricConfig(encodeMapstruct(cfg["vcenter.vm.network.throughput"])),
		VcenterVMNetworkUsage:           toVcenterMetricConfig(encodeMapstruct(cfg["vcenter.vm.network.usage"])),
	}
}

func toVcenterResourceAttributesConfig(cfg map[string]any) vcenter.ResourceAttributesConfig {
	return vcenter.ResourceAttributesConfig{
		VcenterClusterName:               toVcenterResourceAttributeConfig(encodeMapstruct(cfg["vcenter.cluster.name"])),
		VcenterDatastoreName:             toVcenterResourceAttributeConfig(encodeMapstruct(cfg["vcenter.datastore.name"])),
		VcenterHostName:                  toVcenterResourceAttributeConfig(encodeMapstruct(cfg["vcenter.host.name"])),
		VcenterResourcePoolInventoryPath: toVcenterResourceAttributeConfig(encodeMapstruct(cfg["vcenter.resource_pool.inventory_path"])),
		VcenterResourcePoolName:          toVcenterResourceAttributeConfig(encodeMapstruct(cfg["vcenter.resource_pool.name"])),
		VcenterVMID:                      toVcenterResourceAttributeConfig(encodeMapstruct(cfg["vcenter.vm.id"])),
		VcenterVMName:                    toVcenterResourceAttributeConfig(encodeMapstruct(cfg["vcenter.vm.name"])),
	}
}

func toVcenterResourceAttributeConfig(cfg map[string]any) vcenter.ResourceAttributeConfig {
	return vcenter.ResourceAttributeConfig{
		Enabled: cfg["enabled"].(bool),
	}
}
