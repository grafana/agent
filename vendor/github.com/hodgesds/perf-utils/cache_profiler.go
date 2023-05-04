//go:build linux
// +build linux

package perf

import (
	"fmt"
	"sync"

	"go.uber.org/multierr"
	"golang.org/x/sys/unix"
)

type CacheProfilerType int

const (
	// AllCacheProfilers is used to try to configure all cache profilers.
	AllCacheProfilers          CacheProfilerType = 0
	L1DataReadHitProfiler      CacheProfilerType = 1 << iota
	L1DataReadMissProfiler     CacheProfilerType = 1 << iota
	L1DataWriteHitProfiler     CacheProfilerType = 1 << iota
	L1InstrReadMissProfiler    CacheProfilerType = 1 << iota
	L1InstrReadHitProfiler     CacheProfilerType = 1 << iota
	LLReadHitProfiler          CacheProfilerType = 1 << iota
	LLReadMissProfiler         CacheProfilerType = 1 << iota
	LLWriteHitProfiler         CacheProfilerType = 1 << iota
	LLWriteMissProfiler        CacheProfilerType = 1 << iota
	DataTLBReadHitProfiler     CacheProfilerType = 1 << iota
	DataTLBReadMissProfiler    CacheProfilerType = 1 << iota
	DataTLBWriteHitProfiler    CacheProfilerType = 1 << iota
	DataTLBWriteMissProfiler   CacheProfilerType = 1 << iota
	InstrTLBReadHitProfiler    CacheProfilerType = 1 << iota
	InstrTLBReadMissProfiler   CacheProfilerType = 1 << iota
	BPUReadHitProfiler         CacheProfilerType = 1 << iota
	BPUReadMissProfiler        CacheProfilerType = 1 << iota
	NodeCacheReadHitProfiler   CacheProfilerType = 1 << iota
	NodeCacheReadMissProfiler  CacheProfilerType = 1 << iota
	NodeCacheWriteHitProfiler  CacheProfilerType = 1 << iota
	NodeCacheWriteMissProfiler CacheProfilerType = 1 << iota

	// L1DataReadHit is a constant...
	L1DataReadHit = (unix.PERF_COUNT_HW_CACHE_L1D) | (unix.PERF_COUNT_HW_CACHE_OP_READ << 8) | (unix.PERF_COUNT_HW_CACHE_RESULT_ACCESS << 16)
	// L1DataReadMiss is a constant...
	L1DataReadMiss = (unix.PERF_COUNT_HW_CACHE_L1D) | (unix.PERF_COUNT_HW_CACHE_OP_READ << 8) | (unix.PERF_COUNT_HW_CACHE_RESULT_MISS << 16)
	// L1DataWriteHit is a constant...
	L1DataWriteHit = (unix.PERF_COUNT_HW_CACHE_L1D) | (unix.PERF_COUNT_HW_CACHE_OP_WRITE << 8) | (unix.PERF_COUNT_HW_CACHE_RESULT_ACCESS << 16)
	// L1InstrReadMiss is a constant...
	L1InstrReadMiss = (unix.PERF_COUNT_HW_CACHE_L1I) | (unix.PERF_COUNT_HW_CACHE_OP_READ << 8) | (unix.PERF_COUNT_HW_CACHE_RESULT_MISS << 16)
	// L1InstrReadHit is a constant...
	L1InstrReadHit = (unix.PERF_COUNT_HW_CACHE_L1I) | (unix.PERF_COUNT_HW_CACHE_OP_READ << 8) | (unix.PERF_COUNT_HW_CACHE_RESULT_ACCESS << 16)

	// LLReadHit is a constant...
	LLReadHit = (unix.PERF_COUNT_HW_CACHE_LL) | (unix.PERF_COUNT_HW_CACHE_OP_READ << 8) | (unix.PERF_COUNT_HW_CACHE_RESULT_ACCESS << 16)
	// LLReadMiss is a constant...
	LLReadMiss = (unix.PERF_COUNT_HW_CACHE_LL) | (unix.PERF_COUNT_HW_CACHE_OP_READ << 8) | (unix.PERF_COUNT_HW_CACHE_RESULT_MISS << 16)
	// LLWriteHit is a constant...
	LLWriteHit = (unix.PERF_COUNT_HW_CACHE_LL) | (unix.PERF_COUNT_HW_CACHE_OP_WRITE << 8) | (unix.PERF_COUNT_HW_CACHE_RESULT_ACCESS << 16)
	// LLWriteMiss is a constant...
	LLWriteMiss = (unix.PERF_COUNT_HW_CACHE_LL) | (unix.PERF_COUNT_HW_CACHE_OP_WRITE << 8) | (unix.PERF_COUNT_HW_CACHE_RESULT_MISS << 16)

	// DataTLBReadHit is a constant...
	DataTLBReadHit = (unix.PERF_COUNT_HW_CACHE_DTLB) | (unix.PERF_COUNT_HW_CACHE_OP_READ << 8) | (unix.PERF_COUNT_HW_CACHE_RESULT_ACCESS << 16)
	// DataTLBReadMiss is a constant...
	DataTLBReadMiss = (unix.PERF_COUNT_HW_CACHE_DTLB) | (unix.PERF_COUNT_HW_CACHE_OP_READ << 8) | (unix.PERF_COUNT_HW_CACHE_RESULT_MISS << 16)
	// DataTLBWriteHit is a constant...
	DataTLBWriteHit = (unix.PERF_COUNT_HW_CACHE_DTLB) | (unix.PERF_COUNT_HW_CACHE_OP_WRITE << 8) | (unix.PERF_COUNT_HW_CACHE_RESULT_ACCESS << 16)
	// DataTLBWriteMiss is a constant...
	DataTLBWriteMiss = (unix.PERF_COUNT_HW_CACHE_DTLB) | (unix.PERF_COUNT_HW_CACHE_OP_WRITE << 8) | (unix.PERF_COUNT_HW_CACHE_RESULT_MISS << 16)

	// InstrTLBReadHit is a constant...
	InstrTLBReadHit = (unix.PERF_COUNT_HW_CACHE_ITLB) | (unix.PERF_COUNT_HW_CACHE_OP_READ << 8) | (unix.PERF_COUNT_HW_CACHE_RESULT_ACCESS << 16)
	// InstrTLBReadMiss is a constant...
	InstrTLBReadMiss = (unix.PERF_COUNT_HW_CACHE_ITLB) | (unix.PERF_COUNT_HW_CACHE_OP_READ << 8) | (unix.PERF_COUNT_HW_CACHE_RESULT_MISS << 16)

	// BPUReadHit is a constant...
	BPUReadHit = (unix.PERF_COUNT_HW_CACHE_BPU) | (unix.PERF_COUNT_HW_CACHE_OP_READ << 8) | (unix.PERF_COUNT_HW_CACHE_RESULT_ACCESS << 16)
	// BPUReadMiss is a constant...
	BPUReadMiss = (unix.PERF_COUNT_HW_CACHE_BPU) | (unix.PERF_COUNT_HW_CACHE_OP_READ << 8) | (unix.PERF_COUNT_HW_CACHE_RESULT_MISS << 16)

	// NodeCacheReadHit is a constant...
	NodeCacheReadHit = (unix.PERF_COUNT_HW_CACHE_NODE) | (unix.PERF_COUNT_HW_CACHE_OP_READ << 8) | (unix.PERF_COUNT_HW_CACHE_RESULT_ACCESS << 16)
	// NodeCacheReadMiss is a constant...
	NodeCacheReadMiss = (unix.PERF_COUNT_HW_CACHE_NODE) | (unix.PERF_COUNT_HW_CACHE_OP_READ << 8) | (unix.PERF_COUNT_HW_CACHE_RESULT_MISS << 16)
	// NodeCacheWriteHit is a constant...
	NodeCacheWriteHit = (unix.PERF_COUNT_HW_CACHE_NODE) | (unix.PERF_COUNT_HW_CACHE_OP_WRITE << 8) | (unix.PERF_COUNT_HW_CACHE_RESULT_ACCESS << 16)
	// NodeCacheWriteMiss is a constant...
	NodeCacheWriteMiss = (unix.PERF_COUNT_HW_CACHE_NODE) | (unix.PERF_COUNT_HW_CACHE_OP_WRITE << 8) | (unix.PERF_COUNT_HW_CACHE_RESULT_MISS << 16)
)

