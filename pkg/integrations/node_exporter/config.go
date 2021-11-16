package node_exporter //nolint:golint

import (
	"fmt"
	"runtime"
	"strings"
	"time"

	"github.com/cortexproject/cortex/pkg/util/flagext"
	"github.com/go-kit/kit/log"
	"github.com/grafana/agent/pkg/integrations"
	"github.com/grafana/agent/pkg/integrations/config"
	"github.com/prometheus/procfs"
	"github.com/prometheus/procfs/sysfs"
	"gopkg.in/alecthomas/kingpin.v2"
)

// DefaultConfig holds non-zero default options for the Config when it is
// unmarshaled from YAML.
var DefaultConfig = Config{
	Common:     config.DefaultCommon,
	ProcFSPath: procfs.DefaultMountPoint,
	SysFSPath:  sysfs.DefaultMountPoint,
	RootFSPath: "/",

	DiskStatsIgnoredDevices: "^(ram|loop|fd|(h|s|v|xv)d[a-z]|nvme\\d+n\\d+p)\\d+$",

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

func init() {
	// The default values for the filesystem collector are to ignore everything,
	// but some platforms have specific defaults. We'll fill these in below at
	// initialization time, but the values can still be overridden via the config
	// file.
	switch runtime.GOOS {
	case "linux":
		DefaultConfig.FilesystemIgnoredMountPoints = "^/(dev|proc|sys|var/lib/docker/.+)($|/)"
		DefaultConfig.FilesystemIgnoredFSTypes = "^(autofs|binfmt_misc|bpf|cgroup2?|configfs|debugfs|devpts|devtmpfs|fusectl|hugetlbfs|iso9660|mqueue|nsfs|overlay|proc|procfs|pstore|rpc_pipefs|securityfs|selinuxfs|squashfs|sysfs|tracefs)$"
	case "freebsd", "netbsd", "openbsd", "darwin":
		DefaultConfig.FilesystemIgnoredMountPoints = "^/(dev)($|/)"
		DefaultConfig.FilesystemIgnoredFSTypes = "^devfs$"
	}
}

// Config controls the node_exporter integration.
type Config struct {
	Common config.Common `yaml:",inline"`

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
	CPUEnableCPUInfo              bool                `yaml:"enable_cpu_info_metric,omitempty"`
	DiskStatsIgnoredDevices       string              `yaml:"diskstats_ignored_devices,omitempty"`
	FilesystemIgnoredMountPoints  string              `yaml:"filesystem_ignored_mount_points,omitempty"`
	FilesystemIgnoredFSTypes      string              `yaml:"filesystem_ignored_fs_types,omitempty"`
	NetclassIgnoredDevices        string              `yaml:"netclass_ignored_devices,omitempty"`
	NetdevDeviceBlacklist         string              `yaml:"netdev_device_blacklist,omitempty"`
	NetdevDeviceWhitelist         string              `yaml:"netdev_device_whitelist,omitempty"`
	NetstatFields                 string              `yaml:"netstat_fields,omitempty"`
	NTPServer                     string              `yaml:"ntp_server,omitempty"`
	NTPProtocolVersion            int                 `yaml:"ntp_protocol_version,omitempty"`
	NTPServerIsLocal              bool                `yaml:"ntp_server_is_local,omitempty"`
	NTPIPTTL                      int                 `yaml:"ntp_ip_ttl,omitempty"`
	NTPMaxDistance                time.Duration       `yaml:"ntp_max_distance,omitempty"`
	NTPLocalOffsetTolerance       time.Duration       `yaml:"ntp_local_offset_tolerance,omitempty"`
	PerfCPUS                      string              `yaml:"perf_cpus,omitempty"`
	PerfTracepoint                flagext.StringSlice `yaml:"perf_tracepoint,omitempty"`
	PowersupplyIgnoredSupplies    string              `yaml:"powersupply_ignored_supplies,omitempty"`
	RunitServiceDir               string              `yaml:"runit_service_dir,omitempty"`
	SupervisordURL                string              `yaml:"supervisord_url,omitempty"`
	SystemdUnitWhitelist          string              `yaml:"systemd_unit_whitelist,omitempty"`
	SystemdUnitBlacklist          string              `yaml:"systemd_unit_blacklist,omitempty"`
	SystemdEnableTaskMetrics      bool                `yaml:"systemd_enable_task_metrics,omitempty"`
	SystemdEnableRestartsMetrics  bool                `yaml:"systemd_enable_restarts_metrics,omitempty"`
	SystemdEnableStartTimeMetrics bool                `yaml:"systemd_enable_start_time_metrics,omitempty"`
	VMStatFields                  string              `yaml:"vmstat_fields,omitempty"`
	TextfileDirectory             string              `yaml:"textfile_directory,omitempty"`
}

// UnmarshalYAML implements yaml.Unmarshaler for Config.
func (c *Config) UnmarshalYAML(unmarshal func(interface{}) error) error {
	*c = DefaultConfig

	type plain Config
	return unmarshal((*plain)(c))
}

// Name returns the name of the integration that this config represents.
func (c *Config) Name() string {
	return "node_exporter"
}

// CommonConfig returns the common configs that are shared across all integrations.
func (c *Config) CommonConfig() config.Common {
	return c.Common
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

	if collectors[CollectorCPU] {
		flags.addBools(map[*bool]string{
			&c.CPUEnableCPUInfo: "collector.cpu.info",
		})
	}

	if collectors[CollectorDiskstats] {
		flags.add("--collector.diskstats.ignored-devices", c.DiskStatsIgnoredDevices)
	}

	if collectors[CollectorFilesystem] {
		flags.add(
			"--collector.filesystem.ignored-mount-points", c.FilesystemIgnoredMountPoints,
			"--collector.filesystem.ignored-fs-types", c.FilesystemIgnoredFSTypes,
		)
	}

	if collectors[CollectorNetclass] {
		flags.add("--collector.netclass.ignored-devices", c.NetclassIgnoredDevices)
	}

	if collectors[CollectorNetdev] {
		flags.add(
			"--collector.netdev.device-blacklist", c.NetdevDeviceBlacklist,
			"--collector.netdev.device-whitelist", c.NetdevDeviceWhitelist,
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
			"--collector.systemd.unit-whitelist", c.SystemdUnitWhitelist,
			"--collector.systemd.unit-blacklist", c.SystemdUnitBlacklist,
		)

		flags.addBools(map[*bool]string{
			&c.SystemdEnableTaskMetrics:      "collector.systemd.enable-task-metrics",
			&c.SystemdEnableRestartsMetrics:  "collector.systemd.enable-restarts-metrics",
			&c.SystemdEnableStartTimeMetrics: "collector.systemd.enable-start-time-metrics",
		})
	}

	if collectors[CollectorVMStat] {
		flags.add("--collector.vmstat.fields", c.VMStatFields)
	}

	if collectors[CollectorTextfile] {
		flags.add("--collector.textfile.directory", c.TextfileDirectory)
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
