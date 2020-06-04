package node_exporter //nolint:golint

import (
	"flag"
	"fmt"
	"time"

	"github.com/grafana/agent/pkg/integrations/config"
	"gopkg.in/alecthomas/kingpin.v2"
)

// Config controls the node_exporter integration.
type Config struct {
	CommonConfig config.Common `yaml:",inline"`

	// Enabled enables the node_exporter integration.
	Enabled bool

	IncludeExporterMetrics bool `yaml:"include_exporter_metrics"`

	ProcFSPath string `yaml:"procfs_path"`
	SysFSPath  string `yaml:"sysfs_path"`
	RootFSPath string `yaml:"rootfs_path"`

	EnableARPCollector         bool `yaml:"enable_arp_collector"`
	EnableBCacheCollector      bool `yaml:"enable_bcache_collector"`
	EnableBondingCollector     bool `yaml:"enable_bonding_collector"`
	EnableBoottimeCollector    bool `yaml:"enable_boottime_collector"`
	EnableBtrfsCollector       bool `yaml:"enable_btrfs_collector"`
	EnableBuddyinfoCollector   bool `yaml:"enable_buddyinfo_collector"`
	EnableCPUFreqCollector     bool `yaml:"enable_cpufreq_collector"`
	EnableConntrackCollector   bool `yaml:"enable_conntrack_collector"`
	EnableDRBDCollector        bool `yaml:"enable_drbd_collector"`
	EnableDevStatCollector     bool `yaml:"enable_devstat_collector"`
	EnableEDACCollector        bool `yaml:"enable_edac_collector"`
	EnableEntropyCollector     bool `yaml:"enable_entropy_collector"`
	EnableExecCollector        bool `yaml:"enable_exec_collector"`
	EnableFileFDCollector      bool `yaml:"enable_filefd_collector"`
	EnableHWMonCollector       bool `yaml:"enable_hwmon_collector"`
	EnableIPVSCollector        bool `yaml:"enable_ipvs_collector"`
	EnableInfiniBandCollector  bool `yaml:"enable_infiniband_collector"`
	EnableInterruptsCollector  bool `yaml:"enable_interrupts_collector"`
	EnableKSMDCollector        bool `yaml:"enable_ksmd_collector"`
	EnableLoadAvgCollector     bool `yaml:"enable_load_avg_collector"`
	EnableLoginDCollector      bool `yaml:"enable_logind_collector"`
	EnableMDADMCollector       bool `yaml:"enable_mdadm_collector"`
	EnableMemInfoCollector     bool `yaml:"enable_meminfo_collector"`
	EnableMeminfoNUMACollector bool `yaml:"enable_meminfo_numa_collector"`
	EnableMountStatsCollector  bool `yaml:"enable_mountstats_collector"`
	EnableNFSCollector         bool `yaml:"enable_nfs_collector"`
	EnableNFSDCollector        bool `yaml:"enable_nfsd_collector"`
	EnablePressureCollector    bool `yaml:"enable_pressure_collector"`
	EnableProcessesCollector   bool `yaml:"enable_processes_collector"`
	EnableQDiscCollector       bool `yaml:"enable_qdisc_collector"`
	EnableRAPLCollector        bool `yaml:"enable_rapl_collector"`
	EnableSchedStatCollector   bool `yaml:"enable_schedstat_collector"`
	EnableSockStatCollector    bool `yaml:"enable_sockstat_collector"`
	EnableSoftNetCollector     bool `yaml:"enable_softnet_collector"`
	EnableStatCollector        bool `yaml:"enable_stat_collector"`
	EnableTCPStatCollector     bool `yaml:"enable_tcpstat_collector"`
	EnableThermalZoneCollector bool `yaml:"enable_thermal_zone_collector"`
	EnableTimeCollector        bool `yaml:"enable_time_collector"`
	EnableTimexCollector       bool `yaml:"enable_timex_collector"`
	EnableUDPQueuesCollector   bool `yaml:"enable_udp_queues_collector"`
	EnableUNameCollector       bool `yaml:"enable_uname_collector"`
	EnableWiFiCollector        bool `yaml:"enable_wifi_collector"`
	EnableXFSCollector         bool `yaml:"enable_xfs_collector"`
	EnableZFSCollector         bool `yaml:"enable_zfs_collector"`

	CPUCollector         CPUConfig         `yaml:"cpu_collector"`
	DiskStatsCollector   DiskStatsConfig   `yaml:"diskstats_collector"`
	FilesystemCollector  FilesystemConfig  `yaml:"filesystem_collector"`
	NTPCollector         NTPConfig         `yaml:"ntp_collector"`
	NetclassCollector    NetclassConfig    `yaml:"netclass_collector"`
	NetdevCollector      NetdevConfig      `yaml:"netdev_collector"`
	NetstatCollector     NetstatConfig     `yaml:"netstat_collector"`
	PerfCollector        PerfConfig        `yaml:"perf_collector"`
	PowerSupplyCollector PowerSupplyConfig `yaml:"powersupply_collector"`
	RunitCollector       RunitConfig       `yaml:"runit_collector"`
	SupervisordCollector SupervisordConfig `yaml:"supervisord_collector"`
	SystemdCollector     SystemdConfig     `yaml:"systemd_collector"`
	TextfileCollector    TextfileConfig    `yaml:"textfile_collector"`
	VMStatCollector      VMStatConfig      `yaml:"vmstat_collector"`
}

