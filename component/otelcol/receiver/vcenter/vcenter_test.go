package vcenter

import (
	"testing"
	"time"

	"github.com/grafana/river"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/vcenterreceiver"
	"github.com/stretchr/testify/require"
)

func TestArguments_UnmarshalRiver(t *testing.T) {
	in := `
		endpoint = "http://localhost:1234"
		username = "user"
		password = "pass"
		collection_interval = "2m"

		resource_attributes {
			vcenter.cluster.name {
				enabled = true
			}
			vcenter.datastore.name {
				enabled = true
			}
			vcenter.host.name {
				enabled = true
			}
			vcenter.resource_pool.inventory_path {
				enabled = false
			}
			vcenter.resource_pool.name {
				enabled = true
			}
			vcenter.vm.name {
				enabled = true
			}
		}

		metrics {
			vcenter.cluster.cpu.effective {
				enabled = false
			}
			vcenter.cluster.cpu.limit {
				enabled = true
			}
			vcenter.cluster.host.count {
				enabled = true
			}
			vcenter.cluster.memory.effective {
				enabled = true
			}
			vcenter.cluster.memory.limit {
				enabled = true
			}
			vcenter.cluster.memory.used {
				enabled = true
			}
			vcenter.cluster.vm.count {
				enabled = true
			}
			vcenter.datastore.disk.usage {
				enabled = true
			}
			vcenter.datastore.disk.utilization {
				enabled = true
			}
			vcenter.host.cpu.usage {
				enabled = true
			}
			vcenter.host.cpu.utilization {
				enabled = true
			}
			vcenter.host.disk.latency.avg {
				enabled = true
			}
			vcenter.host.disk.latency.max {
				enabled = true
			}
			vcenter.host.disk.throughput {
				enabled = true
			}
			vcenter.host.memory.usage {
				enabled = true
			}
			vcenter.host.memory.utilization {
				enabled = true
			}
			vcenter.host.network.packet.count {
				enabled = true
			}
			vcenter.host.network.packet.errors {
				enabled = true
			}
			vcenter.host.network.throughput {
				enabled = true
			}
			vcenter.host.network.usage {
				enabled = true
			}
			vcenter.resource_pool.cpu.shares {
				enabled = true
			}
			vcenter.resource_pool.cpu.usage {
				enabled = true
			}
			vcenter.resource_pool.memory.shares {
				enabled = true
			}
			vcenter.resource_pool.memory.usage {
				enabled = true
			}
			vcenter.vm.cpu.usage {
				enabled = true
			}
			vcenter.vm.cpu.utilization {
				enabled = true
			}
			vcenter.vm.disk.latency.avg {
				enabled = true
			}
			vcenter.vm.disk.latency.max {
				enabled = true
			}
			vcenter.vm.disk.throughput {
				enabled = true
			}
			vcenter.vm.disk.usage {
				enabled = true
			}
			vcenter.vm.disk.utilization {
				enabled = true
			}
			vcenter.vm.memory.ballooned {
				enabled = true
			}
			vcenter.vm.memory.swapped {
				enabled = true
			}
			vcenter.vm.memory.swapped_ssd {
				enabled = true
			}
			vcenter.vm.memory.usage {
				enabled = true
			}
			vcenter.vm.network.packet.count {
				enabled = true
			}
			vcenter.vm.network.throughput {
				enabled = true
			}
			vcenter.vm.network.usage {
				enabled = true
			}
		}

		output { /* no-op */ }
	`

	var args Arguments
	require.NoError(t, river.Unmarshal([]byte(in), &args))
	args.Convert()
	ext, err := args.Convert()
	require.NoError(t, err)
	otelArgs, ok := (ext).(*vcenterreceiver.Config)

	require.True(t, ok)

	require.Equal(t, "user", otelArgs.Username)
	require.Equal(t, "pass", string(otelArgs.Password))
	require.Equal(t, "http://localhost:1234", otelArgs.Endpoint)

	require.Equal(t, 2*time.Minute, otelArgs.ScraperControllerSettings.CollectionInterval)
	require.Equal(t, time.Second, otelArgs.ScraperControllerSettings.InitialDelay)
	require.Equal(t, 0*time.Second, otelArgs.ScraperControllerSettings.Timeout)

	// Verify ResourceAttributesConfig fields
	require.Equal(t, true, otelArgs.ResourceAttributes.VcenterClusterName.Enabled)
	require.Equal(t, true, otelArgs.ResourceAttributes.VcenterDatastoreName.Enabled)
	require.Equal(t, true, otelArgs.ResourceAttributes.VcenterHostName.Enabled)
	require.Equal(t, false, otelArgs.ResourceAttributes.VcenterResourcePoolInventoryPath.Enabled)
	require.Equal(t, true, otelArgs.ResourceAttributes.VcenterResourcePoolName.Enabled)
	require.Equal(t, true, otelArgs.ResourceAttributes.VcenterVMName.Enabled)
	require.Equal(t, true, otelArgs.ResourceAttributes.VcenterVMID.Enabled)

	// Verify MetricsConfig fields
	require.Equal(t, false, otelArgs.Metrics.VcenterClusterCPUEffective.Enabled)
	require.Equal(t, true, otelArgs.Metrics.VcenterClusterCPULimit.Enabled)
	require.Equal(t, true, otelArgs.Metrics.VcenterClusterHostCount.Enabled)
	require.Equal(t, true, otelArgs.Metrics.VcenterClusterMemoryEffective.Enabled)
	require.Equal(t, true, otelArgs.Metrics.VcenterClusterMemoryLimit.Enabled)
	require.Equal(t, true, otelArgs.Metrics.VcenterClusterMemoryUsed.Enabled)
	require.Equal(t, true, otelArgs.Metrics.VcenterClusterVMCount.Enabled)
	require.Equal(t, true, otelArgs.Metrics.VcenterDatastoreDiskUsage.Enabled)
	require.Equal(t, true, otelArgs.Metrics.VcenterDatastoreDiskUtilization.Enabled)
	require.Equal(t, true, otelArgs.Metrics.VcenterHostCPUUsage.Enabled)
	require.Equal(t, true, otelArgs.Metrics.VcenterHostCPUUtilization.Enabled)
	require.Equal(t, true, otelArgs.Metrics.VcenterHostDiskLatencyAvg.Enabled)
	require.Equal(t, true, otelArgs.Metrics.VcenterHostDiskLatencyMax.Enabled)
	require.Equal(t, true, otelArgs.Metrics.VcenterHostDiskThroughput.Enabled)
	require.Equal(t, true, otelArgs.Metrics.VcenterHostMemoryUsage.Enabled)
	require.Equal(t, true, otelArgs.Metrics.VcenterHostMemoryUtilization.Enabled)
	require.Equal(t, true, otelArgs.Metrics.VcenterHostNetworkPacketCount.Enabled)
	require.Equal(t, true, otelArgs.Metrics.VcenterHostNetworkPacketErrors.Enabled)
	require.Equal(t, true, otelArgs.Metrics.VcenterHostNetworkThroughput.Enabled)
	require.Equal(t, true, otelArgs.Metrics.VcenterHostNetworkUsage.Enabled)
	require.Equal(t, true, otelArgs.Metrics.VcenterResourcePoolCPUShares.Enabled)
	require.Equal(t, true, otelArgs.Metrics.VcenterResourcePoolCPUUsage.Enabled)
	require.Equal(t, true, otelArgs.Metrics.VcenterResourcePoolMemoryShares.Enabled)
	require.Equal(t, true, otelArgs.Metrics.VcenterResourcePoolMemoryUsage.Enabled)
	require.Equal(t, true, otelArgs.Metrics.VcenterVMCPUUsage.Enabled)
	require.Equal(t, true, otelArgs.Metrics.VcenterVMCPUUtilization.Enabled)
	require.Equal(t, true, otelArgs.Metrics.VcenterVMDiskLatencyAvg.Enabled)
	require.Equal(t, true, otelArgs.Metrics.VcenterVMDiskLatencyMax.Enabled)
	require.Equal(t, true, otelArgs.Metrics.VcenterVMDiskThroughput.Enabled)
	require.Equal(t, true, otelArgs.Metrics.VcenterVMDiskUsage.Enabled)
	require.Equal(t, true, otelArgs.Metrics.VcenterVMDiskUtilization.Enabled)
	require.Equal(t, true, otelArgs.Metrics.VcenterVMMemoryBallooned.Enabled)
	require.Equal(t, true, otelArgs.Metrics.VcenterVMMemorySwapped.Enabled)
	require.Equal(t, true, otelArgs.Metrics.VcenterVMMemorySwappedSsd.Enabled)
	require.Equal(t, true, otelArgs.Metrics.VcenterVMMemoryUsage.Enabled)
	require.Equal(t, false, otelArgs.Metrics.VcenterVMMemoryUtilization.Enabled)
	require.Equal(t, true, otelArgs.Metrics.VcenterVMNetworkPacketCount.Enabled)
	require.Equal(t, true, otelArgs.Metrics.VcenterVMNetworkThroughput.Enabled)
	require.Equal(t, true, otelArgs.Metrics.VcenterVMNetworkUsage.Enabled)
}
