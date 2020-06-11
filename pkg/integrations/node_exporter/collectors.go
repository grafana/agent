package node_exporter // nolint:golint

import (
	"fmt"
	"sort"

	"gopkg.in/alecthomas/kingpin.v2"
)

// CollectorState represents the default state of the collector, where it can
// either be enabled or disabled
type CollectorState bool

const (
	// CollectorStateDisabled represents a disabled collector that will not run
	// and collect metrics.
	CollectorStateDisabled CollectorState = false

	// CollectorStateEnabled represents an enabled collector that _will_ run
	// and collect metrics.
	CollectorStateEnabled CollectorState = true
)

// Collector is a specific collector that node_exporter runs.
type Collector string

// Collection of collectors defined by node_exporter
const (
	CollectorARP          = "arp"
	CollectorBCache       = "bcache"
	CollectorBonding      = "bonding"
	CollectorBooTtime     = "boottime"
	CollectorBTRFS        = "btrfs"
	CollectorBuddyInfo    = "buddyinfo"
	CollectorConntrack    = "conntrack"
	CollectorCPU          = "cpu"
	CollectorCPUFreq      = "cpufreq"
	CollectorDevstat      = "devstat"
	CollectorDiskstats    = "diskstats"
	CollectorDRBD         = "drbd"
	CollectorEDAC         = "edac"
	CollectorEntropy      = "entropy"
	CollectorExec         = "exec"
	CollectorFileFD       = "filefd"
	CollectorFilesystem   = "filesystem"
	CollectorHWMon        = "hwmon"
	CollectorInfiniband   = "infiniband"
	CollectorInterrupts   = "interrupts"
	CollectorIPVS         = "ipvs"
	CollectorKSMD         = "ksmd"
	CollectorLoadAvg      = "loadavg"
	CollectorLogind       = "logind"
	CollectorMDADM        = "mdadm"
	CollectorMeminfo      = "meminfo"
	CollectorMeminfoNuma  = "meminfo_numa"
	CollectorMountstats   = "mountstats"
	CollectorNetclass     = "netclass"
	CollectorNetdev       = "netdev"
	CollectorNetstat      = "netstat"
	CollectorNFS          = "nfs"
	CollectorNFSD         = "nfsd"
	CollectorNTP          = "ntp"
	CollectorPerf         = "perf"
	CollectorPowersuppply = "powersupplyclass"
	CollectorPressure     = "pressure"
	CollectorProcesses    = "processes"
	CollectorQDisc        = "qdisc"
	CollectorRAPL         = "rapl"
	CollectorRunit        = "runit"
	CollectorSchedstat    = "schedstat"
	CollectorSockstat     = "sockstat"
	CollectorSoftnet      = "softnet"
	CollectorStat         = "stat"
	CollectorSupervisord  = "supervisord"
	CollectorSystemd      = "systemd"
	CollectorTCPStat      = "tcpstat"
	CollectorTextfile     = "textfile"
	CollectorThermalzone  = "thermal_zone"
	CollectorTime         = "time"
	CollectorTimex        = "timex"
	CollectorUDPQueues    = "udp_queues"
	CollectorUname        = "uname"
	CollectorVMStat       = "vmstat"
	CollectorWiFi         = "wifi"
	CollectorXFS          = "xfs"
	CollectorZFS          = "zfs"
)

