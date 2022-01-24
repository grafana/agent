package node_exporter //nolint:golint

import (
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/go-kit/log"
	"github.com/grafana/agent/pkg/integrations"
	"github.com/grafana/dskit/flagext"
	"github.com/prometheus/procfs"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	// DefaultConfig holds non-zero default options for the Config when it is
	// unmarshaled from YAML.
	//
	// DefaultConfig's defaults are populated from init functions in this package.
	// See the init function here and in node_exporter_linux.go.
	DefaultConfig = Config{
		ProcFSPath: procfs.DefaultMountPoint,
		RootFSPath: "/",

		DiskStatsIgnoredDevices: "^(ram|loop|fd|(h|s|v|xv)d[a-z]|nvme\\d+n\\d+p)\\d+$",

		EthtoolMetricsInclude: ".*",

		FilesystemMountTimeout: 5 * time.Second,

		NTPIPTTL:                1,
		NTPLocalOffsetTolerance: time.Millisecond,
		NTPMaxDistance:          time.Microsecond * 3466080,
		NTPProtocolVersion:      4,
		NTPServer:               "127.0.0.1",

		NetclassIgnoredDevices: "^$",
		NetstatFields:          "^(.*_(InErrors|InErrs)|Ip_Forwarding|Ip(6|Ext)_(InOctets|OutOctets)|Icmp6?_(InMsgs|OutMsgs)|TcpExt_(Listen.*|Syncookies.*|TCPSynRetrans|TCPTimeouts)|Tcp_(ActiveOpens|InSegs|OutSegs|OutRsts|PassiveOpens|RetransSegs|CurrEstab)|Udp6?_(InDatagrams|OutDatagrams|NoPorts|RcvbufErrors|SndbufErrors))$",

		PowersupplyIgnoredSupplies: "^$",

		RunitServiceDir: "/etc/service",

		SupervisordURL: "http://localhost:9001/RPC2",

		SystemdUnitExclude: ".+\\.(automount|device|mount|scope|slice)",
		SystemdUnitInclude: ".+",

		TapestatsIgnoredDevices: "^$",

		VMStatFields: "^(oom_kill|pgpg|pswp|pg.*fault).*",
	}
)

func init() {
	// The default values for the filesystem collector are to ignore everything,
	// but some platforms have specific defaults. We'll fill these in below at
	// initialization time, but the values can still be overridden via the config
	// file.
	switch runtime.GOOS {
	case "linux":
		DefaultConfig.FilesystemMountPointsExclude = "^/(dev|proc|run/credentials/.+|sys|var/lib/docker/.+)($|/)"
		DefaultConfig.FilesystemFSTypesExclude = "^(autofs|binfmt_misc|bpf|cgroup2?|configfs|debugfs|devpts|devtmpfs|fusectl|hugetlbfs|iso9660|mqueue|nsfs|overlay|proc|procfs|pstore|rpc_pipefs|securityfs|selinuxfs|squashfs|sysfs|tracefs)$"
	case "freebsd", "netbsd", "openbsd", "darwin":
		DefaultConfig.FilesystemMountPointsExclude = "^/(dev)($|/)"
		DefaultConfig.FilesystemFSTypesExclude = "^devfs$"
	}

	if url := os.Getenv("SUPERVISORD_URL"); url != "" {
		DefaultConfig.SupervisordURL = url
	}
}

