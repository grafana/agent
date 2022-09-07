package node_exporter

import (
	"os"
	"runtime"
	"time"

	node_integration "github.com/grafana/agent/pkg/integrations/node_exporter"
	"github.com/grafana/dskit/flagext"
	"github.com/prometheus/procfs"
)

var DefaultConfig = Config{
	ProcFSPath: procfs.DefaultMountPoint,
	RootFSPath: "/",
	Disk: DiskStatsConfig{
		IgnoredDevices: node_integration.DefaultConfig.DiskStatsIgnoredDevices,
	},
	EthTool: EthToolConfig{
		MetricsInclude: ".*",
	},
	Filesystem: FilesystemConfig{
		MountTimeout: 5 * time.Second,
	},
	NTP: NTPConfig{
		IPTTL:                1,
		LocalOffsetTolerance: time.Millisecond,
		MaxDistance:          time.Microsecond * 3466080,
		ProtocolVersion:      4,
		Server:               "127.0.0.1",
	},
	Netclass: NetclassConfig{
		IgnoredDevices: "^$",
	},
	Netstat: NetstatConfig{
		Fields: node_integration.DefaultConfig.NetstatFields,
	},
	Powersupply: PowersupplyConfig{
		IgnoredSupplies: "^$",
	},
	Runit: RunitConfig{
		ServiceDir: "/etc/service",
	},
	Supervisord: SupervisordConfig{
		URL: "http://localhost:9001/RPC2",
	},
	Systemd: SystemdConfig{
		UnitExclude: node_integration.DefaultConfig.SystemdUnitExclude,
		UnitInclude: ".+",
	},
	Tapestats: TapestatsConfig{
		IgnoredDevices: "^$",
	},
	VMStat: VMStatConfig{
		Fields: node_integration.DefaultConfig.VMStatFields,
	},
}

func init() {
	// The default values for the filesystem collector are to ignore everything,
	// but some platforms have specific defaults. We'll fill these in below at
	// initialization time, but the values can still be overridden via the config
	// file.
	switch runtime.GOOS {
	case "linux":
		DefaultConfig.Filesystem.MountPointsExclude = "^/(dev|proc|run/credentials/.+|sys|var/lib/docker/.+)($|/)"
		DefaultConfig.Filesystem.FSTypesExclude = "^(autofs|binfmt_misc|bpf|cgroup2?|configfs|debugfs|devpts|devtmpfs|fusectl|hugetlbfs|iso9660|mqueue|nsfs|overlay|proc|procfs|pstore|rpc_pipefs|securityfs|selinuxfs|squashfs|sysfs|tracefs)$"
	case "darwin":
		DefaultConfig.Filesystem.MountPointsExclude = "^/(dev)($|/)"
		DefaultConfig.Filesystem.FSTypesExclude = "^(autofs|devfs)$"
	case "freebsd", "netbsd", "openbsd":
		DefaultConfig.Filesystem.MountPointsExclude = "^/(dev)($|/)"
		DefaultConfig.Filesystem.FSTypesExclude = "^devfs$"
	}
	if url := os.Getenv("SUPERVISORD_URL"); url != "" {
		DefaultConfig.Supervisord.URL = url
	}
}

