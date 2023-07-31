package build

import (
	"fmt"

	"github.com/grafana/agent/component/prometheus/exporter/unix"
	"github.com/grafana/agent/converter/internal/common"
	"github.com/grafana/agent/converter/internal/prometheusconvert"
	"github.com/grafana/agent/pkg/integrations/config"
	"github.com/grafana/agent/pkg/integrations/node_exporter"

	"github.com/prometheus/common/model"
	prom_config "github.com/prometheus/prometheus/config"
)

func (b *IntegrationsV1ConfigBuilder) AppendNodeExporter(nodeConfig *node_exporter.Config, commonConfig *config.Common) {
	if !commonConfig.Enabled {
		return
	}

	args := ToNodeExporter(nodeConfig)
	b.f.Body().AppendBlock(common.NewBlockWithOverride(
		[]string{"prometheus", "exporter", "unix"},
		"",
		args,
	))

	exports := prometheusconvert.NewDiscoverExports("prometheus.exporter.unix.targets")

	scrapeConfigs := []*prom_config.ScrapeConfig{}
	if b.cfg.Integrations.ConfigV1.ScrapeIntegrations {
		// TODO: more from func (m *Manager) instanceConfigForIntegration(p *integrationProcess, cfg ManagerConfig) instance.Config {
		scrapeConfig := prom_config.DefaultScrapeConfig
		// scrapeConfig.JobName = fmt.Sprintf("integrations/%s", nodeConfig.Name())
		scrapeConfig.MetricsPath = fmt.Sprintf("integrations/%s/metrics", nodeConfig.Name())
		scrapeConfig.JobName = fmt.Sprintf("integrations/%s", nodeConfig.Name())
		scrapeConfig.RelabelConfigs = commonConfig.RelabelConfigs
		scrapeConfig.MetricRelabelConfigs = commonConfig.MetricRelabelConfigs

		scrapeConfig.ScrapeInterval = model.Duration(commonConfig.ScrapeInterval)
		if commonConfig.ScrapeInterval == 0 {
			scrapeConfig.ScrapeInterval = b.cfg.Integrations.ConfigV1.PrometheusGlobalConfig.ScrapeInterval
		}

		scrapeConfig.ScrapeTimeout = model.Duration(commonConfig.ScrapeTimeout)
		if commonConfig.ScrapeTimeout == 0 {
			scrapeConfig.ScrapeTimeout = b.cfg.Integrations.ConfigV1.PrometheusGlobalConfig.ScrapeTimeout
		}

		scrapeConfigs = []*prom_config.ScrapeConfig{&scrapeConfig}
	}

	promConfig := &prom_config.Config{
		GlobalConfig:       b.cfg.Integrations.ConfigV1.PrometheusGlobalConfig,
		ScrapeConfigs:      scrapeConfigs,
		RemoteWriteConfigs: b.cfg.Integrations.ConfigV1.PrometheusRemoteWrite,
	}
	b.diags.AddAll(prometheusconvert.AppendAll(b.f, promConfig, b.globalCtx.LabelPrefix, exports.Targets))
}

func ToNodeExporter(config *node_exporter.Config) *unix.Arguments {
	return &unix.Arguments{
		IncludeExporterMetrics: config.IncludeExporterMetrics,
		ProcFSPath:             config.ProcFSPath,
		SysFSPath:              config.SysFSPath,
		RootFSPath:             config.RootFSPath,
		UdevDataPath:           config.UdevDataPath,
		EnableCollectors:       config.EnableCollectors,
		DisableCollectors:      config.DisableCollectors,
		SetCollectors:          config.SetCollectors,
		BCache: unix.BCacheConfig{
			PriorityStats: config.BcachePriorityStats,
		},
		CPU: unix.CPUConfig{
			BugsInclude:    config.CPUBugsInclude,
			EnableCPUGuest: config.CPUEnableCPUGuest,
			EnableCPUInfo:  config.CPUEnableCPUInfo,
			FlagsInclude:   config.CPUFlagsInclude,
		},
		Disk: unix.DiskStatsConfig{
			DeviceExclude: config.DiskStatsDeviceExclude,
			DeviceInclude: config.DiskStatsDeviceInclude,
		},
		EthTool: unix.EthToolConfig{
			DeviceExclude:  config.EthtoolDeviceExclude,
			DeviceInclude:  config.EthtoolDeviceInclude,
			MetricsInclude: config.EthtoolMetricsInclude,
		},
		Filesystem: unix.FilesystemConfig{
			FSTypesExclude:     config.FilesystemFSTypesExclude,
			MountPointsExclude: config.FilesystemMountPointsExclude,
			MountTimeout:       config.FilesystemMountTimeout,
		},
		IPVS: unix.IPVSConfig{
			BackendLabels: config.IPVSBackendLabels,
		},
		NTP: unix.NTPConfig{
			IPTTL:                config.NTPIPTTL,
			LocalOffsetTolerance: config.NTPLocalOffsetTolerance,
			MaxDistance:          config.NTPMaxDistance,
			ProtocolVersion:      config.NTPProtocolVersion,
			Server:               config.NTPServer,
			ServerIsLocal:        config.NTPServerIsLocal,
		},
		Netclass: unix.NetclassConfig{
			IgnoreInvalidSpeedDevice: config.NetclassIgnoreInvalidSpeedDevice,
			IgnoredDevices:           config.NetclassIgnoredDevices,
		},
		Netdev: unix.NetdevConfig{
			AddressInfo:   config.NetdevAddressInfo,
			DeviceExclude: config.NetdevDeviceExclude,
			DeviceInclude: config.NetdevDeviceInclude,
		},
		Netstat: unix.NetstatConfig{
			Fields: config.NetstatFields,
		},
		Perf: unix.PerfConfig{
			CPUS:                     config.PerfCPUS,
			Tracepoint:               config.PerfTracepoint,
			DisableHardwareProfilers: config.PerfDisableHardwareProfilers,
			DisableSoftwareProfilers: config.PerfDisableSoftwareProfilers,
			DisableCacheProfilers:    config.PerfDisableCacheProfilers,
			HardwareProfilers:        config.PerfHardwareProfilers,
			SoftwareProfilers:        config.PerfSoftwareProfilers,
			CacheProfilers:           config.PerfCacheProfilers,
		},
		Powersupply: unix.PowersupplyConfig{
			IgnoredSupplies: config.PowersupplyIgnoredSupplies,
		},
		Runit: unix.RunitConfig{
			ServiceDir: config.RunitServiceDir,
		},
		Supervisord: unix.SupervisordConfig{
			URL: config.SupervisordURL,
		},
		Sysctl: unix.SysctlConfig{
			Include:     config.SysctlInclude,
			IncludeInfo: config.SysctlIncludeInfo,
		},
		Systemd: unix.SystemdConfig{
			EnableRestartsMetrics:  config.SystemdEnableRestartsMetrics,
			EnableStartTimeMetrics: config.SystemdEnableStartTimeMetrics,
			EnableTaskMetrics:      config.SystemdEnableTaskMetrics,
			UnitExclude:            config.SystemdUnitExclude,
			UnitInclude:            config.SystemdUnitInclude,
		},
		Tapestats: unix.TapestatsConfig{
			IgnoredDevices: config.TapestatsIgnoredDevices,
		},
		Textfile: unix.TextfileConfig{
			Directory: config.TextfileDirectory,
		},
		VMStat: unix.VMStatConfig{
			Fields: config.VMStatFields,
		},
	}
}
