//go:build linux
// +build linux

package perf

import (
	"sync"

	"go.uber.org/multierr"
	"golang.org/x/sys/unix"
)

type SoftwareProfilerType int

const (
	AllSoftwareProfilers  SoftwareProfilerType = 0
	CpuClockProfiler      SoftwareProfilerType = 1 << iota
	TaskClockProfiler     SoftwareProfilerType = 1 << iota
	PageFaultProfiler     SoftwareProfilerType = 1 << iota
	ContextSwitchProfiler SoftwareProfilerType = 1 << iota
	CpuMigrationProfiler  SoftwareProfilerType = 1 << iota
	MinorFaultProfiler    SoftwareProfilerType = 1 << iota
	MajorFaultProfiler    SoftwareProfilerType = 1 << iota
	AlignFaultProfiler    SoftwareProfilerType = 1 << iota
	EmuFaultProfiler      SoftwareProfilerType = 1 << iota
)

type softwareProfiler struct {
	// map of perf counter type to file descriptor
	profilers   map[int]Profiler
	profilersMu sync.RWMutex
}

// NewSoftwareProfiler returns a new software profiler.
func NewSoftwareProfiler(pid, cpu int, profilerSet SoftwareProfilerType, opts ...int) (SoftwareProfiler, error) {
	var e error
	profilers := map[int]Profiler{}

	if profilerSet&CpuClockProfiler > 0 || profilerSet == AllSoftwareProfilers {
		cpuClockProfiler, err := NewCPUClockProfiler(pid, cpu, opts...)
		if err != nil {
			e = multierr.Append(e, err)
		} else {
			profilers[unix.PERF_COUNT_SW_CPU_CLOCK] = cpuClockProfiler
		}
	}

	if profilerSet&TaskClockProfiler > 0 || profilerSet == AllSoftwareProfilers {
		taskClockProfiler, err := NewTaskClockProfiler(pid, cpu, opts...)
		if err != nil {
			e = multierr.Append(e, err)
		} else {
			profilers[unix.PERF_COUNT_SW_TASK_CLOCK] = taskClockProfiler
		}
	}

	if profilerSet&PageFaultProfiler > 0 || profilerSet == AllSoftwareProfilers {
		pageFaultProfiler, err := NewPageFaultProfiler(pid, cpu, opts...)
		if err != nil {
			e = multierr.Append(e, err)
		} else {
			profilers[unix.PERF_COUNT_SW_PAGE_FAULTS] = pageFaultProfiler
		}
	}

	if profilerSet&ContextSwitchProfiler > 0 || profilerSet == AllSoftwareProfilers {
		ctxSwitchesProfiler, err := NewCtxSwitchesProfiler(pid, cpu, opts...)
		if err != nil {
			e = multierr.Append(e, err)
		} else {
			profilers[unix.PERF_COUNT_SW_CONTEXT_SWITCHES] = ctxSwitchesProfiler
		}
	}

	if profilerSet&CpuMigrationProfiler > 0 || profilerSet == AllSoftwareProfilers {
		cpuMigrationsProfiler, err := NewCPUMigrationsProfiler(pid, cpu, opts...)
		if err != nil {
			e = multierr.Append(e, err)
		} else {
			profilers[unix.PERF_COUNT_SW_CPU_MIGRATIONS] = cpuMigrationsProfiler
		}
	}

	if profilerSet&MinorFaultProfiler > 0 || profilerSet == AllSoftwareProfilers {
		minorFaultProfiler, err := NewMinorFaultsProfiler(pid, cpu, opts...)
		if err != nil {
			e = multierr.Append(e, err)
		} else {
			profilers[unix.PERF_COUNT_SW_PAGE_FAULTS_MIN] = minorFaultProfiler
		}
	}

	if profilerSet&MajorFaultProfiler > 0 || profilerSet == AllSoftwareProfilers {
		majorFaultProfiler, err := NewMajorFaultsProfiler(pid, cpu, opts...)
		if err != nil {
			e = multierr.Append(e, err)
		} else {
			profilers[unix.PERF_COUNT_SW_PAGE_FAULTS_MAJ] = majorFaultProfiler
		}
	}

	if profilerSet&AlignFaultProfiler > 0 || profilerSet == AllSoftwareProfilers {
		alignFaultsFrontProfiler, err := NewAlignFaultsProfiler(pid, cpu, opts...)
		if err != nil {
			e = multierr.Append(e, err)
		} else {
			profilers[unix.PERF_COUNT_SW_ALIGNMENT_FAULTS] = alignFaultsFrontProfiler
		}
	}

	if profilerSet&EmuFaultProfiler > 0 || profilerSet == AllSoftwareProfilers {
		emuFaultProfiler, err := NewEmulationFaultsProfiler(pid, cpu, opts...)
		if err != nil {
			e = multierr.Append(e, err)
		} else {
			profilers[unix.PERF_COUNT_SW_EMULATION_FAULTS] = emuFaultProfiler
		}
	}

	return &softwareProfiler{
		profilers: profilers,
	}, e
}