type Config struct {
	IncludeExporterMetrics bool   `river:"include_exporter_metrics,attr,optional"`
	ProcFSPath             string `river:"procfs_path,attr,optional"`
	SysFSPath              string `river:"sysfs_path,attr,optional"`
	RootFSPath             string `river:"rootfs_path,attr,optional"`

	// Collectors to mark as enabled
	EnableCollectors flagext.StringSlice `river:"enable_collectors,attr,optional"`

	// Collectors to mark as disabled
	DisableCollectors flagext.StringSlice `river:"disable_collectors,attr,optional"`

	// Overrides the default set of enabled collectors with the collectors
	// listed.
	SetCollectors flagext.StringSlice `river:"set_collectors,attr,optional"`

	// Collector-specific config options
	BCache      BCacheConfig      `river:"bcache,block,optional"`
	CPU         CPUConfig         `river:"cpu,block,optional"`
	Disk        DiskStatsConfig   `river:"disk,block,optional"`
	EthTool     EthToolConfig     `river:"ethtool,block,optional"`
	Filesystem  FilesystemConfig  `river:"filesystem,block,optional"`
	IPVS        IPVSConfig        `river:"ipvs,block,optional"`
	NTP         NTPConfig         `river:"ntp,block,optional"`
	Netclass    NetclassConfig    `river:"netclass,block,optional"`
	Netdev      NetdevConfig      `river:"netdev,block,optional"`
	Netstat     NetstatConfig     `river:"netstat,block,optional"`
	Perf        PerfConfig        `river:"perf,block,optional"`
	Powersupply PowersupplyConfig `river:"powersupply,block,optional"`
	Runit       RunitConfig       `river:"runit,block,optional"`
	Supervisord SupervisordConfig `river:"supervisord,block,optional"`
	Systemd     SystemdConfig     `river:"systemd,block,optional"`
	Tapestats   TapestatsConfig   `river:"tapestats,block,optional"`
	Textfile    TextfileConfig    `river:"textfile,block,optional"`
	VMStat      VMStatConfig      `river:"vmstat,block,optional"`
}

func (c *Config) Convert() *node_integration.Config {
	return &node_integration.Config{
		IncludeExporterMetrics:           c.IncludeExporterMetrics,
		ProcFSPath:                       c.ProcFSPath,
		SysFSPath:                        c.SysFSPath,
		RootFSPath:                       c.RootFSPath,
		EnableCollectors:                 c.EnableCollectors,
		DisableCollectors:                c.DisableCollectors,
		SetCollectors:                    c.SetCollectors,
		BcachePriorityStats:              c.BCache.PriorityStats,
		CPUBugsInclude:                   c.CPU.BugsInclude,
		CPUEnableCPUGuest:                c.CPU.EnableCPUGuest,
		CPUEnableCPUInfo:                 c.CPU.EnableCPUInfo,
		CPUFlagsInclude:                  c.CPU.FlagsInclude,
		DiskStatsIgnoredDevices:          c.Disk.IgnoredDevices,
		EthtoolDeviceExclude:             c.EthTool.DeviceExclude,
		EthtoolDeviceInclude:             c.EthTool.DeviceInclude,
		EthtoolMetricsInclude:            c.EthTool.MetricsInclude,
		FilesystemFSTypesExclude:         c.Filesystem.FSTypesExclude,
		FilesystemMountPointsExclude:     c.Filesystem.MountPointsExclude,
		FilesystemMountTimeout:           c.Filesystem.MountTimeout,
		IPVSBackendLabels:                c.IPVS.BackendLabels,
		NTPIPTTL:                         c.NTP.IPTTL,
		NTPLocalOffsetTolerance:          c.NTP.LocalOffsetTolerance,
		NTPMaxDistance:                   c.NTP.MaxDistance,
		NTPProtocolVersion:               c.NTP.ProtocolVersion,
		NTPServer:                        c.NTP.Server,
		NTPServerIsLocal:                 c.NTP.ServerIsLocal,
		NetclassIgnoreInvalidSpeedDevice: c.Netclass.IgnoreInvalidSpeedDevice,
		NetclassIgnoredDevices:           c.Netclass.IgnoredDevices,
		NetdevAddressInfo:                c.Netdev.AddressInfo,
		NetdevDeviceExclude:              c.Netdev.DeviceExclude,
		NetdevDeviceInclude:              c.Netdev.DeviceInclude,
		NetstatFields:                    c.Netstat.Fields,
		PerfCPUS:                         c.Perf.CPUS,
		PerfTracepoint:                   c.Perf.Tracepoint,
		PowersupplyIgnoredSupplies:       c.Powersupply.IgnoredSupplies,
		RunitServiceDir:                  c.Runit.ServiceDir,
		SupervisordURL:                   c.Supervisord.URL,
		SystemdEnableRestartsMetrics:     c.Systemd.EnableRestartsMetrics,
		SystemdEnableStartTimeMetrics:    c.Systemd.EnableStartTimeMetrics,
		SystemdEnableTaskMetrics:         c.Systemd.EnableTaskMetrics,
		SystemdUnitExclude:               c.Systemd.UnitExclude,
		SystemdUnitInclude:               c.Systemd.UnitInclude,
		TapestatsIgnoredDevices:          c.Tapestats.IgnoredDevices,
		TextfileDirectory:                c.Textfile.Directory,
		VMStatFields:                     c.VMStat.Fields,
	}
}

