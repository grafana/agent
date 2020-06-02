package node_exporter

import (
	"flag"

	"github.com/grafana/agent/pkg/integrations/config"
	"gopkg.in/alecthomas/kingpin.v2"
)

// Config controls the node_exporter integration.
type Config struct {
	CommonConfig config.Common `yaml:",inline"`

	// Enabled enables the node_exporter integration.
	Enabled bool

	IncludeExporterMetrics bool `yaml:"include_exporter_metrics"`

	// Exposes ARP statistics from /proc/net/arp.
	//
	// OS: Linux
	EnableARPCollector bool `yaml:"enable_arp_collector"`

	// Exposes bcache statistics from /sys/fs/bcache.
	//
	// OS: Linux
	EnableBCacheCollector bool `yaml:"enable_bcache_collector"`

	// Exposes the number of configured and active slaves of Linux bonding
	// interfaces.
	//
	// OS: Linux
	EnableBondingCollector bool `yaml:"enable_bonding_collector"`

	// Exposes system boot time derived from the kern.boottime sysctl.
	//
	// OS: Darwin, Dragonfly, FreeBSD, NetBSD, OpenBSD, Solaris
	EnableBoottimeCollector bool `yaml:"enable_boottime_collector"`

	// Shows conntrack statistics (does nothing if no /proc/sys/net/netfilter/
	// present).
	//
	// OS: Linux
	EnableConntrackCollector bool `yaml:"enable_conntrack_collector"`

	// Exposes CPU statistics.
	//
	// OS: Darwin, Dragonfly, FreeBSD, Linux, Solaris
	EnableCPUCollector bool `yaml:"enable_cpu_collector"`

	// Exposes CPU frequency statistics.
	//
	// OS: Linux, Solaris
	EnableCPUFreqCollector bool `yaml:"enable_cpufreq_collector"`

	// Exposes disk I/O statistics.
	//
	// OS: Darwin, Linux, OpenBSD
	EnableDiskstatsCollector bool `yaml:"enable_diskstats_collector"`

	// Exposes error detection and correction statistics.
	//
	// OS: Linux
	EnableEDACCollector bool `yaml:"enable_edac_collector"`

	// Exposes available entropy.
	//
	// OS: Linux
	EnableEntropyCollector bool `yaml:"enable_entropy_collector"`

	// Exposes execution statistics.
	//
	// OS: Dragonfly, FreeBSD
	EnableExecCollector bool `yaml:"enable_exec_collector"`

	// Exposes file descriptor statistics from /proc/sys/fs/file-nr.
	//
	// OS: Linux
	EnableFileFDCollector bool `yaml:"enable_filefd_collector"`

	// Exposes filesystem statistics, such as disk space used.
	//
	// OS: Darwin, Dragonfly, FreeBSD, Linux, OpenBSD
	EnableFilesystemCollector bool `yaml:"enable_filesystem_collector"`

	// Exposes hardware monitoring and sensor data from /sys/class/hwmon.
	//
	// OS: Linux
	EnableHWMonCollector bool `yaml:"enable_hwmon_collector"`

	// Exposes network statistics specific to InfiniBand and Intel OmniPath
	// configurations.
	//
	// OS: Linux
	EnableInfiniBandCollector bool `yaml:"enable_infiniband_collector"`

	// Exposes IPVS status from /proc/net/ip_vs and stats from
	// /proc/net/ip_vs_stats.
	//
	// OS: Linux
	EnableIPVSCollector bool `yaml:"enable_ipvs_collector"`

	// Exposes load average.
	//
	// OS: Darwin, Dragonfly, FreeBSD, Linux, NetBSD, OpenBSD, Solaris
	EnableLoadAvgCollector bool `yaml:"enable_load_avg_collector"`

	// Exposes statistics about devices in /proc/mdstat (does nothing if no
	// /proc/mdstat present).
	//
	// OS: Linux
	EnableMDADMCollector bool `yaml:"enable_mdadm_collector"`

	// Exposes memory statistics.
	//
	// OS: Darwin, Dragonfly, FreeBSD, Linux, OpenBSD
	EnableMemInfoCollector bool `yaml:"enable_meminfo_collector"`

	// Exposes network interface info from /sys/class/net.
	//
	// OS: Linux
	EnableNetClassCollector bool `yaml:"enable_netclass_collector"`

	// Exposes network interface statistics such as bytes transferred.
	//
	// OS: Darwin, Dragonfly, FreeBSD, Linux, OpenBSD
	EnableNetDevCollector bool `yaml:"enable_netdev_collector"`

	// Exposes network statistics from /proc/net/netstat. This is the same
	// information as netstat -s.
	//
	// OS: Linux
	EnableNetStatCollector bool `yaml:"enable_netstat_collector"`

	// Exposes NFS client statistics from /proc/net/rpc/nfs. This is the same
	// information as nfsstat -c.
	//
	// OS: Linux
	EnableNFSCollector bool `yaml:"enable_nfs_collector"`

	// Exposes NFS kernel server statistics from /proc/net/rpc/nfsd. This is the
	// same information as nfsstat -s.
	//
	// OS: Linux
	EnableNFSDCollector bool `yaml:"enable_nfsd_collector"`

	// Exposes pressure stall statistics from /proc/pressure/.
	//
	// OS: Linux (kernel 4.20+ and/or CONFIG_PSI)
	EnablePressureCollector bool `yaml:"enable_pressure_collector"`

	// Exposes various statistics from /sys/class/powercap.
	//
	// OS: Linux
	EnableRAPLCollector bool `yaml:"enable_rapl_collector"`

	// Exposes task scheduler statistics from /proc/schedstat.
	//
	// OS: Linux
	EnableSchedStatCollector bool `yaml:"enable_schedstat_collector"`

	// Exposes various statistics from /proc/net/sockstat.
	//
	// OS: Linux
	EnableSockStatCollector bool `yaml:"enable_sockstat_collector"`

	// Exposes statistics from /proc/net/softnet_stat.
	//
	// OS: Linux
	EnableSoftNetCollector bool `yaml:"enable_softnet_collector"`

	// Exposes various statistics from /proc/stat. This includes boot time, forks
	// and interrupts.
	//
	// OS: Linux
	EnableStatCollector bool `yaml:"enable_stat_collector"`

	// Exposes thermal zone & cooling device statistics from /sys/class/thermal.
	//
	// OS: Linux
	EnableThermalZoneCollector bool `yaml:"enable_thermal_zone_collector"`

	// Exposes the current system time.
	//
	// OS: any
	EnableTimeCollector bool `yaml:"enable_time_collector"`

	// Exposes selected adjtimex(2) system call stats.
	//
	// OS: Linux
	EnableTimexCollector bool `yaml:"enable_timex_collector"`

	// Exposes UDP total lengths of the rx_queue and tx_queue from /proc/net/udp
	// and /proc/net/udp6.
	//
	// OS: Linux
	EnableUDPQueuesCollector bool `yaml:"enable_udp_queues_collector"`

	// Exposes system information as provided by the uname system call.
	//
	// OS: Darwin, FreeBSD, Linux, OpenBSD
	EnableUNameCollector bool `yaml:"enable_uname_collector"`

	// Exposes statistics from /proc/vmstat.
	//
	// OS: Linux
	EnableVMStatCollector bool `yaml:"enable_vmstat_collector"`

	// Exposes XFS runtime statistics.
	//
	// OS: Linux (kernel 4.4+)
	EnableXFSCollector bool `yaml:"enable_xfs_collector"`

	// Exposes ZFS performance statistics.
	//
	// OS: Linux, Solaris
	EnableZFSCollector bool `yaml:"enable_zfs_collector"`

	// Exposes statistics read from local disk.
	//
	// OS: any
	EnableTextfileCollector TextfileConfig `yaml:"textfile_collector"`

	// Exposes statistics of memory fragments as reported by /proc/buddyinfo.
	//
	// OS: Linux
	EnableBuddyinfoCollector bool `yaml:"enable_buddyinfo_collector"`

	// Exposes device statistics.
	//
	// OS: Dragonfly, FreeBSD
	EnableDevStatCollector bool `yaml:"enable_devstat_collector"`

	// Exposes Distributed Replicated Block Device statistics (to version 8.4).
	//
	// OS: Linux
	EnableDRBDCollector bool `yaml:"enable_drbd_collector"`

	// Exposes detailed interrupts statistics.
	//
	// OS: Linux, OpenBSD
	EnableInterruptsCollector bool `yaml:"enable_interrupts_collector"`

	// Exposes kernel and system statistics from /sys/kernel/mm/ksm.
	//
	// OS: Linux
	EnableKSMDCollector bool `yaml:"enable_ksmd_collector"`

	// Exposes session counts from logind.
	//
	// OS: Linux
	EnableLoginDCollector bool `yaml:"enable_logind_collector"`

	// Exposes memory statistics from /proc/meminfo_numa.
	//
	// OS: Linux
	EnableMeminfoNUMACollector bool `yaml:"enable_meminfo_numa_collector"`

	// Exposes filesystem statistics from /proc/self/mountstats. Exposes detailed
	// NFS client statistics.
	//
	// OS: Linux
	EnableMountStatsCollector bool `yaml:"enable_mountstats_collector"`

	// Exposes local NTP daemon helath to check time.
	//
	// OS: any
	EnableNTPCollector bool `yaml:"enable_ntp_collector"`

	// Exposes aggregate process statistics from /proc.
	//
	// OS: Linux
	EnableProcessesCollector bool `yaml:"enable_processes_collector"`

	// Exposes queuing discipline statistics.
	//
	// OS: Linux
	EnableQDiscCollector bool `yaml:"enable_qdisc_collector"`

	// Exposes service status from runit.
	//
	// OS: any
	EnableRunitCollector bool `yaml:"enable_runit_collector"`

	// Exposes service status from supervisord.
	//
	// OS: any
	EnableSupervisorDCollector bool `yaml:"enable_supervisord_collector"`

	// Exposes service and system status from systemd.
	//
	// OS: Linux
	EnableSystemDCollector bool `yaml:"enable_systemd_collector"`

	// Exposes TCP connection status information from /proc/net/tcp and
	// /proc/net/tcp6. (Warning: the current version has potential performance
	// issues in high load situations).
	//
	// OS: Linux
	EnableTCPStatCollector bool `yaml:"enable_tcpstat_collector"`

	// Exposes WiFi device and station statistics.
	//
	// OS: Linux
	EnableWiFiCollector bool `yaml:"enable_wifi_collector"`

	// Exposes perf based metrics (Warning: Metrics are dependent on kernel
	// configuration and settings)
	//
	// OS: Linux
	EnablePerfCollector bool `yaml:"enable_perf_collector"`
}

