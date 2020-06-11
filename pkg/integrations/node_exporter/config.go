package node_exporter //nolint:golint

import (
	"flag"
	"fmt"
	"time"

	"github.com/cortexproject/cortex/pkg/util/flagext"
	"github.com/grafana/agent/pkg/integrations/config"
	"gopkg.in/alecthomas/kingpin.v2"
)

// Config controls the node_exporter integration.
type Config struct {
	CommonConfig config.Common `yaml:",inline"`

	// Enabled enables the node_exporter integration.
	Enabled bool `yaml:"enabled"`

	IncludeExporterMetrics bool `yaml:"include_exporter_metrics"`

	ProcFSPath string `yaml:"procfs_path"`
	SysFSPath  string `yaml:"sysfs_path"`
	RootFSPath string `yaml:"rootfs_path"`

	// Collectors to mark as enabled
	EnableCollectors flagext.StringSlice `yaml:"enable_collectors"`

	// Collectors to mark as disabled
	DisableCollectors flagext.StringSlice `yaml:"disable_collectors"`

	// Overrides the default set of enabled collectors with the collectors
	// listed.
	SetCollectors flagext.StringSlice `yaml:"set_collectors"`

	// Collector-specific config options
	CPUEnableCPUInfo              bool                `yaml:"enable_cpu_info_metric"`
	DiskStatsIgnoredDevices       string              `yaml:"diskstats_ignored_devices"`
	FilesystemIgnoredMountPoints  string              `yaml:"filesystem_ignored_mount_points"`
	FilesystemIgnoredFSTypes      string              `yaml:"filesystem_ignored_fs_types"`
	NetclassIgnoredDevices        string              `yaml:"netclass_ignored_devices"`
	NetdevDeviceBlacklist         string              `yaml:"netdev_device_blacklist"`
	NetdevDeviceWhitelist         string              `yaml:"netdev_device_whitelist"`
	NetstatFields                 string              `yaml:"netstat_fields"`
	NTPServer                     string              `yaml:"ntp_server"`
	NTPProtocolVersion            int                 `yaml:"ntp_protocol_version"`
	NTPServerIsLocal              bool                `yaml:"ntp_server_is_local"`
	NTPIPTTL                      int                 `yaml:"ntp_ip_ttl"`
	NTPMaxDistance                time.Duration       `yaml:"ntp_max_distance"`
	NTPLocalOffsetTolerance       time.Duration       `yaml:"ntp_local_offset_tolerance"`
	PerfCPUS                      string              `yaml:"perf_cpus"`
	PerfTracepoint                flagext.StringSlice `yaml:"perf_tracepoint"`
	PowersupplyIgnoredSupplies    string              `yaml:"powersupply_ignored_supplies"`
	RunitServiceDir               string              `yaml:"runit_service_dir"`
	SupervisordURL                string              `yaml:"supervisord_url"`
	SystemdUnitWhitelist          string              `yaml:"systemd_unit_whitelist"`
	SystemdUnitBlacklist          string              `yaml:"systemd_unit_blacklist"`
	SystemdEnableTaskMetrics      bool                `yaml:"systemd_enable_task_metrics"`
	SystemdEnableRestartsMetrics  bool                `yaml:"systemd_enable_restarts_metrics"`
	SystemdEnableStartTimeMetrics bool                `yaml:"systemd_enable_start_time_metrics"`
	VMStatFields                  string              `yaml:"vmstat_fields"`
	TextfileDirectory             string              `yaml:"textfile_directory"`
}