// UnmarshalRiver implements River unmarshalling for Config
func (c *Config) UnmarshalRiver(f func(interface{}) error) error {
	*c = DefaultConfig

	type cfg Config
	return f((*cfg)(c))
}

type PowersupplyConfig struct {
	IgnoredSupplies string `river:"ignored_supplies,attr,optional"`
}

type RunitConfig struct {
	ServiceDir string `river:"service_dir,attr,optional"`
}

type SupervisordConfig struct {
	URL string `river:"url,attr,optional"`
}

type TapestatsConfig struct {
	IgnoredDevices string `river:"ignored_devices,attr,optional"`
}

type TextfileConfig struct {
	Directory string `river:"directory,attr,optional"`
}

type VMStatConfig struct {
	Fields string `river:"fields,attr,optional"`
}

type NetclassConfig struct {
	IgnoreInvalidSpeedDevice bool   `river:"ignore_invalid_speed_device,attr,optional"`
	IgnoredDevices           string `river:"ignored_devices,attr,optional"`
}

type NetdevConfig struct {
	AddressInfo   bool   `river:"address_info,attr,optional"`
	DeviceExclude string `river:"device_exclude,attr,optional"`
	DeviceInclude string `river:"device_include,attr,optional"`
}

type NetstatConfig struct {
	Fields string `river:"fields,attr,optional"`
}

type PerfConfig struct {
	CPUS       string              `river:"cpus,attr,optional"`
	Tracepoint flagext.StringSlice `river:"tracepoint,attr,optional"`
}

type EthToolConfig struct {
	DeviceExclude  string `river:"device_exclude,attr,optional"`
	DeviceInclude  string `river:"device_include,attr,optional"`
	MetricsInclude string `river:"metrics_include,attr,optional"`
}

type FilesystemConfig struct {
	FSTypesExclude     string        `river:"fs_types_exclude,attr,optional"`
	MountPointsExclude string        `river:"mount_points_exclude,attr,optional"`
	MountTimeout       time.Duration `river:"mount_timeout,attr,optional"`
}

type IPVSConfig struct {
	BackendLabels []string `river:"backend_labels,attr,optional"`
}

type BCacheConfig struct {
	PriorityStats bool `river:"priority_stats,attr,optional"`
}

type CPUConfig struct {
	BugsInclude    string `river:"bugs_include,attr,optional"`
	EnableCPUGuest bool   `river:"guest,attr,optional"`
	EnableCPUInfo  bool   `river:"info,attr,optional"`
	FlagsInclude   string `river:"flags_include,attr,optional"`
}

type DiskStatsConfig struct {
	IgnoredDevices string `river:"ignored_devices,attr,optional"`
}

type NTPConfig struct {
	IPTTL                int           `river:"ip_ttl,attr,optional"`
	LocalOffsetTolerance time.Duration `river:"local_offset_tolerance,attr,optional"`
	MaxDistance          time.Duration `river:"max_distance,attr,optional"`
	ProtocolVersion      int           `river:"protocol_version,attr,optional"`
	Server               string        `river:"server,attr,optional"`
	ServerIsLocal        bool          `river:"server_is_local,attr,optional"`
}

type SystemdConfig struct {
	EnableRestartsMetrics  bool   `river:"enable_restarts,attr,optional"`
	EnableStartTimeMetrics bool   `river:"start_time,attr,optional"`
	EnableTaskMetrics      bool   `river:"task_metrics,attr,optional"`
	UnitExclude            string `river:"unit_exclude,attr,optional"`
	UnitInclude            string `river:"unit_include,attr,optional"`
}