func (c *Config) RegisterFlagsWithPrefix(prefix string, f *flag.FlagSet) {
	c.CommonConfig.RegisterFlagsWithPrefix(prefix+"node_exporter.", f)
	c.EnableTextfileCollector.RegisterFlagsWithPrefix(prefix+"node_exporter.", f)

	f.BoolVar(&c.Enabled, prefix+"node_exporter.enabled", false, "enable the node_exporter integration collect metrics from the host Linux system")
	f.BoolVar(&c.IncludeExporterMetrics, prefix+"node_exporter.include-exporter-metrics", false, "include metrics on the integration itself")

	f.BoolVar(&c.EnableARPCollector, prefix+"node_exporter.enable_arp_collector", true, "Exposes ARP statistics from /proc/net/arp. OS: Linux")
	f.BoolVar(&c.EnableBCacheCollector, prefix+"node_exporter.enable_bcache_collector", true, "Exposes bcache statistics from /sys/fs/bcache. OS: Linux")
	f.BoolVar(&c.EnableBondingCollector, prefix+"node_exporter.enable_bonding_collector", true, "Exposes the number of configured and active slaves of Linux bonding interfaces. OS: Linux")
	f.BoolVar(&c.EnableBoottimeCollector, prefix+"node_exporter.enable_boottime_collector", true, "Exposes system boot time derived from the kern.boottime sysctl. OS: Darwin, Dragonfly, FreeBSD, NetBSD, OpenBSD, Solaris")
	f.BoolVar(&c.EnableConntrackCollector, prefix+"node_exporter.enable_conntrack_collector", true, "Shows conntrack statistics (does nothing if no /proc/sys/net/netfilter/ present). OS: Linux")
	f.BoolVar(&c.EnableCPUCollector, prefix+"node_exporter.enable_cpu_collector", true, "Exposes CPU statistics. OS: Darwin, Dragonfly, FreeBSD, Linux, Solaris")
	f.BoolVar(&c.EnableCPUFreqCollector, prefix+"node_exporter.enable_cpufreq_collector", true, "Exposes CPU frequency statistics. OS: Linux, Solaris")
	f.BoolVar(&c.EnableDiskstatsCollector, prefix+"node_exporter.enable_diskstats_collector", true, "Exposes disk I/O statistics. OS: Darwin, Linux, OpenBSD")
	f.BoolVar(&c.EnableEDACCollector, prefix+"node_exporter.enable_edac_collector", true, "Exposes error detection and correction statistics. OS: Linux")
	f.BoolVar(&c.EnableEntropyCollector, prefix+"node_exporter.enable_entropy_collector", true, "Exposes available entropy. OS: Linux")
	f.BoolVar(&c.EnableExecCollector, prefix+"node_exporter.enable_exec_collector", true, "Exposes execution statistics. OS: Dragonfly, FreeBSD")
	f.BoolVar(&c.EnableFileFDCollector, prefix+"node_exporter.enable_filefd_collector", true, "Exposes file descriptor statistics from /proc/sys/fs/file-nr. OS: Linux")
	f.BoolVar(&c.EnableFilesystemCollector, prefix+"node_exporter.enable_filesystem_collector", true, "Exposes filesystem statistics, such as disk space used. OS: Darwin, Dragonfly, FreeBSD, Linux, OpenBSD")
	f.BoolVar(&c.EnableHWMonCollector, prefix+"node_exporter.enable_hwmon_collector", true, "Exposes hardware monitoring and sensor data from /sys/class/hwmon. OS: Linux")
	f.BoolVar(&c.EnableInfiniBandCollector, prefix+"node_exporter.enable_infiniband_collector", true, "Exposes network statistics specific to InfiniBand and Intel OmniPath configurations. OS: Linux")
	f.BoolVar(&c.EnableIPVSCollector, prefix+"node_exporter.enable_ipvs_collector", true, "Exposes IPVS status from /proc/net/ip_vs and stats from /proc/net/ip_vs_stats. OS: Linux")
	f.BoolVar(&c.EnableLoadAvgCollector, prefix+"node_exporter.enable_load_avg_collector", true, "Exposes load average. OS: Darwin, Dragonfly, FreeBSD, Linux, NetBSD, OpenBSD, Solaris")
	f.BoolVar(&c.EnableMDADMCollector, prefix+"node_exporter.enable_mdadm_collector", true, "Exposes statistics about devices in /proc/mdstat (does nothing if no /proc/mdstat present). OS: Linux")
	f.BoolVar(&c.EnableMemInfoCollector, prefix+"node_exporter.enable_meminfo_collector", true, "Exposes memory statistics. OS: Darwin, Dragonfly, FreeBSD, Linux, OpenBSD")
	f.BoolVar(&c.EnableNetClassCollector, prefix+"node_exporter.enable_netclass_collector", true, "Exposes network interface info from /sys/class/net. OS: Linux")
	f.BoolVar(&c.EnableNetDevCollector, prefix+"node_exporter.enable_netdev_collector", true, "Exposes network interface statistics such as bytes transferred. OS: Darwin, Dragonfly, FreeBSD, Linux, OpenBSD")
	f.BoolVar(&c.EnableNetStatCollector, prefix+"node_exporter.enable_netstat_collector", true, "Exposes network statistics from /proc/net/netstat. This is the same information as netstat -s. OS: Linux")
	f.BoolVar(&c.EnableNFSCollector, prefix+"node_exporter.enable_nfs_collector", true, "Exposes NFS client statistics from /proc/net/rpc/nfs. This is the same information as nfsstat -c. OS: Linux")
	f.BoolVar(&c.EnableNFSDCollector, prefix+"node_exporter.enable_nfsd_collector", true, "Exposes NFS kernel server statistics from /proc/net/rpc/nfsd. This is the same information as nfsstat -s. OS: Linux")
	f.BoolVar(&c.EnablePressureCollector, prefix+"node_exporter.enable_pressure_collector", true, "Exposes pressure stall statistics from /proc/pressure/. OS: Linux (kernel 4.20+ and/or CONFIG_PSI)")
	f.BoolVar(&c.EnableRAPLCollector, prefix+"node_exporter.enable_rapl_collector", true, "Exposes various statistics from /sys/class/powercap. OS: Linux")
	f.BoolVar(&c.EnableSchedStatCollector, prefix+"node_exporter.enable_schedstat_collector", true, "Exposes task scheduler statistics from /proc/schedstat. OS: Linux")
	f.BoolVar(&c.EnableSockStatCollector, prefix+"node_exporter.enable_sockstat_collector", true, "Exposes various statistics from /proc/net/sockstat. OS: Linux")
	f.BoolVar(&c.EnableSoftNetCollector, prefix+"node_exporter.enable_softnet_collector", true, "Exposes statistics from /proc/net/softnet_stat. OS: Linux")
	f.BoolVar(&c.EnableStatCollector, prefix+"node_exporter.enable_stat_collector", true, "Exposes various statistics from /proc/stat. This includes boot time, forks and interrupts. OS: Linux")
	f.BoolVar(&c.EnableThermalZoneCollector, prefix+"node_exporter.enable_thermal_zone_collector", true, "Exposes thermal zone & cooling device statistics from /sys/class/thermal. OS: Linux")
	f.BoolVar(&c.EnableTimeCollector, prefix+"node_exporter.enable_time_collector", true, "Exposes the current system time. OS: any")
	f.BoolVar(&c.EnableTimexCollector, prefix+"node_exporter.enable_timex_collector", true, "Exposes selected adjtimex(2) system call stats. OS: Linux")
	f.BoolVar(&c.EnableUDPQueuesCollector, prefix+"node_exporter.enable_udp_queues_collector", true, "Exposes UDP total lengths of the rx_queue and tx_queue from /proc/net/udp and /proc/net/udp6. OS: Linux")
	f.BoolVar(&c.EnableUNameCollector, prefix+"node_exporter.enable_uname_collector", true, "Exposes system information as provided by the uname system call. OS: Darwin, FreeBSD, Linux, OpenBSD")
	f.BoolVar(&c.EnableVMStatCollector, prefix+"node_exporter.enable_vmstat_collector", true, "Exposes statistics from /proc/vmstat. OS: Linux")
	f.BoolVar(&c.EnableXFSCollector, prefix+"node_exporter.enable_xfs_collector", true, "Exposes XFS runtime statistics. OS: Linux (kernel 4.4+)")
	f.BoolVar(&c.EnableZFSCollector, prefix+"node_exporter.enable_zfs_collector", true, "Exposes ZFS performance statistics. OS: Linux, Solaris")

	// Disabled by default
	f.BoolVar(&c.EnableBuddyinfoCollector, prefix+"node_exporter.enable_buddyinfo_collector", false, "Exposes statistics of memory fragments as reported by /proc/buddyinfo. OS: Linux")
	f.BoolVar(&c.EnableDevStatCollector, prefix+"node_exporter.enable_devstat_collector", false, "Exposes device statistics. OS: Dragonfly, FreeBSD")
	f.BoolVar(&c.EnableDRBDCollector, prefix+"node_exporter.enable_drbd_collector", false, "Exposes Distributed Replicated Block Device statistics (to version 8.4). OS: Linux")
	f.BoolVar(&c.EnableInterruptsCollector, prefix+"node_exporter.enable_interrupts_collector", false, "Exposes detailed interrupts statistics. OS: Linux, OpenBSD")
	f.BoolVar(&c.EnableKSMDCollector, prefix+"node_exporter.enable_ksmd_collector", false, "Exposes kernel and system statistics from /sys/kernel/mm/ksm. OS: Linux")
	f.BoolVar(&c.EnableLoginDCollector, prefix+"node_exporter.enable_logind_collector", false, "Exposes session counts from logind. OS: Linux")
	f.BoolVar(&c.EnableMeminfoNUMACollector, prefix+"node_exporter.enable_meminfo_numa_collector", false, "Exposes memory statistics from /proc/meminfo_numa. OS: Linux")
	f.BoolVar(&c.EnableMountStatsCollector, prefix+"node_exporter.enable_mountstats_collector", false, "Exposes filesystem statistics from /proc/self/mountstats. Exposes detailed NFS client statistics. OS: Linux")
	f.BoolVar(&c.EnableNTPCollector, prefix+"node_exporter.enable_ntp_collector", false, "Exposes local NTP daemon helath to check time. OS: any")
	f.BoolVar(&c.EnableProcessesCollector, prefix+"node_exporter.enable_processes_collector", false, "Exposes aggregate process statistics from /proc. OS: Linux")
	f.BoolVar(&c.EnableQDiscCollector, prefix+"node_exporter.enable_qdisc_collector", false, "Exposes queuing discipline statistics. OS: Linux")
	f.BoolVar(&c.EnableRunitCollector, prefix+"node_exporter.enable_runit_collector", false, "Exposes service status from runit. OS: any")
	f.BoolVar(&c.EnableSupervisorDCollector, prefix+"node_exporter.enable_supervisord_collector", false, "Exposes service status from supervisord. OS: any")
	f.BoolVar(&c.EnableSystemDCollector, prefix+"node_exporter.enable_systemd_collector", false, "Exposes service and system status from systemd. OS: Linux")
	f.BoolVar(&c.EnableTCPStatCollector, prefix+"node_exporter.enable_tcpstat_collector", false, "Exposes TCP connection status information from /proc/net/tcp and /proc/net/tcp6. (Warning: the current version has potential performance issues in high load situations). OS: Linux")
	f.BoolVar(&c.EnableWiFiCollector, prefix+"node_exporter.enable_wifi_collector", false, "Exposes WiFi device and station statistics. OS: Linux")
	f.BoolVar(&c.EnablePerfCollector, prefix+"node_exporter.enable_perf_collector", false, "Exposes perf based metrics (Warning: Metrics are dependent on kernel configuration and settings). OS: Linux")
}