// Config controls the node_exporter integration.
type Config struct {
	IncludeExporterMetrics bool `yaml:"include_exporter_metrics,omitempty"`

	ProcFSPath string `yaml:"procfs_path,omitempty"`
	SysFSPath  string `yaml:"sysfs_path,omitempty"`
	RootFSPath string `yaml:"rootfs_path,omitempty"`

	// Collectors to mark as enabled
	EnableCollectors flagext.StringSlice `yaml:"enable_collectors,omitempty"`

	// Collectors to mark as disabled
	DisableCollectors flagext.StringSlice `yaml:"disable_collectors,omitempty"`

	// Overrides the default set of enabled collectors with the collectors
	// listed.
	SetCollectors flagext.StringSlice `yaml:"set_collectors,omitempty"`

	// Collector-specific config options
	BcachePriorityStats              bool                `yaml:"enable_bcache_priority_stats,omitempty"`
	CPUBugsInclude                   string              `yaml:"cpu_bugs_include,omitempty"`
	CPUEnableCPUGuest                bool                `yaml:"enable_cpu_guest_seconds_metric,omitempty"`
	CPUEnableCPUInfo                 bool                `yaml:"enable_cpu_info_metric,omitempty"`
	CPUFlagsInclude                  string              `yaml:"cpu_flags_include,omitempty"`
	DiskStatsIgnoredDevices          string              `yaml:"diskstats_ignored_devices,omitempty"`
	EthtoolDeviceExclude             string              `yaml:"ethtool_device_exclude,omitempty"`
	EthtoolDeviceInclude             string              `yaml:"ethtool_device_include,omitempty"`
	EthtoolMetricsInclude            string              `yaml:"ethtool_metrics_include,omitempty"`
	FilesystemFSTypesExclude         string              `yaml:"filesystem_fs_types_exclude,omitempty"`
	FilesystemMountPointsExclude     string              `yaml:"filesystem_mount_points_exclude,omitempty"`
	FilesystemMountTimeout           time.Duration       `yaml:"filesystem_mount_timeout,omitempty"`
	IPVSBackendLabels                []string            `yaml:"ipvs_backend_labels,omitempty"`
	NTPIPTTL                         int                 `yaml:"ntp_ip_ttl,omitempty"`
	NTPLocalOffsetTolerance          time.Duration       `yaml:"ntp_local_offset_tolerance,omitempty"`
	NTPMaxDistance                   time.Duration       `yaml:"ntp_max_distance,omitempty"`
	NTPProtocolVersion               int                 `yaml:"ntp_protocol_version,omitempty"`
	NTPServer                        string              `yaml:"ntp_server,omitempty"`
	NTPServerIsLocal                 bool                `yaml:"ntp_server_is_local,omitempty"`
	NetclassIgnoreInvalidSpeedDevice bool                `yaml:"netclass_ignore_invalid_speed_device,omitempty"`
	NetclassIgnoredDevices           string              `yaml:"netclass_ignored_devices,omitempty"`
	NetdevAddressInfo                bool                `yaml:"netdev_address_info,omitempty"`
	NetdevDeviceExclude              string              `yaml:"netdev_device_exclude,omitempty"`
	NetdevDeviceInclude              string              `yaml:"netdev_device_include,omitempty"`
	NetstatFields                    string              `yaml:"netstat_fields,omitempty"`
	PerfCPUS                         string              `yaml:"perf_cpus,omitempty"`
	PerfTracepoint                   flagext.StringSlice `yaml:"perf_tracepoint,omitempty"`
	PowersupplyIgnoredSupplies       string              `yaml:"powersupply_ignored_supplies,omitempty"`
	RunitServiceDir                  string              `yaml:"runit_servic e_dir,omitempty"`
	SupervisordURL                   string              `yaml:"supervisord_url,omitempty"`
	SystemdEnableRestartsMetrics     bool                `yaml:"systemd_enable_restarts_metrics,omitempty"`
	SystemdEnableStartTimeMetrics    bool                `yaml:"systemd_enable_start_time_metrics,omitempty"`
	SystemdEnableTaskMetrics         bool                `yaml:"systemd_enable_task_metrics,omitempty"`
	SystemdUnitExclude               string              `yaml:"systemd_unit_exclude,omitempty"`
	SystemdUnitInclude               string              `yaml:"systemd_unit_include,omitempty"`
	TapestatsIgnoredDevices          string              `yaml:"tapestats_ignored_devices,omitempty"`
	TextfileDirectory                string              `yaml:"textfile_directory,omitempty"`
	VMStatFields                     string              `yaml:"vmstat_fields,omitempty"`

	UnmarshalWarnings []string `yaml:"-"`
}

