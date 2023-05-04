package collector

import (
	"sort"
	"strconv"
	"strings"

	"github.com/leoluk/perflib_exporter/perflib"
	"github.com/prometheus-community/windows_exporter/log"
	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/sys/windows/registry"
	"gopkg.in/alecthomas/kingpin.v2"
)

// ...
const (
	// TODO: Make package-local
	Namespace = "windows"

	// Conversion factors
	ticksToSecondsScaleFactor = 1 / 1e7
	windowsEpoch              = 116444736000000000
)

// getWindowsVersion reads the version number of the OS from the Registry
// See https://docs.microsoft.com/en-us/windows/desktop/sysinfo/operating-system-version
func getWindowsVersion() float64 {
	k, err := registry.OpenKey(registry.LOCAL_MACHINE, `SOFTWARE\Microsoft\Windows NT\CurrentVersion`, registry.QUERY_VALUE)
	if err != nil {
		log.Warn("Couldn't open registry", err)
		return 0
	}
	defer func() {
		err = k.Close()
		if err != nil {
			log.Warnf("Failed to close registry key: %v", err)
		}
	}()

	currentv, _, err := k.GetStringValue("CurrentVersion")
	if err != nil {
		log.Warn("Couldn't open registry to determine current Windows version:", err)
		return 0
	}

	currentv_flt, err := strconv.ParseFloat(currentv, 64)

	log.Debugf("Detected Windows version %f\n", currentv_flt)

	return currentv_flt
}

// Config is a generic configuration for a Collector.
type Config interface {
	// Name should return the name of the collector.
	Name() string

	// RegisterFlags should add config options to the provided Kingpin
	// application.
	RegisterFlags(ka *kingpin.Application)

	// Build should build a collector from the Config.
	Build() (Collector, error)
}

// AllConfigs returns the full set of configs for available Collectors.
func AllConfigs() []Config {
	return []Config{
		&ADConfig{},
		&ADFSConfig{},
		&CacheConfig{},
		&ContainerMetricsConfig{},
		&CPUConfig{},
		&CpuInfoConfig{},
		&CSConfig{},
		&DFSRConfig{},
		&DHCPConfig{},
		&DNSConfig{},
		&ExchangeConfig{},
		&FSRMQuotaConfig{},
		&HyperVConfig{},
		&IISConfig{},
		&LogicalDiskConfig{},
		&LogonConfig{},
		&MemoryConfig{},
		&MSMQConfig{},
		&MSSQLConfig{},
		&NetworkConfig{},
		&NETFramework_NETCLRExceptionsConfig{},
		&NETFramework_NETCLRInteropConfig{},
		&NETFramework_NETCLRJitConfig{},
		&NETFramework_NETCLRLoadingConfig{},
		&NETFramework_NETCLRLocksAndThreadsConfig{},
		&NETFramework_NETCLRMemoryConfig{},
		&NETFramework_NETCLRRemotingConfig{},
		&NETFramework_NETCLRSecurityConfig{},
		&OSConfig{},
		&ProcessConfig{},
		&RemoteFxConfig{},
		&ServiceConfig{},
		&SMTPConfig{},
		&SystemConfig{},
		&TCPConfig{},
		&TerminalServicesConfig{},
		&TextFileConfig{},
		&ThermalZoneConfig{},
		&TimeConfig{},
		&VmwareConfig{},
	}
}

// Collector is the interface a collector has to implement.
type Collector interface {
	// PerfCounterNames should return perf counter keys used by the collector.
	PerfCounterNames() []string

	// Get new metrics and expose them via prometheus registry.
	Collect(ctx *ScrapeContext, ch chan<- prometheus.Metric) (err error)
}

type ScrapeContext struct {
	perfObjects map[string]*perflib.PerfObject
}

// PrepareScrapeContext creates a ScrapeContext to be used during a single scrape
func PrepareScrapeContext(collectors []Collector) (*ScrapeContext, error) {
	q := getPerfQuery(collectors) // TODO: Memoize
	objs, err := getPerflibSnapshot(q)
	if err != nil {
		return nil, err
	}

	return &ScrapeContext{objs}, nil
}

func getPerfQuery(collectors []Collector) string {
	var parts []string
	for _, c := range collectors {
		counters := c.PerfCounterNames()
		indicies := make([]string, 0, len(counters))
		for _, cn := range counters {
			indicies = append(indicies, MapCounterToIndex(cn))
		}

		parts = append(parts, strings.Join(indicies, " "))
	}
	return strings.Join(parts, " ")
}

// Names returns the names of the provided collectors.
func Names(cs []Config) []string {
	res := make([]string, len(cs))
	for i := range cs {
		res[i] = cs[i].Name()
	}
	return res
}

func boolToFloat(b bool) float64 {
	if b {
		return 1.0
	}
	return 0.0
}

func find(slice []string, val string) bool {
	for _, item := range slice {
		if item == val {
			return true
		}
	}
	return false
}

// Used by more complex collectors where user input specifies enabled child collectors.
// Splits provided child collectors and deduplicate.
func expandEnabledChildCollectors(enabled string) []string {
	separated := strings.Split(enabled, ",")
	unique := map[string]bool{}
	for _, s := range separated {
		if s != "" {
			unique[s] = true
		}
	}
	result := make([]string, 0, len(unique))
	for s := range unique {
		result = append(result, s)
	}
	// Ensure result is ordered, to prevent test failure
	sort.Strings(result)
	return result
}