func (c *Config) RegisterFlagsWithPrefix(prefix string, f *flag.FlagSet) {
	prefix = prefix + "node_exporter."

	c.CommonConfig.RegisterFlagsWithPrefix(prefix, f)

	f.BoolVar(&c.Enabled, prefix+"enabled", false, "enable the node_exporter integration to collect metrics from the host Linux system")
	f.BoolVar(&c.IncludeExporterMetrics, prefix+"include-exporter-metrics", false, "include metrics on the integration itself")

	f.Var(&c.SetCollectors, prefix+"set-collectors",
		"collectors to set as enabled. any collector not defined here will be marked as disabled. Pass multiple times to add more collectors.")
	f.Var(&c.EnableCollectors, prefix+"enable-collectors",
		"collectors to enable in addition to those already enabled. Pass multiple times to enable more collectors.")
	f.Var(&c.DisableCollectors, prefix+"disable-collectors",
		"collectors to disable from the set of those already enabled. Pass multiple times to disable more collectors.")

	f.StringVar(&c.ProcFSPath, prefix+"procfs-path", "/proc", "procfs mountpoint")
	f.StringVar(&c.SysFSPath, prefix+"sysfs-path", "/sys", "sysfs mountpoint")
	f.StringVar(&c.SysFSPath, prefix+"rootfs-path", "/", "rootfs mountpoint")

	// Flags for controlling collectors
	f.BoolVar(&c.CPUEnableCPUInfo, prefix+"cpu-enable-cpu_info-metric",
		false,
		"Enable metric cpu_info for the cpu collector")
	f.StringVar(&c.DiskStatsIgnoredDevices, prefix+"diskstats-ignored-devices",
		"^(ram|loop|fd|(h|s|v|xv)d[a-z]|nvme\\d+n\\d+p)\\d+$",
		"Regexp of devices to ignore for diskstats.")
	f.StringVar(&c.FilesystemIgnoredMountPoints, prefix+"filesystem-ignored-mount-points",
		"^/(dev|proc|sys|var/lib/docker/.+)($|/)",
		"Regexp of mount points to ignore for filesystem collector.")
	f.StringVar(&c.FilesystemIgnoredFSTypes, prefix+"filesystem-ignored-fs-types",
		"^(autofs|binfmt_misc|bpf|cgroup2?|configfs|debugfs|devpts|devtmpfs|fusectl|hugetlbfs|iso9660|mqueue|nsfs|overlay|proc|procfs|pstore|rpc_pipefs|securityfs|selinuxfs|squashfs|sysfs|tracefs)$",
		"Regexmp of filesystem types to ignore for filesystem collector.")
	f.StringVar(&c.NetclassIgnoredDevices, prefix+"netclass-ignored-devices",
		"^$",
		"Regexp of net devices to ignore for netclass collector.")
	f.StringVar(&c.NetdevDeviceBlacklist, prefix+"netdev-device-blacklist",
		"",
		"Regexp of net devices to blacklist (mutually exclusive to device-whitelist)")
	f.StringVar(&c.NetdevDeviceWhitelist, prefix+"netdev-device-whitelist",
		"",
		"Regexp of net devices to whitelist (mutually exclusive to device-blacklist)")
	f.StringVar(&c.NetstatFields, prefix+"netstat-fields",
		"^(.*_(InErrors|InErrs)|Ip_Forwarding|Ip(6|Ext)_(InOctets|OutOctets)|Icmp6?_(InMsgs|OutMsgs)|TcpExt_(Listen.*|Syncookies.*|TCPSynRetrans)|Tcp_(ActiveOpens|InSegs|OutSegs|PassiveOpens|RetransSegs|CurrEstab)|Udp6?_(InDatagrams|OutDatagrams|NoPorts|RcvbufErrors|SndbufErrors))$",
		"Regexp of fields to return for netstat collector.")
	f.StringVar(&c.NTPServer, prefix+"ntp-server",
		"127.0.0.1",
		"NTP server to use for ntp collector")
	f.IntVar(&c.NTPProtocolVersion, prefix+"ntp-protocol-version",
		4,
		"NTP protocol version")
	f.BoolVar(&c.NTPServerIsLocal, prefix+"ntp-server-is-local",
		false,
		"Certify that collector.ntp.server address is not a public ntp server")
	f.IntVar(&c.NTPIPTTL, prefix+"ntp-ip-ttl",
		1,
		"IP TTL to use while sending NTP query")
	f.DurationVar(&c.NTPMaxDistance, prefix+"ntp-max-distance",
		time.Microsecond*3466080,
		"Max accumulated distance to the root")
	f.DurationVar(&c.NTPLocalOffsetTolerance, prefix+"ntp-local-offset-tolerance",
		time.Millisecond,
		"Offset between local clock and local ntpd time to tolerate")
	f.StringVar(&c.PerfCPUS, prefix+"perf-cpus",
		"",
		"List of CPUs from which perf metrics should be collected")
	f.Var(&c.PerfTracepoint, prefix+"perf-tracepoint",
		"perf tracepoint that should be collected for the perf collector")
	f.StringVar(&c.PowersupplyIgnoredSupplies, prefix+"powersupply-ignored-supplies",
		"^$",
		"Regexp of power supplies to ignore for powersupplyclass collector.")
	f.StringVar(&c.RunitServiceDir, prefix+"runit-service-dir",
		"/etc/service",
		"Path to runit service directory.")
	f.StringVar(&c.SupervisordURL, prefix+"supervisord-url",
		"http://localhost:9001/RPC2",
		"XML RPC endpoint")
	f.StringVar(&c.SystemdUnitWhitelist, prefix+"systemd-unit-whitelist",
		".+",
		"Regexp of systemd units to whitelist. Units must both match whitelist and not match blacklist to be included.")
	f.StringVar(&c.SystemdUnitBlacklist, prefix+"systemd-unit-blacklist",
		".+\\.(automount|device|mount|scope|slice)",
		"Regexp of systemd units to blacklist. Units must both match whitelist and not match blacklist to be included.")
	f.BoolVar(&c.SystemdEnableTaskMetrics, prefix+"systemd-enable-task-metrics",
		false,
		"Enables service unit tasks metrics unit_tasks_current and unit_tasks_max")
	f.BoolVar(&c.SystemdEnableRestartsMetrics, prefix+"systemd-enable-restarts-metrics",
		false,
		"Enables service unit metric service_restart_total")
	f.BoolVar(&c.SystemdEnableStartTimeMetrics, prefix+"systemd-enable-start-time-metrics",
		false,
		"Enables service unit metric unit_start_time_seconds")
	f.StringVar(&c.VMStatFields, prefix+"vmstat-fields",
		"^(oom_kill|pgpg|pswp|pg.*fault).*",
		"Regexp of fields to return for vmstat collector.")
	f.StringVar(&c.TextfileDirectory, prefix+"textfile-directory",
		"",
		"directory for the textfile collector to read from")

}