func (c *Config) RegisterFlagsWithPrefix(prefix string, f *flag.FlagSet) {
	prefix = prefix + "node_exporter."

	c.CommonConfig.RegisterFlagsWithPrefix(prefix, f)

	c.CPUCollector.RegisterFlagsWithPrefix(prefix, f)
	c.DiskStatsCollector.RegisterFlagsWithPrefix(prefix, f)
	c.FilesystemCollector.RegisterFlagsWithPrefix(prefix, f)
	c.NTPCollector.RegisterFlagsWithPrefix(prefix, f)
	c.NetclassCollector.RegisterFlagsWithPrefix(prefix, f)
	c.NetdevCollector.RegisterFlagsWithPrefix(prefix, f)
	c.NetstatCollector.RegisterFlagsWithPrefix(prefix, f)
	c.PerfCollector.RegisterFlagsWithPrefix(prefix, f)
	c.PowerSupplyCollector.RegisterFlagsWithPrefix(prefix, f)
	c.RunitCollector.RegisterFlagsWithPrefix(prefix, f)
	c.SupervisordCollector.RegisterFlagsWithPrefix(prefix, f)
	c.SystemdCollector.RegisterFlagsWithPrefix(prefix, f)
	c.TextfileCollector.RegisterFlagsWithPrefix(prefix, f)
	c.VMStatCollector.RegisterFlagsWithPrefix(prefix, f)

	f.BoolVar(&c.Enabled, prefix+"enabled", false, "enable the node_exporter integration collect metrics from the host Linux system")
	f.BoolVar(&c.IncludeExporterMetrics, prefix+"include-exporter-metrics", false, "include metrics on the integration itself")

	f.StringVar(&c.ProcFSPath, prefix+"procfs-path", "/proc", "procfs mountpoint")
	f.StringVar(&c.SysFSPath, prefix+"sysfs-path", "/sys", "sysfs mountpoint")
	f.StringVar(&c.SysFSPath, prefix+"rootfs-path", "/", "rootfs mountpoint")

	f.BoolVar(&c.EnableARPCollector, prefix+"enable-arp-collector", true, "Exposes ARP statistics from /proc/net/arp. OS: Linux")
	f.BoolVar(&c.EnableBCacheCollector, prefix+"enable-bcache-collector", true, "Exposes bcache statistics from /sys/fs/bcache. OS: Linux")
	f.BoolVar(&c.EnableBondingCollector, prefix+"enable-bonding-collector", true, "Exposes the number of configured and active slaves of Linux bonding interfaces. OS: Linux")
	f.BoolVar(&c.EnableBoottimeCollector, prefix+"enable-boottime-collector", true, "Exposes system boot time derived from the kern.boottime sysctl. OS: Darwin, Dragonfly, FreeBSD, NetBSD, OpenBSD, Solaris")
	f.BoolVar(&c.EnableBtrfsCollector, prefix+"enable-btrfs-collector", true, "Exposes statistics on btrfs. OS: Linux")
	f.BoolVar(&c.EnableCPUFreqCollector, prefix+"enable-cpufreq-collector", true, "Exposes CPU frequency statistics. OS: Linux, Solaris")
	f.BoolVar(&c.EnableConntrackCollector, prefix+"enable-conntrack-collector", true, "Shows conntrack statistics (does nothing if no /proc/sys/net/netfilter/ present). OS: Linux")
	f.BoolVar(&c.EnableEDACCollector, prefix+"enable-edac-collector", true, "Exposes error detection and correction statistics. OS: Linux")
	f.BoolVar(&c.EnableEntropyCollector, prefix+"enable-entropy-collector", true, "Exposes available entropy. OS: Linux")
	f.BoolVar(&c.EnableExecCollector, prefix+"enable-exec-collector", true, "Exposes execution statistics. OS: Dragonfly, FreeBSD")
	f.BoolVar(&c.EnableFileFDCollector, prefix+"enable-filefd-collector", true, "Exposes file descriptor statistics from /proc/sys/fs/file-nr. OS: Linux")
	f.BoolVar(&c.EnableHWMonCollector, prefix+"enable-hwmon-collector", true, "Exposes hardware monitoring and sensor data from /sys/class/hwmon. OS: Linux")
	f.BoolVar(&c.EnableIPVSCollector, prefix+"enable-ipvs-collector", true, "Exposes IPVS status from /proc/net/ip_vs and stats from /proc/net/ip_vs_stats. OS: Linux")
	f.BoolVar(&c.EnableInfiniBandCollector, prefix+"enable-infiniband-collector", true, "Exposes network statistics specific to InfiniBand and Intel OmniPath configurations. OS: Linux")
	f.BoolVar(&c.EnableLoadAvgCollector, prefix+"enable-load-avg-collector", true, "Exposes load average. OS: Darwin, Dragonfly, FreeBSD, Linux, NetBSD, OpenBSD, Solaris")
	f.BoolVar(&c.EnableMDADMCollector, prefix+"enable-mdadm-collector", true, "Exposes statistics about devices in /proc/mdstat (does nothing if no /proc/mdstat present). OS: Linux")
	f.BoolVar(&c.EnableMemInfoCollector, prefix+"enable-meminfo-collector", true, "Exposes memory statistics. OS: Darwin, Dragonfly, FreeBSD, Linux, OpenBSD")
	f.BoolVar(&c.EnableNFSCollector, prefix+"enable-nfs-collector", true, "Exposes NFS client statistics from /proc/net/rpc/nfs. This is the same information as nfsstat -c. OS: Linux")
	f.BoolVar(&c.EnableNFSDCollector, prefix+"enable-nfsd-collector", true, "Exposes NFS kernel server statistics from /proc/net/rpc/nfsd. This is the same information as nfsstat -s. OS: Linux")
	f.BoolVar(&c.EnablePressureCollector, prefix+"enable-pressure-collector", true, "Exposes pressure stall statistics from /proc/pressure/. OS: Linux (kernel 4.20+ and/or CONFIG_PSI)")
	f.BoolVar(&c.EnableRAPLCollector, prefix+"enable-rapl-collector", true, "Exposes various statistics from /sys/class/powercap. OS: Linux")
	f.BoolVar(&c.EnableSchedStatCollector, prefix+"enable-schedstat-collector", true, "Exposes task scheduler statistics from /proc/schedstat. OS: Linux")
	f.BoolVar(&c.EnableSockStatCollector, prefix+"enable-sockstat-collector", true, "Exposes various statistics from /proc/net/sockstat. OS: Linux")
	f.BoolVar(&c.EnableSoftNetCollector, prefix+"enable-softnet-collector", true, "Exposes statistics from /proc/net/softnet_stat. OS: Linux")
	f.BoolVar(&c.EnableStatCollector, prefix+"enable-stat-collector", true, "Exposes various statistics from /proc/stat. This includes boot time, forks and interrupts. OS: Linux")
	f.BoolVar(&c.EnableThermalZoneCollector, prefix+"enable-thermal-zone-collector", true, "Exposes thermal zone & cooling device statistics from /sys/class/thermal. OS: Linux")
	f.BoolVar(&c.EnableTimeCollector, prefix+"enable-time-collector", true, "Exposes the current system time. OS: any")
	f.BoolVar(&c.EnableTimexCollector, prefix+"enable-timex-collector", true, "Exposes selected adjtimex(2) system call stats. OS: Linux")
	f.BoolVar(&c.EnableUDPQueuesCollector, prefix+"enable-udp-queues-collector", true, "Exposes UDP total lengths of the rx_queue and tx_queue from /proc/net/udp and /proc/net/udp6. OS: Linux")
	f.BoolVar(&c.EnableUNameCollector, prefix+"enable-uname-collector", true, "Exposes system information as provided by the uname system call. OS: Darwin, FreeBSD, Linux, OpenBSD")
	f.BoolVar(&c.EnableXFSCollector, prefix+"enable-xfs-collector", true, "Exposes XFS runtime statistics. OS: Linux (kernel 4.4+)")
	f.BoolVar(&c.EnableZFSCollector, prefix+"enable-zfs-collector", true, "Exposes ZFS performance statistics. OS: Linux, Solaris")

	// Disabled by default
	f.BoolVar(&c.EnableBuddyinfoCollector, prefix+"enable-buddyinfo-collector", false, "Exposes statistics of memory fragments as reported by /proc/buddyinfo. OS: Linux")
	f.BoolVar(&c.EnableDRBDCollector, prefix+"enable-drbd-collector", false, "Exposes Distributed Replicated Block Device statistics (to version 8.4). OS: Linux")
	f.BoolVar(&c.EnableDevStatCollector, prefix+"enable-devstat-collector", false, "Exposes device statistics. OS: Dragonfly, FreeBSD")
	f.BoolVar(&c.EnableInterruptsCollector, prefix+"enable-interrupts-collector", false, "Exposes detailed interrupts statistics. OS: Linux, OpenBSD")
	f.BoolVar(&c.EnableKSMDCollector, prefix+"enable-ksmd-collector", false, "Exposes kernel and system statistics from /sys/kernel/mm/ksm. OS: Linux")
	f.BoolVar(&c.EnableLoginDCollector, prefix+"enable-logind-collector", false, "Exposes session counts from logind. OS: Linux")
	f.BoolVar(&c.EnableMeminfoNUMACollector, prefix+"enable-meminfo-numa-collector", false, "Exposes memory statistics from /proc/meminfo_numa. OS: Linux")
	f.BoolVar(&c.EnableMountStatsCollector, prefix+"enable-mountstats-collector", false, "Exposes filesystem statistics from /proc/self/mountstats. Exposes detailed NFS client statistics. OS: Linux")
	f.BoolVar(&c.EnableProcessesCollector, prefix+"enable-processes-collector", false, "Exposes aggregate process statistics from /proc. OS: Linux")
	f.BoolVar(&c.EnableQDiscCollector, prefix+"enable-qdisc-collector", false, "Exposes queuing discipline statistics. OS: Linux")
	f.BoolVar(&c.EnableTCPStatCollector, prefix+"enable-tcpstat-collector", false, "Exposes TCP connection status information from /proc/net/tcp and /proc/net/tcp6. (Warning: the current version has potential performance issues in high load situations). OS: Linux")
	f.BoolVar(&c.EnableWiFiCollector, prefix+"enable-wifi-collector", false, "Exposes WiFi device and station statistics. OS: Linux")
}

