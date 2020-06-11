package node_exporter //nolint:golint

import (
	"fmt"
	"time"

	"github.com/cortexproject/cortex/pkg/util/flagext"
	"github.com/grafana/agent/pkg/integrations/config"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	// DefaultConfig holds non-zero default options for the Config when it is
	// unmarshaled from YAML.
	DefaultConfig = Config{
		ProcFSPath: "/proc",
		SysFSPath:  "/sys",
		RootFSPath: "/",

		DiskStatsIgnoredDevices: "^(ram|loop|fd|(h|s|v|xv)d[a-z]|nvme\\d+n\\d+p)\\d+$",

		FilesystemIgnoredMountPoints: "^/(dev|proc|sys|var/lib/docker/.+)($|/)",
		FilesystemIgnoredFSTypes:     "^(autofs|binfmt_misc|bpf|cgroup2?|configfs|debugfs|devpts|devtmpfs|fusectl|hugetlbfs|iso9660|mqueue|nsfs|overlay|proc|procfs|pstore|rpc_pipefs|securityfs|selinuxfs|squashfs|sysfs|tracefs)$",

		NetclassIgnoredDevices: "^$",
		NetstatFields:          "^(.*_(InErrors|InErrs)|Ip_Forwarding|Ip(6|Ext)_(InOctets|OutOctets)|Icmp6?_(InMsgs|OutMsgs)|TcpExt_(Listen.*|Syncookies.*|TCPSynRetrans)|Tcp_(ActiveOpens|InSegs|OutSegs|PassiveOpens|RetransSegs|CurrEstab)|Udp6?_(InDatagrams|OutDatagrams|NoPorts|RcvbufErrors|SndbufErrors))$",

		NTPServer:               "127.0.0.1",
		NTPProtocolVersion:      4,
		NTPIPTTL:                1,
		NTPMaxDistance:          time.Microsecond * 3466080,
		NTPLocalOffsetTolerance: time.Millisecond,

		PowersupplyIgnoredSupplies: "^$",

		RunitServiceDir: "/etc/service",

		SupervisordURL: "http://localhost:9001/RPC2",

		SystemdUnitWhitelist: ".+",
		SystemdUnitBlacklist: ".+\\.(automount|device|mount|scope|slice)",
		VMStatFields:         "^(oom_kill|pgpg|pswp|pg.*fault).*",
	}
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

// UnmarshalYAML implements yaml.Unmarshaler for Config.
func (c *Config) UnmarshalYAML(unmarshal func(interface{}) error) error {
	*c = DefaultConfig

	type plain Config
	return unmarshal((*plain)(c))
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