// TODO(rfratto): missing flags
// --collector.cpu.info
// --collector.diskstats.ignored-devices
// --collector.filesystem.ignored-mount-points
// --collector.filesystem.ignored-fs-types
// --collector.netclass.ignored-devices
// --collector.netdev.device-blacklist
// --collector.netdev.device-whitelist
// --collector.netstat.fields
// --collector.ntp.server
// --collector.ntp-protocol-version
// --collector.ntp-server-is-local
// --collector.ntp.ip-ttl
// --collector.ntp.max-distance
// --collector.ntp.local-offset-tolerance
// --path.procfs
// --path.sysfs
// --path.rootfs
// --collector.perf.cpus
// --collector.perf.tracepoint
// --collector.powersupply.ignored-supplies
// --collector.runit.servicedir
// --collector.supervisord.url
// --collector.systemd.unit-whitelist
// --collector.systemd.unit-blacklist
// --collector.systemd.enable-task-metrics
// --collector.systemd.enable-restarts-metrics
// --collector.systemd.enable-start-time-metrics
// --collector.vmstat.fields
// -

type CPUConfig struct {
	Enabled bool `yaml:"enabled"`                // --collector.cpu
	Info    bool `yaml:"enable_cpu_info_metric"` // --collector.cpu.info
}

type DiskStatsConfig struct {
	Enabled        bool   `yaml:"enabled"`         // --collector.diskstats
	IgnoredDevices string `yaml:"ignored_devices"` // --collector.diskstats.ignored-devices
}