type CPUConfig struct {
	Enabled bool `yaml:"enabled"`
	Info    bool `yaml:"enable_cpu_info_metric"`
}

func (c *CPUConfig) RegisterFlagsWithPrefix(prefix string, f *flag.FlagSet) {
	f.BoolVar(&c.Enabled, prefix+"cpu-collector.enabled", true, "Exposes CPU statistics. OS: Darwin, Dragonfly, FreeBSD, Linux, Solaris")
	f.BoolVar(&c.Info, prefix+"cpu-collector.enable-cpu_info-metric", false, "Enable metric cpu_info")
}

func (c *CPUConfig) NodeExporterFlags() (flags []string) {
	if kingpin.CommandLine.GetFlag("collector.cpu") == nil {
		return
	}

	if c.Enabled {
		flags = append(flags, "--collector.cpu")
		flags = append(flags, booleanFlagMap(map[*bool]string{
			&c.Info: "collector.cpu.info",
		})...)
	} else {
		flags = append(flags, "--no-collector.cpu")
	}

	return
}

type DiskStatsConfig struct {
	Enabled        bool   `yaml:"enabled"`
	IgnoredDevices string `yaml:"ignored_devices"`
}

func (c *DiskStatsConfig) RegisterFlagsWithPrefix(prefix string, f *flag.FlagSet) {
	f.BoolVar(&c.Enabled, prefix+"diskstats-collector.enabled", true, "Exposes disk I/O statistics. OS: Darwin, Linux, OpenBSD")
	f.StringVar(&c.IgnoredDevices, prefix+"diskstats-collector.ignored-devices", "^(ram|loop|fd|(h|s|v|xv)d[a-z]|nvme\\d+n\\d+p)\\d+$", "Regexp of devices to ignore for diskstats.")
}