type cacheProfiler struct {
	// map of perf counter type to file descriptor
	profilers   map[int]Profiler
	profilersMu sync.RWMutex
}

// NewCacheProfiler returns a new cache profiler.
func NewCacheProfiler(pid, cpu int, profilerSet CacheProfilerType, opts ...int) (CacheProfiler, error) {
	profilers := map[int]Profiler{}
	var e error

	// L1 data
	if profilerSet&L1DataReadHitProfiler > 0 || profilerSet == AllCacheProfilers {
		op := unix.PERF_COUNT_HW_CACHE_OP_READ
		result := unix.PERF_COUNT_HW_CACHE_RESULT_ACCESS
		l1dataReadHit, err := NewL1DataProfiler(pid, cpu, op, result, opts...)
		if err != nil {
			e = multierr.Append(e,
				fmt.Errorf("Failed to setup L1 data read hit profiler: pid (%d) cpu (%d) %q", pid, cpu, err))
		} else {
			profilers[L1DataReadHit] = l1dataReadHit
		}
	}

	if profilerSet&L1DataReadMissProfiler > 0 || profilerSet == AllCacheProfilers {
		op := unix.PERF_COUNT_HW_CACHE_OP_READ
		result := unix.PERF_COUNT_HW_CACHE_RESULT_MISS
		l1dataReadMiss, err := NewL1DataProfiler(pid, cpu, op, result, opts...)
		if err != nil {
			e = multierr.Append(e,
				fmt.Errorf("Failed to setup L1 data read miss profiler: pid (%d) cpu (%d) %q", pid, cpu, err))
		} else {
			profilers[L1DataReadMiss] = l1dataReadMiss
		}
	}

	if profilerSet&L1DataWriteHitProfiler > 0 || profilerSet == AllCacheProfilers {
		op := unix.PERF_COUNT_HW_CACHE_OP_WRITE
		result := unix.PERF_COUNT_HW_CACHE_RESULT_ACCESS
		l1dataWriteHit, err := NewL1DataProfiler(pid, cpu, op, result, opts...)
		if err != nil {
			e = multierr.Append(e,
				fmt.Errorf("Failed to setup L1 data write profiler: pid (%d) cpu (%d) %q", pid, cpu, err))
		} else {
			profilers[L1DataWriteHit] = l1dataWriteHit
		}
	}

	// L1 instruction
	if profilerSet&L1InstrReadHitProfiler > 0 || profilerSet == AllCacheProfilers {
		op := unix.PERF_COUNT_HW_CACHE_OP_READ
		result := unix.PERF_COUNT_HW_CACHE_RESULT_ACCESS
		l1instrReadHit, err := NewL1InstrProfiler(pid, cpu, op, result, opts...)
		if err != nil {
			e = multierr.Append(e,
				fmt.Errorf("Failed to setup L1 instruction read hit profiler: pid (%d) cpu (%d) %q", pid, cpu, err))
		} else {
			profilers[L1InstrReadHit] = l1instrReadHit
		}
	}

	if profilerSet&L1InstrReadMissProfiler > 0 || profilerSet == AllCacheProfilers {
		op := unix.PERF_COUNT_HW_CACHE_OP_READ
		result := unix.PERF_COUNT_HW_CACHE_RESULT_MISS
		l1InstrReadMiss, err := NewL1InstrProfiler(pid, cpu, op, result, opts...)
		if err != nil {
			e = multierr.Append(e,
				fmt.Errorf("Failed to setup L1 instruction read miss profiler: pid (%d) cpu (%d) %q", pid, cpu, err))
		} else {
			profilers[L1InstrReadMiss] = l1InstrReadMiss
		}
	}

	// Last Level
	if profilerSet&LLReadHitProfiler > 0 || profilerSet == AllCacheProfilers {
		op := unix.PERF_COUNT_HW_CACHE_OP_READ
		result := unix.PERF_COUNT_HW_CACHE_RESULT_ACCESS
		llReadHit, err := NewLLCacheProfiler(pid, cpu, op, result, opts...)
		if err != nil {
			e = multierr.Append(e,
				fmt.Errorf("Failed to setup last level read hit profiler: pid (%d) cpu (%d) %q", pid, cpu, err))
		} else {
			profilers[LLReadHit] = llReadHit
		}
	}

	if profilerSet&LLReadMissProfiler > 0 || profilerSet == AllCacheProfilers {
		op := unix.PERF_COUNT_HW_CACHE_OP_READ
		result := unix.PERF_COUNT_HW_CACHE_RESULT_MISS
		llReadMiss, err := NewLLCacheProfiler(pid, cpu, op, result, opts...)
		if err != nil {
			e = multierr.Append(e,
				fmt.Errorf("Failed to setup last level read miss profiler: pid (%d) cpu (%d) %q", pid, cpu, err))
		} else {
			profilers[LLReadMiss] = llReadMiss
		}
	}

	if profilerSet&LLWriteHitProfiler > 0 || profilerSet == AllCacheProfilers {
		op := unix.PERF_COUNT_HW_CACHE_OP_WRITE
		result := unix.PERF_COUNT_HW_CACHE_RESULT_ACCESS
		llWriteHit, err := NewLLCacheProfiler(pid, cpu, op, result, opts...)
		if err != nil {
			e = multierr.Append(e,
				fmt.Errorf("Failed to setup last level write hit profiler: pid (%d) cpu (%d) %q", pid, cpu, err))
		} else {
			profilers[LLWriteHit] = llWriteHit
		}
	}

	if profilerSet&LLWriteMissProfiler > 0 || profilerSet == AllCacheProfilers {
		op := unix.PERF_COUNT_HW_CACHE_OP_WRITE
		result := unix.PERF_COUNT_HW_CACHE_RESULT_MISS
		llWriteMiss, err := NewLLCacheProfiler(pid, cpu, op, result, opts...)
		if err != nil {
			e = multierr.Append(e,
				fmt.Errorf("Failed to setup last level write miss profiler: pid (%d) cpu (%d) %q", pid, cpu, err))
		} else {
			profilers[LLWriteMiss] = llWriteMiss
		}
	}

	// dTLB
	if profilerSet&DataTLBReadHitProfiler > 0 || profilerSet == AllCacheProfilers {
		op := unix.PERF_COUNT_HW_CACHE_OP_READ
		result := unix.PERF_COUNT_HW_CACHE_RESULT_ACCESS
		dTLBReadHit, err := NewDataTLBProfiler(pid, cpu, op, result, opts...)
		if err != nil {
			e = multierr.Append(e, fmt.Errorf("Failed to setup dTLB read hit profiler: pid (%d) cpu (%d) %q", pid, cpu, err))
		} else {
			profilers[DataTLBReadHit] = dTLBReadHit
		}
	}

	if profilerSet&DataTLBReadMissProfiler > 0 || profilerSet == AllCacheProfilers {
		op := unix.PERF_COUNT_HW_CACHE_OP_READ
		result := unix.PERF_COUNT_HW_CACHE_RESULT_MISS
		dTLBReadMiss, err := NewDataTLBProfiler(pid, cpu, op, result, opts...)
		if err != nil {
			e = multierr.Append(e, fmt.Errorf(
				"Failed to setup dTLB read miss profiler: pid (%d) cpu (%d) %q", pid, cpu, err))
		} else {
			profilers[DataTLBReadMiss] = dTLBReadMiss
		}
	}

	if profilerSet&DataTLBWriteHitProfiler > 0 || profilerSet == AllCacheProfilers {
		op := unix.PERF_COUNT_HW_CACHE_OP_WRITE
		result := unix.PERF_COUNT_HW_CACHE_RESULT_ACCESS
		dTLBWriteHit, err := NewDataTLBProfiler(pid, cpu, op, result, opts...)
		if err != nil {
			e = multierr.Append(e, fmt.Errorf(
				"Failed to setup dTLB write hit profiler: pid (%d) cpu (%d) %q", pid, cpu, err))
		} else {
			profilers[DataTLBWriteHit] = dTLBWriteHit
		}
	}

	if profilerSet&DataTLBWriteMissProfiler > 0 || profilerSet == AllCacheProfilers {
		op := unix.PERF_COUNT_HW_CACHE_OP_WRITE
		result := unix.PERF_COUNT_HW_CACHE_RESULT_MISS
		dTLBWriteMiss, err := NewDataTLBProfiler(pid, cpu, op, result, opts...)
		if err != nil {
			e = multierr.Append(e, fmt.Errorf("Failed to setup dTLB write miss profiler: pid (%d) cpu (%d) %q", pid, cpu, err))
		} else {
			profilers[DataTLBWriteMiss] = dTLBWriteMiss
		}
	}

	// iTLB
	if profilerSet&InstrTLBReadHitProfiler > 0 || profilerSet == AllCacheProfilers {
		op := unix.PERF_COUNT_HW_CACHE_OP_READ
		result := unix.PERF_COUNT_HW_CACHE_RESULT_ACCESS
		iTLBReadHit, err := NewInstrTLBProfiler(pid, cpu, op, result, opts...)
		if err != nil {
			e = multierr.Append(e,
				fmt.Errorf(
					"Failed to setup iTLB read hit profiler: pid (%d) cpu (%d) %q", pid, cpu, err))
		} else {
			profilers[InstrTLBReadHit] = iTLBReadHit
		}
	}

	if profilerSet&InstrTLBReadMissProfiler > 0 || profilerSet == AllCacheProfilers {
		op := unix.PERF_COUNT_HW_CACHE_OP_READ
		result := unix.PERF_COUNT_HW_CACHE_RESULT_MISS
		iTLBReadMiss, err := NewInstrTLBProfiler(pid, cpu, op, result, opts...)
		if err != nil {
			e = multierr.Append(e, fmt.Errorf("Failed to setup iTLB read miss profiler: pid (%d) cpu (%d) %q", pid, cpu, err))
		} else {
			profilers[InstrTLBReadMiss] = iTLBReadMiss
		}
	}

	// BPU
	if profilerSet&BPUReadHitProfiler > 0 || profilerSet == AllCacheProfilers {
		op := unix.PERF_COUNT_HW_CACHE_OP_READ
		result := unix.PERF_COUNT_HW_CACHE_RESULT_ACCESS
		bpuReadHit, err := NewBPUProfiler(pid, cpu, op, result, opts...)
		if err != nil {
			e = multierr.Append(e,
				fmt.Errorf(
					"Failed to setup BPU read hit profiler: pid (%d) cpu (%d) %q", pid, cpu, err))
		} else {
			profilers[BPUReadHit] = bpuReadHit
		}
	}

	if profilerSet&BPUReadMissProfiler > 0 || profilerSet == AllCacheProfilers {
		op := unix.PERF_COUNT_HW_CACHE_OP_READ
		result := unix.PERF_COUNT_HW_CACHE_RESULT_MISS
		bpuReadMiss, err := NewBPUProfiler(pid, cpu, op, result, opts...)
		if err != nil {
			e = multierr.Append(e,
				fmt.Errorf(
					"Failed to setup BPU read miss profiler: pid (%d) cpu (%d) %q", pid, cpu, err))
		} else {
			profilers[BPUReadMiss] = bpuReadMiss
		}
	}

	// Node
	if profilerSet&NodeCacheReadHitProfiler > 0 || profilerSet == AllCacheProfilers {
		op := unix.PERF_COUNT_HW_CACHE_OP_READ
		result := unix.PERF_COUNT_HW_CACHE_RESULT_ACCESS
		nodeReadHit, err := NewNodeCacheProfiler(pid, cpu, op, result, opts...)
		if err != nil {
			e = multierr.Append(e,
				fmt.Errorf(
					"Failed to setup node cache read hit profiler: pid (%d) cpu (%d) %q", pid, cpu, err))
		} else {
			profilers[NodeCacheReadHit] = nodeReadHit
		}
	}

	if profilerSet&NodeCacheReadMissProfiler > 0 || profilerSet == AllCacheProfilers {
		op := unix.PERF_COUNT_HW_CACHE_OP_READ
		result := unix.PERF_COUNT_HW_CACHE_RESULT_MISS
		nodeReadMiss, err := NewNodeCacheProfiler(pid, cpu, op, result, opts...)
		if err != nil {
			e = multierr.Append(e,
				fmt.Errorf(
					"Failed to setup node cache read miss profiler: pid (%d) cpu (%d) %q", pid, cpu, err))
		} else {
			profilers[NodeCacheReadMiss] = nodeReadMiss
		}
	}

	return &cacheProfiler{
		profilers: profilers,
	}, e
}