type FilesystemConfig struct {
	Enabled bool `yaml:"enabled"` // --collector.filesystem

}

type TextfileConfig struct {
	Enabled   bool   `yaml:"enabled"`
	Directory string `yaml:"directory"`
}

func (c *TextfileConfig) RegisterFlagsWithPrefix(prefix string, f *flag.FlagSet) {
	f.BoolVar(&c.Enabled, prefix+"textfile.enabled", true, "enables the textfile collector")
	f.StringVar(&c.Directory, prefix+"textfile.directory", "", "directory for the textfile collector to read from")
}

// MapConfigToNodeExporterFlags takes in a node_exporter Config and converts
// it to the set of flags that node_exporter usually expects when running as a
// separate binary.
func MapConfigToNodeExporterFlags(c *Config) []string {
	var flags []string

	// TODO(rfratto): missing collectors
	//
	// btrfs
	// buddyinfo
	// powersupplyclass
	//
	collectorMap := map[*bool]string{
		&c.EnableARPCollector:         "arp",
		&c.EnableBCacheCollector:      "bcache",
		&c.EnableBondingCollector:     "bonding",
		&c.EnableBoottimeCollector:    "boottime",
		&c.EnableConntrackCollector:   "conntrack",
		&c.EnableCPUCollector:         "cpu",
		&c.EnableCPUFreqCollector:     "cpufreq",
		&c.EnableDiskstatsCollector:   "diskstats",
		&c.EnableEDACCollector:        "edac",
		&c.EnableEntropyCollector:     "entropy",
		&c.EnableExecCollector:        "exec",
		&c.EnableFileFDCollector:      "filefd",
		&c.EnableFilesystemCollector:  "filesystem",
		&c.EnableHWMonCollector:       "hwmon",
		&c.EnableInfiniBandCollector:  "infiniband",
		&c.EnableIPVSCollector:        "ipvs",
		&c.EnableLoadAvgCollector:     "loadavg",
		&c.EnableMDADMCollector:       "mdadm",
		&c.EnableMemInfoCollector:     "memoinfo",
		&c.EnableNetClassCollector:    "netclass",
		&c.EnableNetDevCollector:      "netdev",
		&c.EnableNetStatCollector:     "netstat",
		&c.EnableNFSCollector:         "nfs",
		&c.EnableNFSDCollector:        "nfsd",
		&c.EnablePressureCollector:    "pressure",
		&c.EnableRAPLCollector:        "rapl",
		&c.EnableSchedStatCollector:   "schedstat",
		&c.EnableSockStatCollector:    "sockstat",
		&c.EnableSoftNetCollector:     "softnet",
		&c.EnableStatCollector:        "stat",
		&c.EnableThermalZoneCollector: "thermal_zone",
		&c.EnableTimeCollector:        "time",
		&c.EnableTimexCollector:       "timex",
		&c.EnableUDPQueuesCollector:   "udp_queues",
		&c.EnableUNameCollector:       "uname",
		&c.EnableVMStatCollector:      "vmstat",
		&c.EnableXFSCollector:         "xfs",
		&c.EnableZFSCollector:         "zfs",
		&c.EnableBuddyinfoCollector:   "buddyinfo",
		&c.EnableDevStatCollector:     "devstat",
		&c.EnableDRBDCollector:        "drbd",
		&c.EnableInterruptsCollector:  "interrupts",
		&c.EnableKSMDCollector:        "ksmd",
		&c.EnableLoginDCollector:      "logind",
		&c.EnableMeminfoNUMACollector: "meminfo_numa",
		&c.EnableMountStatsCollector:  "mountstats",
		&c.EnableNTPCollector:         "ntp",
		&c.EnableProcessesCollector:   "processes",
		&c.EnableQDiscCollector:       "qdisc",
		&c.EnableRunitCollector:       "runit",
		&c.EnableSupervisorDCollector: "supervisord",
		&c.EnableSystemDCollector:     "systemd",
		&c.EnableTCPStatCollector:     "tcpstat",
		&c.EnableWiFiCollector:        "wifi",
		&c.EnablePerfCollector:        "perf",
	}

	var (
		enabledPrefix  = "--collector."
		disabledPrefix = "--no-collector."
	)

	for setting, key := range collectorMap {
		// The flag might not exist on this platform, so skip it if it's not
		// defined.
		if kingpin.CommandLine.GetFlag("collector."+key) == nil {
			continue
		}

		if *setting == false {
			flags = append(flags, disabledPrefix+key)
		} else {
			flags = append(flags, enabledPrefix+key)
		}
	}

	// Handle textfile, which is a special case since it has a separate config
	// struct.
	if c.EnableTextfileCollector.Enabled {
		flags = append(flags,
			enabledPrefix+"textfile",
			"--collector.textfile.directory",
			c.EnableTextfileCollector.Directory,
		)
	} else {
		flags = append(flags, disabledPrefix+"textfile")
	}

	return flags
}