func (c *DiskStatsConfig) NodeExporterFlags() (flags []string) {
	if kingpin.CommandLine.GetFlag("collector.diskstats") == nil {
		return
	}

	if c.Enabled {
		flags = append(flags,
			"--collector.diskstats",
			"--collector.diskstats.ignored-devices", c.IgnoredDevices,
		)
	} else {
		flags = append(flags, "--no-collector.diskstats")
	}

	return
}

type FilesystemConfig struct {
	Enabled            bool   `yaml:"enabled"`
	IgnoredMountPoints string `yaml:"ignored_mount_points"`
	IgnoredFSTypes     string `yaml:"ignored_fs_types"`
}

func (c *FilesystemConfig) RegisterFlagsWithPrefix(prefix string, f *flag.FlagSet) {
	f.BoolVar(&c.Enabled, prefix+"filesystem-collector.enabled", true, "Exposes filesystem statistics, such as disk space used. OS: Darwin, Dragonfly, FreeBSD, Linux, OpenBSD")
	f.StringVar(&c.IgnoredMountPoints, prefix+"filesystem-collector.ignored-mount-points", "^/(dev|proc|sys|var/lib/docker/.+)($|/)", "Regexp of mount points to ignore for filesystem collector.")
	f.StringVar(&c.IgnoredFSTypes, prefix+"filesystem-collector.ignored-fs-types", "^(autofs|binfmt_misc|bpf|cgroup2?|configfs|debugfs|devpts|devtmpfs|fusectl|hugetlbfs|iso9660|mqueue|nsfs|overlay|proc|procfs|pstore|rpc_pipefs|securityfs|selinuxfs|squashfs|sysfs|tracefs)$", "Regexmp of filesystem types to ignore for filesystem collector.")
}

func (c *FilesystemConfig) NodeExporterFlags() (flags []string) {
	if kingpin.CommandLine.GetFlag("collector.filesystem") == nil {
		return
	}

	if c.Enabled {
		flags = append(flags,
			"--collector.filesystem",
			"--collector.filesystem.ignored-mount-points", c.IgnoredMountPoints,
			"--collector.filesystem.ignored-fs-types", c.IgnoredFSTypes,
		)
	} else {
		flags = append(flags, "--no-collector.filesystem")
	}

	return
}

type NetclassConfig struct {
	Enabled        bool   `yaml:"enabled"`
	IgnoredDevices string `yaml:"ignored_devices"`
}

