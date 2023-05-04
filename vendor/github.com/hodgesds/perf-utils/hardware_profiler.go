//go:build linux
// +build linux

package perf

import (
	"fmt"
	"sync"

	"go.uber.org/multierr"
	"golang.org/x/sys/unix"
)

type HardwareProfilerType int

const (
	AllHardwareProfilers          HardwareProfilerType = 0
	CpuCyclesProfiler             HardwareProfilerType = 1 << iota
	CpuInstrProfiler              HardwareProfilerType = 1 << iota
	CacheRefProfiler              HardwareProfilerType = 1 << iota
	CacheMissesProfiler           HardwareProfilerType = 1 << iota
	BranchInstrProfiler           HardwareProfilerType = 1 << iota
	BranchMissesProfiler          HardwareProfilerType = 1 << iota
	BusCyclesProfiler             HardwareProfilerType = 1 << iota
	StalledCyclesBackendProfiler  HardwareProfilerType = 1 << iota
	StalledCyclesFrontendProfiler HardwareProfilerType = 1 << iota
	RefCpuCyclesProfiler          HardwareProfilerType = 1 << iota
)

type hardwareProfiler struct {
	// map of perf counter type to file descriptor
	profilers   map[int]Profiler
	profilersMu sync.RWMutex
}

// NewHardwareProfiler returns a new hardware profiler.
func NewHardwareProfiler(pid, cpu int, profilerSet HardwareProfilerType, opts ...int) (HardwareProfiler, error) {
	var e error
	profilers := map[int]Profiler{}

	if profilerSet&CpuCyclesProfiler > 0 || profilerSet == AllHardwareProfilers {
		cpuCycleProfiler, err := NewCPUCycleProfiler(pid, cpu, opts...)
		if err != nil {
			e = multierr.Append(e,
				fmt.Errorf("Failed to setup CPU cycle profiler: pid (%d) cpu (%d) %q", pid, cpu, err))
		} else {
			profilers[unix.PERF_COUNT_HW_CPU_CYCLES] = cpuCycleProfiler
		}
	}

	if profilerSet&CpuInstrProfiler > 0 || profilerSet == AllHardwareProfilers {
		instrProfiler, err := NewInstrProfiler(pid, cpu, opts...)
		if err != nil {
			e = multierr.Append(e,
				fmt.Errorf("Failed to CPU setup instruction profiler: pid (%d) cpu (%d) %q", pid, cpu, err))
		} else {
			profilers[unix.PERF_COUNT_HW_INSTRUCTIONS] = instrProfiler
		}
	}

	if profilerSet&CacheRefProfiler > 0 || profilerSet == AllHardwareProfilers {
		cacheRefProfiler, err := NewCacheRefProfiler(pid, cpu, opts...)
		if err != nil {
			e = multierr.Append(e,
				fmt.Errorf("Failed to setup cache ref profiler: pid (%d) cpu (%d) %q", pid, cpu, err))
		} else {
			profilers[unix.PERF_COUNT_HW_CACHE_REFERENCES] = cacheRefProfiler
		}
	}

	if profilerSet&CacheMissesProfiler > 0 || profilerSet == AllHardwareProfilers {
		cacheMissesProfiler, err := NewCacheMissesProfiler(pid, cpu, opts...)
		if err != nil {
			e = multierr.Append(e,
				fmt.Errorf("Failed to setup cache misses profiler: pid (%d) cpu (%d) %q", pid, cpu, err))
		} else {
			profilers[unix.PERF_COUNT_HW_CACHE_MISSES] = cacheMissesProfiler
		}
	}

	if profilerSet&BranchInstrProfiler > 0 || profilerSet == AllHardwareProfilers {
		branchInstrProfiler, err := NewBranchInstrProfiler(pid, cpu, opts...)
		if err != nil {
			e = multierr.Append(e,
				fmt.Errorf("Failed to setup branch instruction profiler: pid (%d) cpu (%d) %q", pid, cpu, err))
		} else {
			profilers[unix.PERF_COUNT_HW_BRANCH_INSTRUCTIONS] = branchInstrProfiler
		}
	}

	if profilerSet&BranchMissesProfiler > 0 || profilerSet == AllHardwareProfilers {
		branchMissesProfiler, err := NewBranchMissesProfiler(pid, cpu, opts...)
		if err != nil {
			e = multierr.Append(e,
				fmt.Errorf("Failed to setup branch miss profiler: pid (%d) cpu (%d) %q", pid, cpu, err))
		} else {
			profilers[unix.PERF_COUNT_HW_BRANCH_MISSES] = branchMissesProfiler
		}
	}

	if profilerSet&BusCyclesProfiler > 0 || profilerSet == AllHardwareProfilers {
		busCyclesProfiler, err := NewBusCyclesProfiler(pid, cpu, opts...)
		if err != nil {
			e = multierr.Append(e,
				fmt.Errorf("Failed to setup bus cycles profiler: pid (%d) cpu (%d) %q", pid, cpu, err))
		} else {
			profilers[unix.PERF_COUNT_HW_BUS_CYCLES] = busCyclesProfiler
		}
	}

	if profilerSet&StalledCyclesFrontendProfiler > 0 || profilerSet == AllHardwareProfilers {
		stalledCyclesFrontProfiler, err := NewStalledCyclesFrontProfiler(pid, cpu, opts...)
		if err != nil {
			e = multierr.Append(e,
				fmt.Errorf("Failed to setup stalled fronted cycles profiler: pid (%d) cpu (%d) %q", pid, cpu, err))
		} else {
			profilers[unix.PERF_COUNT_HW_STALLED_CYCLES_FRONTEND] = stalledCyclesFrontProfiler
		}
	}

	if profilerSet&StalledCyclesBackendProfiler > 0 || profilerSet == AllHardwareProfilers {
		stalledCyclesBackProfiler, err := NewStalledCyclesBackProfiler(pid, cpu, opts...)
		if err != nil {
			e = multierr.Append(e,
				fmt.Errorf("Failed to setup stalled backend cycles profiler: pid (%d) cpu (%d) %q", pid, cpu, err))
		} else {
			profilers[unix.PERF_COUNT_HW_STALLED_CYCLES_BACKEND] = stalledCyclesBackProfiler
		}
	}

	if profilerSet&RefCpuCyclesProfiler > 0 || profilerSet == AllHardwareProfilers {
		refCPUCyclesProfiler, err := NewRefCPUCyclesProfiler(pid, cpu, opts...)
		if err != nil {
			e = multierr.Append(e,
				fmt.Errorf("Failed to setup ref CPU cycles profiler: pid (%d) cpu (%d) %q", pid, cpu, err))
		} else {
			profilers[unix.PERF_COUNT_HW_REF_CPU_CYCLES] = refCPUCyclesProfiler
		}
	}

	return &hardwareProfiler{
		profilers: profilers,
	}, e
}

