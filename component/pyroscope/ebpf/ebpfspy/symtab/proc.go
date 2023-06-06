package symtab

import (
	"debug/elf"
	"fmt"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"golang.org/x/exp/slices"
)

type ProcTable struct {
	logger     log.Logger
	ranges     []elfRange
	file2Table map[file]*ElfTable
	options    ProcTableOptions
	rootFS     string
}

type ProcTableOptions struct {
	Pid int
	ElfTableOptions
}

func NewProcTable(logger log.Logger, options ProcTableOptions) *ProcTable {
	return &ProcTable{
		logger:     logger,
		file2Table: make(map[file]*ElfTable),
		options:    options,
		rootFS:     path.Join("/proc", strconv.Itoa(options.Pid), "root"),
	}
}

type elfRange struct {
	mapRange *ProcMap
	// may be nil
	elfTable *ElfTable
}

func (p *ProcTable) Refresh() {
	procMaps, err := os.ReadFile(fmt.Sprintf("/proc/%d/maps", p.options.Pid))
	if err != nil {
		return // todo return err
	}
	p.refresh(string(procMaps))
}

func (p *ProcTable) refresh(procMaps string) {
	// todo support perf map files
	for i := range p.ranges {
		p.ranges[i].elfTable = nil
	}
	p.ranges = p.ranges[:0]
	filesToKeep := make(map[file]struct{})
	maps, err := parseProcMapsExecutableModules(procMaps)
	if err != nil {
		return
	}
	for _, m := range maps {
		p.ranges = append(p.ranges, elfRange{
			mapRange: m,
		})
		r := &p.ranges[len(p.ranges)-1]
		e := p.getElfTable(r)
		if e != nil {
			r.elfTable = e
			filesToKeep[r.mapRange.file()] = struct{}{}
		}
	}
	var filesToDelete []file
	for f := range p.file2Table {
		_, keep := filesToKeep[f]
		if !keep {
			filesToDelete = append(filesToDelete, f)
		}
	}
	for _, f := range filesToDelete {
		delete(p.file2Table, f)
	}
}

func (p *ProcTable) getElfTable(r *elfRange) *ElfTable {
	f := r.mapRange.file()
	e, ok := p.file2Table[f]
	if !ok {
		e = p.createElfTable(r)
		if e != nil {
			p.file2Table[f] = e
		}
	}
	return e
}

func (p *ProcTable) Resolve(pc uint64) Symbol {
	i, found := slices.BinarySearchFunc(p.ranges, pc, binarySearchElfRange)
	if !found {
		return Symbol{}
	}
	r := p.ranges[i]
	t := r.elfTable
	if t == nil {
		return Symbol{}
	}
	s := t.Resolve(pc)
	if s == nil {
		moduleOffset := pc - t.base
		return Symbol{Start: moduleOffset, Module: r.mapRange.Pathname}
	}

	return Symbol{Start: s.Start, Name: s.Name, Module: r.mapRange.Pathname}
}

func (*ProcTable) Close() {
}

func (p *ProcTable) createElfTable(m *elfRange) *ElfTable {
	if !strings.HasPrefix(m.mapRange.Pathname, "/") {
		return nil
	}
	file := m.mapRange.Pathname
	e, err := NewElfTable(p.logger, p.rootFS, file, p.options.ElfTableOptions)

	if err != nil {
		level.Debug(p.logger).Log(
			"msg", "elf table creation failed",
			"err", err,
			"file", file,
			"fs", p.rootFS,
		)
		return nil
	}

	if p.rebase(m, e) {
		return e
	}
	level.Error(p.logger).Log(
		"msg", "failed to find a base for elf table",
		"file", file,
		"fs", p.rootFS,
	)
	return nil
}

func (p *ProcTable) rebase(m *elfRange, e *ElfTable) bool {
	if e.typ == elf.ET_EXEC {
		return true
	}
	for _, executable := range e.executables {
		if uint64(m.mapRange.Offset) == executable.Off {
			base := m.mapRange.StartAddr - executable.Vaddr
			e.Rebase(base)
			return true
		}
	}
	return false
}

func binarySearchElfRange(e elfRange, pc uint64) int {
	if pc < e.mapRange.StartAddr {
		return 1
	}
	if pc >= e.mapRange.EndAddr {
		return -1
	}
	return 0
}