func (c *NetclassConfig) RegisterFlagsWithPrefix(prefix string, f *flag.FlagSet) {
	f.BoolVar(&c.Enabled, prefix+"netclass-collector.enabled", true, "Exposes network interface info from /sys/class/net. OS: Linux")
	f.StringVar(&c.IgnoredDevices, prefix+"netclass-collector.ignored-devices", "^$", "Regexp of net devices to ignore for netclass collector.")
}

func (c *NetclassConfig) NodeExporterFlags() (flags []string) {
	if kingpin.CommandLine.GetFlag("collector.netclass") == nil {
		return
	}

	if c.Enabled {
		flags = append(flags,
			"--collector.netclass",
			"--collector.netclass.ignored-devices", c.IgnoredDevices,
		)
	} else {
		flags = append(flags, "--no-collector.netclass")
	}

	return
}

type NetdevConfig struct {
	Enabled         bool   `yaml:"enabled"`
	DeviceBlacklist string `yaml:"device_blacklist"`
	DeviceWhitelist string `yaml:"device_whitelist"`
}

func (c *NetdevConfig) RegisterFlagsWithPrefix(prefix string, f *flag.FlagSet) {
	f.BoolVar(&c.Enabled, prefix+"netdev-collector.enabled", true, "Exposes network interface statistics such as bytes transferred. OS: Darwin, Dragonfly, FreeBSD, Linux, OpenBSD")
	f.StringVar(&c.DeviceBlacklist, prefix+"netdev-collector.device-blacklist", "", "Regexp of net devices to blacklist (mutually exclusive to device-whitelist)")
	f.StringVar(&c.DeviceWhitelist, prefix+"netdev-collector.device-whitelist", "", "Regexp of net devices to whitelist (mutually exclusive to device-blacklist)")
}

func (c *NetdevConfig) NodeExporterFlags() (flags []string) {
	if kingpin.CommandLine.GetFlag("collector.netdev") == nil {
		return
	}

	if c.Enabled {
		flags = append(flags,
			"--collector.netdev",
			"--collector.netdev.device-blacklist", c.DeviceBlacklist,
			"--collector.netdev.device-whitelist", c.DeviceWhitelist,
		)
	} else {
		flags = append(flags, "--no-collector.netdev")
	}

	return
}

type NetstatConfig struct {
	Enabled bool   `yaml:"enabled"`
	Fields  string `yaml:"fields"`
}

func (c *NetstatConfig) RegisterFlagsWithPrefix(prefix string, f *flag.FlagSet) {
	f.BoolVar(&c.Enabled, prefix+"netstat-collector.enabled", true, "Exposes network statistics from /proc/net/netstat. This is the same information as netstat -s. OS: Linux")
	f.StringVar(&c.Fields, prefix+"netstat-collector.fields", "^(.*_(InErrors|InErrs)|Ip_Forwarding|Ip(6|Ext)_(InOctets|OutOctets)|Icmp6?_(InMsgs|OutMsgs)|TcpExt_(Listen.*|Syncookies.*|TCPSynRetrans)|Tcp_(ActiveOpens|InSegs|OutSegs|PassiveOpens|RetransSegs|CurrEstab)|Udp6?_(InDatagrams|OutDatagrams|NoPorts|RcvbufErrors|SndbufErrors))$", "Regexp of fields to return for netstat collector.")
}

func (c *NetstatConfig) NodeExporterFlags() (flags []string) {
	if kingpin.CommandLine.GetFlag("collector.netstat") == nil {
		return
	}

	if c.Enabled {
		flags = append(flags,
			"--collector.netstat",
			"--collector.netstat.fields", c.Fields,
		)
	} else {
		flags = append(flags, "--no-collector.netstat")
	}

	return
}

type NTPConfig struct {
	Enabled              bool          `yaml:"enabled"`
	Server               string        `yaml:"server"`
	ProtocolVersion      int           `yaml:"protocol_version"`
	ServerIsLocal        bool          `yaml:"server-is-local"`
	IPTTL                int           `yaml:"ip_ttl"`
	MaxDistance          time.Duration `yaml:"max_distance"`
	LocalOffsetTolerance time.Duration `yaml:"local_offset_tolerance"`
}

func (c *NTPConfig) RegisterFlagsWithPrefix(prefix string, f *flag.FlagSet) {
	f.BoolVar(&c.Enabled, prefix+"ntp-collector.enabled", false, "Exposes local NTP daemon helath to check time. OS: any")
	f.StringVar(&c.Server, prefix+"ntp-collector.server", "127.0.0.1", "NTP server to use for ntp collector")
	f.IntVar(&c.ProtocolVersion, prefix+"ntp-collector.protocol-version", 4, "NTP protocol version")
	f.BoolVar(&c.ServerIsLocal, prefix+"ntp-collector.server-is-local", false, "Certify that collector.ntp.server address is not a public ntp server")
	f.IntVar(&c.IPTTL, prefix+"ntp-collector.ip-ttl", 1, "IP TTL to use while sending NTP query")
	f.DurationVar(&c.MaxDistance, prefix+"ntp-collector.max-distance", time.Microsecond*3466080, "Max accumulated distance to the root")
	f.DurationVar(&c.LocalOffsetTolerance, prefix+"ntp-collector.local-offset-tolerance", time.Millisecond, "Offset between local clock and local ntpd time to tolerate")
}