// MapConfigToNodeExporterFlags takes in a node_exporter Config and converts
// it to the set of flags that node_exporter usually expects when running as a
// separate binary.
func MapConfigToNodeExporterFlags(c *Config) []string {
	collectors := make(map[string]CollectorState, len(Collectors))
	for k, v := range Collectors {
		collectors[k] = v
	}

	// Override the set of defaults with the provided set of collectors if
	// set_collectors has at least one element in it.
	if len(c.SetCollectors) != 0 {
		customDefaults := map[string]struct{}{}
		for _, c := range c.SetCollectors {
			customDefaults[c] = struct{}{}
		}

		for k := range collectors {
			_, shouldEnable := customDefaults[k]
			if shouldEnable {
				collectors[k] = CollectorStateEnabled
			} else {
				collectors[k] = CollectorStateDisabled
			}
		}
	}

	// Explicitly disable/enable specific collectors
	for _, c := range c.DisableCollectors {
		collectors[c] = CollectorStateDisabled
	}
	for _, c := range c.EnableCollectors {
		collectors[c] = CollectorStateEnabled
	}

	DisableUnavailableCollectors(collectors)
	flags := MapCollectorsToFlags(collectors)

	flags = append(flags,
		"--path.procfs", c.ProcFSPath,
		"--path.sysfs", c.SysFSPath,
		"--path.rootfs", c.RootFSPath,
	)

	if collectors[CollectorCPU] {
		flags = append(flags, booleanFlagMap(map[*bool]string{
			&c.CPUEnableCPUInfo: "collector.cpu.info",
		})...)
	}

	if collectors[CollectorDiskstats] {
		flags = append(flags, "--collector.diskstats.ignored-devices", c.DiskStatsIgnoredDevices)
	}

	if collectors[CollectorFilesystem] {
		flags = append(flags,
			"--collector.filesystem.ignored-mount-points", c.FilesystemIgnoredMountPoints,
			"--collector.filesystem.ignored-fs-types", c.FilesystemIgnoredFSTypes,
		)
	}

	if collectors[CollectorNetclass] {
		flags = append(flags, "--collector.netclass.ignored-devices", c.NetclassIgnoredDevices)
	}

	if collectors[CollectorNetdev] {
		flags = append(flags,
			"--collector.netdev.device-blacklist", c.NetdevDeviceBlacklist,
			"--collector.netdev.device-whitelist", c.NetdevDeviceWhitelist,
		)
	}

	if collectors[CollectorNetstat] {
		flags = append(flags, "--collector.netstat.fields", c.NetstatFields)
	}

	if collectors[CollectorNTP] {
		flags = append(flags,
			"--collector.ntp.server", c.NTPServer,
			"--collector.ntp.protocol-version", fmt.Sprintf("%d", c.NTPProtocolVersion),
			"--collector.ntp.ip-ttl", fmt.Sprintf("%d", c.NTPIPTTL),
			"--collector.ntp.max-distance", c.NTPMaxDistance.String(),
			"--collector.ntp.local-offset-tolerance", c.NTPLocalOffsetTolerance.String(),
		)

		flags = append(flags, booleanFlagMap(map[*bool]string{
			&c.NTPServerIsLocal: "collector.ntp.server-is-local",
		})...)
	}

	if collectors[CollectorPerf] {
		flags = append(flags, "--collector.perf.cpus", c.PerfCPUS)

		for _, tp := range c.PerfTracepoint {
			flags = append(flags, "--collector.perf.tracepoint", tp)
		}
	}

	if collectors[CollectorPowersuppply] {
		flags = append(flags, "--collector.powersupply.ignored-supplies", c.PowersupplyIgnoredSupplies)
	}

	if collectors[CollectorRunit] {
		flags = append(flags, "--collector.runit.servicedir", c.RunitServiceDir)
	}

	if collectors[CollectorSupervisord] {
		flags = append(flags, "--collector.supervisord.url", c.SupervisordURL)
	}

	if collectors[CollectorSystemd] {
		flags = append(flags,
			"--collector.systemd.unit-whitelist", c.SystemdUnitWhitelist,
			"--collector.systemd.unit-blacklist", c.SystemdUnitBlacklist,
		)

		flags = append(flags, booleanFlagMap(map[*bool]string{
			&c.SystemdEnableTaskMetrics:      "collector.systemd.enable-task-metrics",
			&c.SystemdEnableRestartsMetrics:  "collector.systemd.enable-restarts-metrics",
			&c.SystemdEnableStartTimeMetrics: "collector.systemd.enable-start-time-metrics",
		})...)
	}

	if collectors[CollectorVMStat] {
		flags = append(flags, "--collector.vmstat.fields", c.VMStatFields)
	}

	if collectors[CollectorTextfile] {
		flags = append(flags, "--collector.textfile.directory", c.TextfileDirectory)
	}

	return flags
}

func booleanFlagMap(m map[*bool]string) (flags []string) {
	for setting, key := range m {
		// The flag might not exist on this platform, so skip it if it's not
		// defined.
		if kingpin.CommandLine.GetFlag(key) == nil {
			continue
		}

		if *setting {
			flags = append(flags, "--"+key)
		} else {
			flags = append(flags, "--no-"+key)
		}
	}

	return
}