// UnmarshalYAML implements yaml.Unmarshaler for Config.
func (c *Config) UnmarshalYAML(unmarshal func(interface{}) error) error {
	*c = DefaultConfig

	type baseConfig Config
	type config struct {
		baseConfig `yaml:",inline"`

		// Deprecated field names:
		NetdevDeviceWhitelist        string `yaml:"netdev_device_whitelist,omitempty"`
		NetdevDeviceBlacklist        string `yaml:"netdev_device_blacklist,omitempty"`
		SystemdUnitWhitelist         string `yaml:"systemd_unit_whitelist,omitempty"`
		SystemdUnitBlacklist         string `yaml:"systemd_unit_blacklist,omitempty"`
		FilesystemIgnoredMountPoints string `yaml:"filesystem_ignored_mount_points,omitempty"`
		FilesystemIgnoredFSTypes     string `yaml:"filesystem_ignored_fs_types,omitempty"`
	}

	var fc config // our full config (schema + deprecated names)
	fc.baseConfig = baseConfig(*c)

	type migratedField struct {
		OldName, NewName   string
		OldValue, NewValue *string

		defaultValue string
	}
	migratedFields := []*migratedField{
		{
			OldName: "netdev_device_whitelist", NewName: "netdev_device_include",
			OldValue: &fc.NetdevDeviceWhitelist, NewValue: &fc.NetdevDeviceInclude,
		},
		{
			OldName: "netdev_device_blacklist", NewName: "netdev_device_exclude",
			OldValue: &fc.NetdevDeviceBlacklist, NewValue: &fc.NetdevDeviceExclude,
		},
		{
			OldName: "systemd_unit_whitelist", NewName: "systemd_unit_include",
			OldValue: &fc.SystemdUnitWhitelist, NewValue: &fc.SystemdUnitInclude,
		},
		{
			OldName: "systemd_unit_blacklist", NewName: "systemd_unit_exclude",
			OldValue: &fc.SystemdUnitBlacklist, NewValue: &fc.SystemdUnitExclude,
		},
		{
			OldName: "filesystem_ignored_mount_points", NewName: "filesystem_mount_points_exclude",
			OldValue: &fc.FilesystemIgnoredMountPoints, NewValue: &fc.FilesystemMountPointsExclude,
		},
		{
			OldName: "filesystem_ignored_fs_types", NewName: "filesystem_fs_types_exclude",
			OldValue: &fc.FilesystemIgnoredFSTypes, NewValue: &fc.FilesystemFSTypesExclude,
		},
	}

	// We don't know when fields are unmarshaled unless they have non-zero
	// values. Defaults stop us from being able to check, so we'll temporarily
	// cache the default and make sure both the old and new migrated fields are
	// zero.
	for _, mf := range migratedFields {
		mf.defaultValue = *mf.NewValue
		*mf.NewValue = ""
	}

	if err := unmarshal(&fc); err != nil {
		return err
	}

	for _, mf := range migratedFields {
		switch {
		case *mf.OldValue != "" && *mf.NewValue != "": // New set, old set
			return fmt.Errorf("only one of %q and %q may be specified", mf.OldName, mf.NewName)

		case *mf.NewValue == "" && *mf.OldValue != "": // New unset, old set
			*mf.NewValue = *mf.OldValue

			warning := fmt.Sprintf("%q is deprecated by %q and will be removed in a future version", mf.OldName, mf.NewName)
			fc.UnmarshalWarnings = append(fc.UnmarshalWarnings, warning)

		case *mf.NewValue == "" && *mf.OldValue == "": // Neither set.
			// Copy the default back to mf.NewValue.
			*mf.NewValue = mf.defaultValue

		case *mf.NewValue != "" && *mf.OldValue == "": // New set, old unset
			// Nothing to do
		}
	}

	*c = (Config)(fc.baseConfig)
	return nil
}

// Name returns the name of the integration that this config represents.
func (c *Config) Name() string {
	return "node_exporter"
}

// InstanceKey returns the hostname:port of the agent process.
func (c *Config) InstanceKey(agentKey string) (string, error) {
	return agentKey, nil
}

// NewIntegration converts this config into an instance of an integration.
func (c *Config) NewIntegration(l log.Logger) (integrations.Integration, error) {
	return New(l, c)
}