// HasProfilers returns if there are any configured profilers.
func (p *softwareProfiler) HasProfilers() bool {
	p.profilersMu.RLock()
	defer p.profilersMu.RUnlock()
	return len(p.profilers) >= 0
}

// Start is used to start the SoftwareProfiler.
func (p *softwareProfiler) Start() error {
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

// Reset is used to reset the SoftwareProfiler.
func (p *softwareProfiler) Reset() error {
	var err error
	p.profilersMu.RLock()
	for _, profiler := range p.profilers {
		err = multierr.Append(err, profiler.Reset())
	}
	p.profilersMu.RUnlock()
	return err
}

// Stop is used to reset the SoftwareProfiler.
func (p *softwareProfiler) Stop() error {
	var err error
	p.profilersMu.RLock()
	for _, profiler := range p.profilers {
		err = multierr.Append(err, profiler.Stop())
	}
	p.profilersMu.RUnlock()
	return err
}

// Close is used to reset the SoftwareProfiler.
func (p *softwareProfiler) Close() error {
	var err error
	p.profilersMu.RLock()
	for _, profiler := range p.profilers {
		err = multierr.Append(err, profiler.Close())
	}
	p.profilersMu.RUnlock()
	return err
}

// Profile is used to read the SoftwareProfiler SoftwareProfile it returns an
// error only if all profiles fail.
func (p *softwareProfiler) Profile(swProfile *SoftwareProfile) error {
	var err error
	swProfile.Reset()
	p.profilersMu.RLock()
	for profilerType, profiler := range p.profilers {
		profileVal := ProfileValuePool.Get().(*ProfileValue)
		err2 := profiler.Profile(profileVal)
		err = multierr.Append(err, err2)
		if err2 == nil {
			if swProfile.TimeEnabled == nil {
				swProfile.TimeEnabled = &profileVal.TimeEnabled
			}
			if swProfile.TimeRunning == nil {
				swProfile.TimeRunning = &profileVal.TimeRunning
			}
			switch profilerType {
			case unix.PERF_COUNT_SW_CPU_CLOCK:
				swProfile.CPUClock = &profileVal.Value
			case unix.PERF_COUNT_SW_TASK_CLOCK:
				swProfile.TaskClock = &profileVal.Value
			case unix.PERF_COUNT_SW_PAGE_FAULTS:
				swProfile.PageFaults = &profileVal.Value
			case unix.PERF_COUNT_SW_CONTEXT_SWITCHES:
				swProfile.ContextSwitches = &profileVal.Value
			case unix.PERF_COUNT_SW_CPU_MIGRATIONS:
				swProfile.CPUMigrations = &profileVal.Value
			case unix.PERF_COUNT_SW_PAGE_FAULTS_MIN:
				swProfile.MinorPageFaults = &profileVal.Value
			case unix.PERF_COUNT_SW_PAGE_FAULTS_MAJ:
				swProfile.MajorPageFaults = &profileVal.Value
			case unix.PERF_COUNT_SW_ALIGNMENT_FAULTS:
				swProfile.AlignmentFaults = &profileVal.Value
			case unix.PERF_COUNT_SW_EMULATION_FAULTS:
				swProfile.EmulationFaults = &profileVal.Value
			default:
			}
		}
	}
	p.profilersMu.RUnlock()
	return nil
}