// HasProfilers returns if there are any configured profilers.
func (p *cacheProfiler) HasProfilers() bool {
	p.profilersMu.RLock()
	defer p.profilersMu.RUnlock()
	return len(p.profilers) >= 0
}

// Start is used to start the CacheProfiler, it will return an error if no
// profilers are configured.
func (p *cacheProfiler) Start() error {
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

// Reset is used to reset the CacheProfiler.
func (p *cacheProfiler) Reset() error {
	var err error
	p.profilersMu.RLock()
	for _, profiler := range p.profilers {
		err = multierr.Append(err, profiler.Reset())
	}
	p.profilersMu.RUnlock()
	return err
}

// Stop is used to stop the CacheProfiler.
func (p *cacheProfiler) Stop() error {
	var err error
	p.profilersMu.RLock()
	for _, profiler := range p.profilers {
		err = multierr.Append(err, profiler.Stop())
	}
	p.profilersMu.RUnlock()
	return err
}

// Close is used to reset the CacheProfiler.
func (p *cacheProfiler) Close() error {
	var err error
	p.profilersMu.RLock()
	for _, profiler := range p.profilers {
		err = multierr.Append(err, profiler.Close())
	}
	p.profilersMu.RUnlock()
	return err
}

// Profile is used to read the CacheProfiler CacheProfile it returns an
// error only if all profiles fail.
func (p *cacheProfiler) Profile(cacheProfile *CacheProfile) error {
	var err error
	cacheProfile.Reset()
	p.profilersMu.RLock()
	for profilerType, profiler := range p.profilers {
		profileVal := ProfileValuePool.Get().(*ProfileValue)
		err2 := profiler.Profile(profileVal)
		err = multierr.Append(err, err2)
		if err2 == nil {
			if cacheProfile.TimeEnabled == nil {
				cacheProfile.TimeEnabled = &profileVal.TimeEnabled
			}
			if cacheProfile.TimeRunning == nil {
				cacheProfile.TimeRunning = &profileVal.TimeRunning
			}
			switch {
			// L1 data
			case (profilerType ^ L1DataReadHit) == 0:
				cacheProfile.L1DataReadHit = &profileVal.Value
			case (profilerType ^ L1DataReadMiss) == 0:
				cacheProfile.L1DataReadMiss = &profileVal.Value
			case (profilerType ^ L1DataWriteHit) == 0:
				cacheProfile.L1DataWriteHit = &profileVal.Value

			// L1 instruction
			case (profilerType ^ L1InstrReadMiss) == 0:
				cacheProfile.L1InstrReadMiss = &profileVal.Value

			// Last Level
			case (profilerType ^ LLReadHit) == 0:
				cacheProfile.LastLevelReadHit = &profileVal.Value
			case (profilerType ^ LLReadMiss) == 0:
				cacheProfile.LastLevelReadMiss = &profileVal.Value
			case (profilerType ^ LLWriteHit) == 0:
				cacheProfile.LastLevelWriteHit = &profileVal.Value
			case (profilerType ^ LLWriteMiss) == 0:
				cacheProfile.LastLevelWriteMiss = &profileVal.Value

			// dTLB
			case (profilerType ^ DataTLBReadHit) == 0:
				cacheProfile.DataTLBReadHit = &profileVal.Value
			case (profilerType ^ DataTLBReadMiss) == 0:
				cacheProfile.DataTLBReadMiss = &profileVal.Value
			case (profilerType ^ DataTLBWriteHit) == 0:
				cacheProfile.DataTLBWriteHit = &profileVal.Value
			case (profilerType ^ DataTLBWriteMiss) == 0:
				cacheProfile.DataTLBWriteMiss = &profileVal.Value

			// iTLB
			case (profilerType ^ InstrTLBReadHit) == 0:
				cacheProfile.InstrTLBReadHit = &profileVal.Value
			case (profilerType ^ InstrTLBReadMiss) == 0:
				cacheProfile.InstrTLBReadMiss = &profileVal.Value

			// BPU
			case (profilerType ^ BPUReadHit) == 0:
				cacheProfile.BPUReadHit = &profileVal.Value
			case (profilerType ^ BPUReadMiss) == 0:
				cacheProfile.BPUReadMiss = &profileVal.Value

			// node
			case (profilerType ^ NodeCacheReadHit) == 0:
				cacheProfile.NodeReadHit = &profileVal.Value
			case (profilerType ^ NodeCacheReadMiss) == 0:
				cacheProfile.NodeReadMiss = &profileVal.Value
			case (profilerType ^ NodeCacheWriteHit) == 0:
				cacheProfile.NodeWriteHit = &profileVal.Value
			case (profilerType ^ NodeCacheWriteMiss) == 0:
				cacheProfile.NodeWriteMiss = &profileVal.Value
			}
		}
	}
	p.profilersMu.RUnlock()
	return err
}