// Collectors holds a map of known collector names to their default
// state.
var Collectors = map[string]CollectorState{
	CollectorARP:          CollectorStateEnabled,
	CollectorBCache:       CollectorStateEnabled,
	CollectorBonding:      CollectorStateEnabled,
	CollectorBooTtime:     CollectorStateEnabled,
	CollectorBTRFS:        CollectorStateEnabled,
	CollectorBuddyInfo:    CollectorStateDisabled,
	CollectorConntrack:    CollectorStateEnabled,
	CollectorCPU:          CollectorStateEnabled,
	CollectorCPUFreq:      CollectorStateEnabled,
	CollectorDevstat:      CollectorStateDisabled,
	CollectorDiskstats:    CollectorStateEnabled,
	CollectorDRBD:         CollectorStateDisabled,
	CollectorEDAC:         CollectorStateEnabled,
	CollectorEntropy:      CollectorStateEnabled,
	CollectorExec:         CollectorStateEnabled,
	CollectorFileFD:       CollectorStateEnabled,
	CollectorFilesystem:   CollectorStateEnabled,
	CollectorHWMon:        CollectorStateEnabled,
	CollectorInfiniband:   CollectorStateEnabled,
	CollectorInterrupts:   CollectorStateDisabled,
	CollectorIPVS:         CollectorStateEnabled,
	CollectorKSMD:         CollectorStateDisabled,
	CollectorLoadAvg:      CollectorStateEnabled,
	CollectorLogind:       CollectorStateDisabled,
	CollectorMDADM:        CollectorStateEnabled,
	CollectorMeminfo:      CollectorStateEnabled,
	CollectorMeminfoNuma:  CollectorStateDisabled,
	CollectorMountstats:   CollectorStateDisabled,
	CollectorNetclass:     CollectorStateEnabled,
	CollectorNetdev:       CollectorStateEnabled,
	CollectorNetstat:      CollectorStateEnabled,
	CollectorNFS:          CollectorStateEnabled,
	CollectorNFSD:         CollectorStateEnabled,
	CollectorNTP:          CollectorStateDisabled,
	CollectorPerf:         CollectorStateDisabled,
	CollectorPowersuppply: CollectorStateEnabled,
	CollectorPressure:     CollectorStateEnabled,
	CollectorProcesses:    CollectorStateDisabled,
	CollectorQDisc:        CollectorStateDisabled,
	CollectorRAPL:         CollectorStateEnabled,
	CollectorRunit:        CollectorStateDisabled,
	CollectorSchedstat:    CollectorStateEnabled,
	CollectorSockstat:     CollectorStateEnabled,
	CollectorSoftnet:      CollectorStateEnabled,
	CollectorStat:         CollectorStateEnabled,
	CollectorSupervisord:  CollectorStateDisabled,
	CollectorSystemd:      CollectorStateDisabled,
	CollectorTCPStat:      CollectorStateDisabled,
	CollectorTextfile:     CollectorStateEnabled,
	CollectorThermalzone:  CollectorStateEnabled,
	CollectorTime:         CollectorStateEnabled,
	CollectorTimex:        CollectorStateEnabled,
	CollectorUDPQueues:    CollectorStateEnabled,
	CollectorUname:        CollectorStateEnabled,
	CollectorVMStat:       CollectorStateEnabled,
	CollectorWiFi:         CollectorStateDisabled,
	CollectorXFS:          CollectorStateEnabled,
	CollectorZFS:          CollectorStateEnabled,
}

// MapCollectorsToFlags takes in a map of collector keys and their states and
// converts them into flags that node_exporter expects. Collectors that are not
// defined will be ignored, which will be the case for collectors that are not
// supported on the host system.
func MapCollectorsToFlags(cs map[string]CollectorState) (flags []string) {
	for collector, state := range cs {
		flag := fmt.Sprintf("collector.%s", collector)

		// Skip the flag if it's not defined in kingpin
		if kingpin.CommandLine.GetFlag(flag) == nil {
			continue
		}

		switch state {
		case CollectorStateEnabled:
			flags = append(flags, "--"+flag)
		case CollectorStateDisabled:
			flags = append(flags, "--no-"+flag)
		}
	}

	sort.Strings(flags)
	return
}

// DisableUnavailableCollectors disables collectors that are not available on
// the host machine.
func DisableUnavailableCollectors(cs map[string]CollectorState) {
	for collector := range cs {
		flag := fmt.Sprintf("collector.%s", collector)

		// If kingpin doesn't have the flag, the collector is unavailable.
		if kingpin.CommandLine.GetFlag(flag) == nil {
			cs[collector] = CollectorStateDisabled
		}
	}
}
