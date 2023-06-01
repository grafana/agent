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
	file2Table map[string]*ElfTable
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
		file2Table: make(map[string]*ElfTable),
		options:    options,
		rootFS:     path.Join("/proc", strconv.Itoa(options.Pid), "root"),
	}
}

type elfRange struct {
	mapRange procMapEntry
	elfTable *ElfTable
}

func (p *ProcTable) Refresh() {
	procMaps, err := os.ReadFile(fmt.Sprintf("/proc/%d/maps", p.options.Pid))
	if err != nil {
		return // todo return err
	}
	p.refresh(procMaps)
}

func (p *ProcTable) refresh(procMaps []byte) {
	// todo perf map files

	// todo remove ElfTables which are no longer in mappings ranges
	for i := range p.ranges {
		p.ranges[i].elfTable = nil
	}
	p.ranges = p.ranges[:0]

	maps, err := parseProcMaps(procMaps)
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
		}
	}
}

func (p *ProcTable) getElfTable(r *elfRange) *ElfTable {
	e, ok := p.file2Table[r.mapRange.file]
	if !ok {
		e = p.createElfTable(r)
		p.file2Table[r.mapRange.file] = e
	}
	return e
}

func (p *ProcTable) Resolve(pc uint64) *Symbol {
	i, found := slices.BinarySearchFunc(p.ranges, pc, binarySearchElfRange)
	if !found {
		return nil
	}
	t := p.ranges[i].elfTable
	if t == nil {
		return nil
	}
	sym := t.Resolve(pc)
	return sym
}

func (*ProcTable) Close() {
}

func (p *ProcTable) createElfTable(m *elfRange) *ElfTable {
	if !strings.HasPrefix(m.mapRange.file, "/") {
		return nil
	}
	file := m.mapRange.file
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
		if m.mapRange.offset == executable.Off {
			base := m.mapRange.start - executable.Vaddr
			e.Rebase(base)
			return true
		}
	}
	return false
}

func binarySearchElfRange(e elfRange, pc uint64) int {
	if pc < e.mapRange.start {
		return 1
	}
	if pc >= e.mapRange.end {
		return -1
	}
	return 0
}