func (c *NTPConfig) NodeExporterFlags() (flags []string) {
	if kingpin.CommandLine.GetFlag("collector.ntp") == nil {
		return
	}

	if c.Enabled {
		flags = append(flags,
			"--collector.ntp",
			"--collector.ntp.server", c.Server,
			"--collector.ntp.protocol-version", fmt.Sprintf("%d", c.ProtocolVersion),
			"--collector.ntp.ip-ttl", fmt.Sprintf("%d", c.IPTTL),
			"--collector.ntp.max-distance", c.MaxDistance.String(),
			"--collector.ntp.local-offset-tolerance", c.LocalOffsetTolerance.String(),
		)

		flags = append(flags, booleanFlagMap(map[*bool]string{
			&c.ServerIsLocal: "collector.ntp.server-is-local",
		})...)
	} else {
		flags = append(flags, "--no-collector.ntp")
	}

	return
}

type PerfConfig struct {
	Enabled    bool     `yaml:"enabled"`
	CPUS       string   `yaml:"cpus"`
	Tracepoint []string `yaml:"tracepoint"`
}

func (c *PerfConfig) RegisterFlagsWithPrefix(prefix string, f *flag.FlagSet) {
	f.BoolVar(&c.Enabled, prefix+"perf-collector.enabled", false, "Exposes perf based metrics (Warning: Metrics are dependent on kernel configuration and settings). OS: Linux")
	f.StringVar(&c.CPUS, prefix+"perf-collector.cpus", "", "List of CPUs from which perf metrics should be collected")
	// TODO: support tracepoint config by flag. not 100% necessary since we anticipate most configuration
	// will be done via YAML.
}

func (c *PerfConfig) NodeExporterFlags() (flags []string) {
	if kingpin.CommandLine.GetFlag("collector.perf") == nil {
		return
	}

	if c.Enabled {
		flags = append(flags,
			"--collector.perf",
			"--collector.perf.cpus", c.CPUS,
		)

		// Kingpin expects flags that support multiple values to be passed in
		// multiple times.
		for _, tp := range c.Tracepoint {
			flags = append(flags, "--collector.perf.tracepoint", tp)
		}
	} else {
		flags = append(flags, "--no-collector.perf")
	}

	return
}

type PowerSupplyConfig struct {
	Enabled         bool   `yaml:"enabled"`
	IgnoredSupplies string `yaml:"ignored_supplies"`
}

func (c *PowerSupplyConfig) RegisterFlagsWithPrefix(prefix string, f *flag.FlagSet) {
	f.BoolVar(&c.Enabled, prefix+"powersupply-collector.enabled", true, "Enable the powersupply collector.")
	f.StringVar(&c.IgnoredSupplies, prefix+"powersupply-collector.ignored-supplies", "^$", "Regexp of power supplies to ignore for powersupplyclass collector.")
}

func (c *PowerSupplyConfig) NodeExporterFlags() (flags []string) {
	if kingpin.CommandLine.GetFlag("collector.powersupply") == nil {
		return
	}

	if c.Enabled {
		flags = append(flags,
			"--collector.powersupply",
			"--collector.powersupply.ignored-supplies", c.IgnoredSupplies,
		)
	} else {
		flags = append(flags, "--no-collector.powersupply")
	}

	return
}

type RunitConfig struct {
	Enabled    bool   `yaml:"enabled"`
	ServiceDir string `yaml:"service_dir"`
}

func (c *RunitConfig) RegisterFlagsWithPrefix(prefix string, f *flag.FlagSet) {
	f.BoolVar(&c.Enabled, prefix+"runit-collector.enabled", false, "Exposes service status from runit. OS: any")
	f.StringVar(&c.ServiceDir, prefix+"runit-collector.service-dir", "/etc/service", "Path to runit service directory.")
}

func (c *RunitConfig) NodeExporterFlags() (flags []string) {
	if kingpin.CommandLine.GetFlag("collector.runit") == nil {
		return
	}

	if c.Enabled {
		flags = append(flags,
			"--collector.runit",
			"--collector.runit.servicedir", c.ServiceDir,
		)
	} else {
		flags = append(flags, "--no-collector.runit")
	}

	return
}

type SupervisordConfig struct {
	Enabled bool   `yaml:"enabled"`
	URL     string `yaml:"url"`
}

func (c *SupervisordConfig) RegisterFlagsWithPrefix(prefix string, f *flag.FlagSet) {
	f.BoolVar(&c.Enabled, prefix+"supervisord-collector.enabled", false, "Exposes service status from supervisord. OS: any")
	f.StringVar(&c.URL, prefix+"supervisord-collector.url", "http://localhost:9001/RPC2", "XML RFC endpoint")
}

