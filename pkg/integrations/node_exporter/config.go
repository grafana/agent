package node_exporter //nolint:golint

import (
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/go-kit/log"
	"github.com/grafana/agent/pkg/integrations"
	integrations_v2 "github.com/grafana/agent/pkg/integrations/v2"
	"github.com/grafana/agent/pkg/integrations/v2/metricsutils"
	"github.com/grafana/dskit/flagext"
	"github.com/hashicorp/hcl/v2"
	"github.com/prometheus/procfs"
	"github.com/rfratto/gohcl"
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
	case "darwin":
		DefaultConfig.FilesystemMountPointsExclude = "^/(dev)($|/)"
		DefaultConfig.FilesystemFSTypesExclude = "^(autofs|devfs)$"
	case "freebsd", "netbsd", "openbsd":
		DefaultConfig.FilesystemMountPointsExclude = "^/(dev)($|/)"
		DefaultConfig.FilesystemFSTypesExclude = "^devfs$"
	}

	if url := os.Getenv("SUPERVISORD_URL"); url != "" {
		DefaultConfig.SupervisordURL = url
	}
}

// Config controls the node_exporter integration.
type Config struct {
	IncludeExporterMetrics bool `yaml:"include_exporter_metrics,omitempty" hcl:"include_exporter_metrics,optional"`

	ProcFSPath string `yaml:"procfs_path,omitempty" hcl:"procfs_path,optional"`
	SysFSPath  string `yaml:"sysfs_path,omitempty" hcl:"sysfs_path,optional"`
	RootFSPath string `yaml:"rootfs_path,omitempty" hcl:"rootfs_path,optional"`

	// Collectors to mark as enabled
	EnableCollectors flagext.StringSlice `yaml:"enable_collectors,omitempty" hcl:"enable_collectors,optional"`

	// Collectors to mark as disabled
	DisableCollectors flagext.StringSlice `yaml:"disable_collectors,omitempty" hcl:"disable_collectors,optional"`

	// Overrides the default set of enabled collectors with the collectors
	// listed.
	SetCollectors flagext.StringSlice `yaml:"set_collectors,omitempty" hcl:"set_collectors,optional"`

	// Collector-specific config options
	BcachePriorityStats              bool                `yaml:"enable_bcache_priority_stats,omitempty" hcl:"enable_bcache_priority_stats,optional"`
	CPUBugsInclude                   string              `yaml:"cpu_bugs_include,omitempty" hcl:"cpu_bugs_include,optional"`
	CPUEnableCPUGuest                bool                `yaml:"enable_cpu_guest_seconds_metric,omitempty" hcl:"enable_cpu_guest_seconds_metric,optional"`
	CPUEnableCPUInfo                 bool                `yaml:"enable_cpu_info_metric,omitempty" hcl:"enable_cpu_info_metric,optional"`
	CPUFlagsInclude                  string              `yaml:"cpu_flags_include,omitempty" hcl:"cpu_flags_include,optional"`
	DiskStatsIgnoredDevices          string              `yaml:"diskstats_ignored_devices,omitempty" hcl:"diskstats_ignored_devices,optional"`
	EthtoolDeviceExclude             string              `yaml:"ethtool_device_exclude,omitempty" hcl:"ethtool_device_exclude,optional"`
	EthtoolDeviceInclude             string              `yaml:"ethtool_device_include,omitempty" hcl:"ethtool_device_include,optional"`
	EthtoolMetricsInclude            string              `yaml:"ethtool_metrics_include,omitempty" hcl:"ethtool_metrics_include,optional"`
	FilesystemFSTypesExclude         string              `yaml:"filesystem_fs_types_exclude,omitempty" hcl:"filesystem_fs_types_exclude,optional"`
	FilesystemMountPointsExclude     string              `yaml:"filesystem_mount_points_exclude,omitempty" hcl:"filesystem_mount_points_exclude,optional"`
	FilesystemMountTimeout           time.Duration       `yaml:"filesystem_mount_timeout,omitempty" hcl:"filesystem_mount_timeout,optional"`
	IPVSBackendLabels                []string            `yaml:"ipvs_backend_labels,omitempty" hcl:"ipvs_backend_labels,optional"`
	NTPIPTTL                         int                 `yaml:"ntp_ip_ttl,omitempty" hcl:"ntp_ip_ttl,optional"`
	NTPLocalOffsetTolerance          time.Duration       `yaml:"ntp_local_offset_tolerance,omitempty" hcl:"ntp_local_offset_tolerance,optional"`
	NTPMaxDistance                   time.Duration       `yaml:"ntp_max_distance,omitempty" hcl:"ntp_max_distance,optional"`
	NTPProtocolVersion               int                 `yaml:"ntp_protocol_version,omitempty" hcl:"ntp_protocol_version,optional"`
	NTPServer                        string              `yaml:"ntp_server,omitempty" hcl:"ntp_server,optional"`
	NTPServerIsLocal                 bool                `yaml:"ntp_server_is_local,omitempty" hcl:"ntp_server_is_local,optional"`
	NetclassIgnoreInvalidSpeedDevice bool                `yaml:"netclass_ignore_invalid_speed_device,omitempty" hcl:"netclass_ignore_invalid_speed_device,optional"`
	NetclassIgnoredDevices           string              `yaml:"netclass_ignored_devices,omitempty" hcl:"netclass_ignored_devices,optional"`
	NetdevAddressInfo                bool                `yaml:"netdev_address_info,omitempty" hcl:"netdev_address_info,optional"`
	NetdevDeviceExclude              string              `yaml:"netdev_device_exclude,omitempty" hcl:"netdev_device_exclude,optional"`
	NetdevDeviceInclude              string              `yaml:"netdev_device_include,omitempty" hcl:"netdev_device_include,optional"`
	NetstatFields                    string              `yaml:"netstat_fields,omitempty" hcl:"netstat_fields,optional"`
	PerfCPUS                         string              `yaml:"perf_cpus,omitempty" hcl:"perf_cpus,optional"`
	PerfTracepoint                   flagext.StringSlice `yaml:"perf_tracepoint,omitempty" hcl:"perf_tracepoint,optional"`
	PowersupplyIgnoredSupplies       string              `yaml:"powersupply_ignored_supplies,omitempty" hcl:"powersupply_ignored_supplies,optional"`
	RunitServiceDir                  string              `yaml:"runit_service_dir,omitempty" hcl:"runit_service_dir,optional"`
	SupervisordURL                   string              `yaml:"supervisord_url,omitempty" hcl:"supervisord_url,optional"`
	SystemdEnableRestartsMetrics     bool                `yaml:"systemd_enable_restarts_metrics,omitempty" hcl:"systemd_enable_restarts_metrics,optional"`
	SystemdEnableStartTimeMetrics    bool                `yaml:"systemd_enable_start_time_metrics,omitempty" hcl:"systemd_enable_start_time_metrics,optional"`
	SystemdEnableTaskMetrics         bool                `yaml:"systemd_enable_task_metrics,omitempty" hcl:"systemd_enable_task_metrics,optional"`
	SystemdUnitExclude               string              `yaml:"systemd_unit_exclude,omitempty" hcl:"systemd_unit_exclude,optional"`
	SystemdUnitInclude               string              `yaml:"systemd_unit_include,omitempty" hcl:"systemd_unit_include,optional"`
	TapestatsIgnoredDevices          string              `yaml:"tapestats_ignored_devices,omitempty" hcl:"tapestats_ignored_devices,optional"`
	TextfileDirectory                string              `yaml:"textfile_directory,omitempty" hcl:"textfile_directory,optional"`
	VMStatFields                     string              `yaml:"vmstat_fields,omitempty" hcl:"vmstat_fields,optional"`

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

func (c *Config) DecodeHCL(body hcl.Body, ctx *hcl.EvalContext) error {
	*c = DefaultConfig

	type cfg Config
	//TODO: (cpeterson) This does not have feature parity with yaml.
	// It does not map migrated fields like UnmarshalYAML does
	return gohcl.DecodeBody(body, ctx, (*cfg)(c))
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

func init() {
	integrations.RegisterIntegration(&Config{})
	integrations_v2.RegisterLegacy(&Config{}, integrations_v2.TypeSingleton, metricsutils.Shim)
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