// HasProfilers returns if there are any configured profilers.
func (p *hardwareProfiler) HasProfilers() bool {
	p.profilersMu.RLock()
	defer p.profilersMu.RUnlock()
	return len(p.profilers) >= 0
}

// Start is used to start the HardwareProfiler.
func (p *hardwareProfiler) Start() error {
	if !p.HasProfilers() {
		return ErrNoProfiler
	}
	var err error
	p.profilersMu.RLock()
	for _, profiler := range p.profilers {
		err = multierr.Append(err, profiler.Start())
	}
	p.profilersMu.RUnlock()
	return err
}

// Reset is used to reset the HardwareProfiler.
func (p *hardwareProfiler) Reset() error {
	var err error
	p.profilersMu.RLock()
	for _, profiler := range p.profilers {
		err = multierr.Append(err, profiler.Reset())
	}
	p.profilersMu.RUnlock()
	return err
}

// Stop is used to reset the HardwareProfiler.
func (p *hardwareProfiler) Stop() error {
	var err error
	p.profilersMu.RLock()
	for _, profiler := range p.profilers {
		err = multierr.Append(err, profiler.Stop())
	}
	p.profilersMu.RUnlock()
	return err
}

// Close is used to reset the HardwareProfiler.
func (p *hardwareProfiler) Close() error {
	var err error
	p.profilersMu.RLock()
	for _, profiler := range p.profilers {
		err = multierr.Append(err, profiler.Close())
	}
	p.profilersMu.RUnlock()
	return err
}

// Profile is used to read the HardwareProfiler HardwareProfile it returns an
// error only if all profiles fail.
func (p *hardwareProfiler) Profile(hwProfile *HardwareProfile) error {
	var err error
	hwProfile.Reset()
	p.profilersMu.RLock()
	for profilerType, profiler := range p.profilers {
		profileVal := ProfileValuePool.Get().(*ProfileValue)
		err2 := profiler.Profile(profileVal)
		err = multierr.Append(err, err2)
		if err2 == nil {
			if hwProfile.TimeEnabled == nil {
				hwProfile.TimeEnabled = &profileVal.TimeEnabled
			}
			if hwProfile.TimeRunning == nil {
				hwProfile.TimeRunning = &profileVal.TimeRunning
			}
			switch profilerType {
			case unix.PERF_COUNT_HW_CPU_CYCLES:
				hwProfile.CPUCycles = &profileVal.Value
			case unix.PERF_COUNT_HW_INSTRUCTIONS:
				hwProfile.Instructions = &profileVal.Value
			case unix.PERF_COUNT_HW_CACHE_REFERENCES:
				hwProfile.CacheRefs = &profileVal.Value
			case unix.PERF_COUNT_HW_CACHE_MISSES:
				hwProfile.CacheMisses = &profileVal.Value
			case unix.PERF_COUNT_HW_BRANCH_INSTRUCTIONS:
				hwProfile.BranchInstr = &profileVal.Value
			case unix.PERF_COUNT_HW_BRANCH_MISSES:
				hwProfile.BranchMisses = &profileVal.Value
			case unix.PERF_COUNT_HW_BUS_CYCLES:
				hwProfile.BusCycles = &profileVal.Value
			case unix.PERF_COUNT_HW_STALLED_CYCLES_FRONTEND:
				hwProfile.StalledCyclesFrontend = &profileVal.Value
			case unix.PERF_COUNT_HW_STALLED_CYCLES_BACKEND:
				hwProfile.StalledCyclesBackend = &profileVal.Value
			case unix.PERF_COUNT_HW_REF_CPU_CYCLES:
				hwProfile.RefCPUCycles = &profileVal.Value
			}
		}
	}
	p.profilersMu.RUnlock()
	return err
}