func (c *SupervisordConfig) NodeExporterFlags() (flags []string) {
	if kingpin.CommandLine.GetFlag("collector.supervisord") == nil {
		return
	}

	if c.Enabled {
		flags = append(flags,
			"--collector.supervisord",
			"--collector.supervisord.url", c.URL,
		)
	} else {
		flags = append(flags, "--no-collector.supervisord")
	}

	return
}

type SystemdConfig struct {
	Enabled                bool   `yaml:"enabled"`
	UnitWhitelist          string `yaml:"unit_whitelist"`
	UnitBlacklist          string `yaml:"unit_blacklist"`
	EnableTaskMetrics      bool   `yaml:"enable_task_metrics"`
	EnableRestartsMetrics  bool   `yaml:"enable_restarts_metrics"`
	EnableStartTimeMetrics bool   `yaml:"enable_start_time_metrics"`
}

func (c *SystemdConfig) RegisterFlagsWithPrefix(prefix string, f *flag.FlagSet) {
	f.BoolVar(&c.Enabled, prefix+"systemd-collector.enabled", false, "Exposes service and system status from systemd. OS: Linux")
	f.StringVar(&c.UnitWhitelist, prefix+"systemd-collector.unit-whitelist", ".+", "Regexp of systemd units to whitelist. Units must both match whitelist and not match blacklist to be included.")
	f.StringVar(&c.UnitBlacklist, prefix+"systemd-collector.unit-blacklist", ".+\\.(automount|device|mount|scope|slice)", "Regexp of systemd units to blacklist. Units must both match whitelist and not match blacklist to be included.")
	f.BoolVar(&c.EnableTaskMetrics, prefix+"systemd-collector.enable-task-metrics", false, "Enables service unit tasks metrics unit_tasks_current and unit_tasks_max")
	f.BoolVar(&c.EnableRestartsMetrics, prefix+"systemd-collector.enable-restarts-metrics", false, "Enables service unit metric service_restart_total")
	f.BoolVar(&c.EnableStartTimeMetrics, prefix+"systemd-collector.enable-start-time-metrics", false, "Enables service unit metric unit_start_time_seconds")
}

func (c *SystemdConfig) NodeExporterFlags() (flags []string) {
	if kingpin.CommandLine.GetFlag("collector.systemd") == nil {
		return
	}

	if c.Enabled {
		flags = append(flags,
			"--collector.systemd",
			"--collector.systemd.unit-whitelist", c.UnitWhitelist,
			"--collector.systemd.unit-blacklist", c.UnitBlacklist,
		)

		// Map in boolean values
		flags = append(flags, booleanFlagMap(map[*bool]string{
			&c.EnableTaskMetrics:      "collector.systemd.enable-task-metrics",
			&c.EnableRestartsMetrics:  "collector.systemd.enable-restarts-metrics",
			&c.EnableStartTimeMetrics: "collector.systemd.enable-start-time-metrics",
		})...)
	} else {
		flags = append(flags, "--no-collector.systemd")
	}

	return
}

type VMStatConfig struct {
	Enabled bool   `yaml:"enabled"`
	Fields  string `yaml:"fields"`
}

func (c *VMStatConfig) RegisterFlagsWithPrefix(prefix string, f *flag.FlagSet) {
	f.BoolVar(&c.Enabled, prefix+"vmstat-collector.enabled", true, "Exposes statistics from /proc/vmstat. OS: Linux")
	f.StringVar(&c.Fields, prefix+"vmstat-collector.fields", "^(oom_kill|pgpg|pswp|pg.*fault).*", "Regexp of fields to return for vmstat collector.")
}

func (c *VMStatConfig) NodeExporterFlags() (flags []string) {
	if kingpin.CommandLine.GetFlag("collector.vmstat") == nil {
		return
	}

	if c.Enabled {
		flags = append(flags,
			"--collector.vmstat",
			"--collector.vmstat.fields", c.Fields,
		)
	} else {
		flags = append(flags, "--no-collector.vmstat")
	}

	return
}

type TextfileConfig struct {
	Enabled   bool   `yaml:"enabled"`
	Directory string `yaml:"directory"`
}

func (c *TextfileConfig) RegisterFlagsWithPrefix(prefix string, f *flag.FlagSet) {
	f.BoolVar(&c.Enabled, prefix+"textfile-collector.enabled", true, "enables the textfile collector")
	f.StringVar(&c.Directory, prefix+"textfil-collectore.directory", "", "directory for the textfile collector to read from")
}

func (c *TextfileConfig) NodeExporterFlags() (flags []string) {
	if kingpin.CommandLine.GetFlag("collector.textfile") == nil {
		return
	}

	if c.Enabled {
		flags = append(flags,
			"--collector.textfile",
			"--collector.textfile.directory", c.Directory,
		)
	} else {
		flags = append(flags, "--no-collector.textfile")
	}

	return
}