// MapConfigToNodeExporterFlags takes in a node_exporter Config and converts
// it to the set of flags that node_exporter usually expects when running as a
// separate binary.
func MapConfigToNodeExporterFlags(c *Config) (accepted []string, ignored []string) {
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

	var flags flags
	flags.accepted = append(flags.accepted, MapCollectorsToFlags(collectors)...)

	flags.add(
		"--path.procfs", c.ProcFSPath,
		"--path.sysfs", c.SysFSPath,
		"--path.rootfs", c.RootFSPath,
	)

	if collectors[CollectorBCache] {
		flags.addBools(map[*bool]string{
			&c.BcachePriorityStats: "collector.bcache.priorityStats",
		})
	}

	if collectors[CollectorCPU] {
		flags.addBools(map[*bool]string{
			&c.CPUEnableCPUGuest: "collector.cpu.guest",
			&c.CPUEnableCPUInfo:  "collector.cpu.info",
		})
		flags.add("--collector.cpu.info.flags-include", c.CPUFlagsInclude)
		flags.add("--collector.cpu.info.bugs-include", c.CPUBugsInclude)
	}

	if collectors[CollectorDiskstats] {
		flags.add("--collector.diskstats.ignored-devices", c.DiskStatsIgnoredDevices)
	}

	if collectors[CollectorEthtool] {
		flags.add("--collector.ethtool.device-include", c.EthtoolDeviceInclude)
		flags.add("--collector.ethtool.device-exclude", c.EthtoolDeviceExclude)
		flags.add("--collector.ethtool.metrics-include", c.EthtoolMetricsInclude)
	}

	if collectors[CollectorFilesystem] {
		flags.add(
			"--collector.filesystem.mount-timeout", c.FilesystemMountTimeout.String(),
			"--collector.filesystem.mount-points-exclude", c.FilesystemMountPointsExclude,
			"--collector.filesystem.fs-types-exclude", c.FilesystemFSTypesExclude,
		)
	}

	if collectors[CollectorIPVS] {
		flags.add("--collector.ipvs.backend-labels", strings.Join(c.IPVSBackendLabels, ","))
	}

	if collectors[CollectorNetclass] {
		flags.addBools(map[*bool]string{
			&c.NetclassIgnoreInvalidSpeedDevice: "collector.netclass.ignore-invalid-speed",
		})

		flags.add("--collector.netclass.ignored-devices", c.NetclassIgnoredDevices)
	}

	if collectors[CollectorNetdev] {
		flags.addBools(map[*bool]string{
			&c.NetdevAddressInfo: "collector.netdev.address-info",
		})

		flags.add(
			"--collector.netdev.device-include", c.NetdevDeviceInclude,
			"--collector.netdev.device-exclude", c.NetdevDeviceExclude,
		)
	}

	if collectors[CollectorNetstat] {
		flags.add("--collector.netstat.fields", c.NetstatFields)
	}

	if collectors[CollectorNTP] {
		flags.add(
			"--collector.ntp.server", c.NTPServer,
			"--collector.ntp.protocol-version", fmt.Sprintf("%d", c.NTPProtocolVersion),
			"--collector.ntp.ip-ttl", fmt.Sprintf("%d", c.NTPIPTTL),
			"--collector.ntp.max-distance", c.NTPMaxDistance.String(),
			"--collector.ntp.local-offset-tolerance", c.NTPLocalOffsetTolerance.String(),
		)

		flags.addBools(map[*bool]string{
			&c.NTPServerIsLocal: "collector.ntp.server-is-local",
		})
	}

	if collectors[CollectorPerf] {
		flags.add("--collector.perf.cpus", c.PerfCPUS)

		for _, tp := range c.PerfTracepoint {
			flags.add("--collector.perf.tracepoint", tp)
		}
	}

	if collectors[CollectorPowersuppply] {
		flags.add("--collector.powersupply.ignored-supplies", c.PowersupplyIgnoredSupplies)
	}

	if collectors[CollectorRunit] {
		flags.add("--collector.runit.servicedir", c.RunitServiceDir)
	}

	if collectors[CollectorSupervisord] {
		flags.add("--collector.supervisord.url", c.SupervisordURL)
	}

	if collectors[CollectorSystemd] {
		flags.add(
			"--collector.systemd.unit-include", c.SystemdUnitInclude,
			"--collector.systemd.unit-exclude", c.SystemdUnitExclude,
		)

		flags.addBools(map[*bool]string{
			&c.SystemdEnableTaskMetrics:      "collector.systemd.enable-task-metrics",
			&c.SystemdEnableRestartsMetrics:  "collector.systemd.enable-restarts-metrics",
			&c.SystemdEnableStartTimeMetrics: "collector.systemd.enable-start-time-metrics",
		})
	}

	if collectors[CollectorTapestats] {
		flags.add("--collector.tapestats.ignored-devices", c.TapestatsIgnoredDevices)
	}

	if collectors[CollectorTextfile] {
		flags.add("--collector.textfile.directory", c.TextfileDirectory)
	}

	if collectors[CollectorVMStat] {
		flags.add("--collector.vmstat.fields", c.VMStatFields)
	}

	return flags.accepted, flags.ignored
}

type flags struct {
	accepted []string
	ignored  []string
}

// add pushes new flags as key value pairs. If the flag isn't registered with kingpin,
// it will be ignored.
func (f *flags) add(kvp ...string) {
	if (len(kvp) % 2) != 0 {
		panic("missing value for added flag")
	}

	for i := 0; i < len(kvp); i += 2 {
		key := kvp[i+0]
		value := kvp[i+1]

		rawFlag := strings.TrimPrefix(key, "--")
		if kingpin.CommandLine.GetFlag(rawFlag) == nil {
			f.ignored = append(f.ignored, rawFlag)
			continue
		}

		f.accepted = append(f.accepted, key, value)
	}
}

func (f *flags) addBools(m map[*bool]string) {
	for setting, key := range m {
		// The flag might not exist on this platform, so skip it if it's not
		// defined.
		if kingpin.CommandLine.GetFlag(key) == nil {
			f.ignored = append(f.ignored, key)
			continue
		}

		if *setting {
			f.accepted = append(f.accepted, "--"+key)
		} else {
			f.accepted = append(f.accepted, "--no-"+key)
		}
	}
}
