package symtab

import (
	"fmt"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/go-kit/log"
	"golang.org/x/exp/slices"
)

type ProcTable struct {
	logger     log.Logger
	ranges     []elfRange
	file2Table map[file]*ElfTable
	options    ProcTableOptions
	rootFS     string
}

func (p *ProcTable) DebugString() string {
	return "ProcTable{TODO}"
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
		e = p.createElfTable(r.mapRange)
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
	moduleOffset := pc - t.base
	if s == "" {
		return Symbol{Start: moduleOffset, Module: r.mapRange.Pathname}
	}

	return Symbol{Start: moduleOffset, Name: s, Module: r.mapRange.Pathname}
}

func (p *ProcTable) createElfTable(m *ProcMap) *ElfTable {
	if !strings.HasPrefix(m.Pathname, "/") {
		return nil
	}
	e := NewElfTable(p.logger, m, p.rootFS, m.Pathname, p.options.ElfTableOptions)
	return e
}

func (p *ProcTable) Cleanup() {
	for _, table := range p.file2Table {
		table.Cleanup()
	}
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