// MapConfigToNodeExporterFlags takes in a node_exporter Config and converts
// it to the set of flags that node_exporter usually expects when running as a
// separate binary.
func MapConfigToNodeExporterFlags(c *Config) []string {
	collectorMap := map[*bool]string{
		&c.EnableARPCollector:         "collector.arp",
		&c.EnableBCacheCollector:      "collector.bcache",
		&c.EnableBondingCollector:     "collector.bonding",
		&c.EnableBtrfsCollector:       "collector.btrfs",
		&c.EnableBoottimeCollector:    "collector.boottime",
		&c.EnableConntrackCollector:   "collector.conntrack",
		&c.EnableCPUFreqCollector:     "collector.cpufreq",
		&c.EnableEDACCollector:        "collector.edac",
		&c.EnableEntropyCollector:     "collector.entropy",
		&c.EnableExecCollector:        "collector.exec",
		&c.EnableFileFDCollector:      "collector.filefd",
		&c.EnableHWMonCollector:       "collector.hwmon",
		&c.EnableInfiniBandCollector:  "collector.infiniband",
		&c.EnableIPVSCollector:        "collector.ipvs",
		&c.EnableLoadAvgCollector:     "collector.loadavg",
		&c.EnableMDADMCollector:       "collector.mdadm",
		&c.EnableMemInfoCollector:     "collector.memoinfo",
		&c.EnableNFSCollector:         "collector.nfs",
		&c.EnableNFSDCollector:        "collector.nfsd",
		&c.EnablePressureCollector:    "collector.pressure",
		&c.EnableRAPLCollector:        "collector.rapl",
		&c.EnableSchedStatCollector:   "collector.schedstat",
		&c.EnableSockStatCollector:    "collector.sockstat",
		&c.EnableSoftNetCollector:     "collector.softnet",
		&c.EnableStatCollector:        "collector.stat",
		&c.EnableThermalZoneCollector: "collector.thermal_zone",
		&c.EnableTimeCollector:        "collector.time",
		&c.EnableTimexCollector:       "collector.timex",
		&c.EnableUDPQueuesCollector:   "collector.udp_queues",
		&c.EnableUNameCollector:       "collector.uname",
		&c.EnableXFSCollector:         "collector.xfs",
		&c.EnableZFSCollector:         "collector.zfs",
		&c.EnableBuddyinfoCollector:   "collector.buddyinfo",
		&c.EnableDevStatCollector:     "collector.devstat",
		&c.EnableDRBDCollector:        "collector.drbd",
		&c.EnableInterruptsCollector:  "collector.interrupts",
		&c.EnableKSMDCollector:        "collector.ksmd",
		&c.EnableLoginDCollector:      "collector.logind",
		&c.EnableMeminfoNUMACollector: "collector.meminfo_numa",
		&c.EnableMountStatsCollector:  "collector.mountstats",
		&c.EnableProcessesCollector:   "collector.processes",
		&c.EnableQDiscCollector:       "collector.qdisc",
		&c.EnableTCPStatCollector:     "collector.tcpstat",
		&c.EnableWiFiCollector:        "collector.wifi",
	}

	flags := booleanFlagMap(collectorMap)

	// Append collector flags from collectors that have extra options
	flags = append(flags, c.TextfileCollector.NodeExporterFlags()...)
	flags = append(flags, c.CPUCollector.NodeExporterFlags()...)
	flags = append(flags, c.DiskStatsCollector.NodeExporterFlags()...)
	flags = append(flags, c.FilesystemCollector.NodeExporterFlags()...)
	flags = append(flags, c.NTPCollector.NodeExporterFlags()...)
	flags = append(flags, c.NetclassCollector.NodeExporterFlags()...)
	flags = append(flags, c.NetdevCollector.NodeExporterFlags()...)
	flags = append(flags, c.NetstatCollector.NodeExporterFlags()...)
	flags = append(flags, c.PerfCollector.NodeExporterFlags()...)
	flags = append(flags, c.PowerSupplyCollector.NodeExporterFlags()...)
	flags = append(flags, c.RunitCollector.NodeExporterFlags()...)
	flags = append(flags, c.SupervisordCollector.NodeExporterFlags()...)
	flags = append(flags, c.SystemdCollector.NodeExporterFlags()...)
	flags = append(flags, c.VMStatCollector.NodeExporterFlags()...)

	// Non-boolean collector flags
	flags = append(flags,
		"--path.procfs", c.ProcFSPath,
		"--path.sysfs", c.SysFSPath,
		"--path.rootfs", c.RootFSPath,
	)

	return flags
}

func booleanFlagMap(m map[*bool]string) []string {
	var flags []string

	var (
		yesPrefix = "--"
		noPrefix  = "--no-"
	)

	for setting, key := range m {
		// The flag might not exist on this platform, so skip it if it's not
		// defined.
		if kingpin.CommandLine.GetFlag(key) == nil {
			continue
		}

		if *setting {
			flags = append(flags, yesPrefix+key)
		} else {
			flags = append(flags, noPrefix+key)
		}
	}

	return flags
}
