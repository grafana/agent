package symtab

import (
	"debug/elf"
	"fmt"
	"os"
	"path"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	elf2 "github.com/grafana/agent/component/pyroscope/ebpf/ebpfspy/symtab/elf"
)

var (
	errElfBaseNotFound = fmt.Errorf("elf base not found")
)

type ElfTable struct {
	fs          string
	elfFilePath string
	table       SymbolNameResolver
	base        uint64

	loaded bool
	err    error

	options ElfTableOptions
	logger  log.Logger
	procMap *ProcMap
}

type ElfTableOptions struct {
	ElfCache *ElfCache
}

func NewElfTable(logger log.Logger, procMap *ProcMap, fs string, elfFilePath string, options ElfTableOptions) *ElfTable {
	res := &ElfTable{
		procMap:     procMap,
		fs:          fs,
		elfFilePath: elfFilePath,
		logger:      logger,
		options:     options,
		table:       &noopSymbolNameResolver{},
	}
	return res
}

func (et *ElfTable) findBase(e *elf2.MMapedElfFile) bool {
	m := et.procMap
	if e.FileHeader.Type == elf.ET_EXEC {
		et.base = 0
		return true
	}
	for _, prog := range e.Progs {
		if prog.Type == elf.PT_LOAD && (prog.Flags&elf.PF_X != 0) {
			if uint64(m.Offset) == prog.Off {
				et.base = m.StartAddr - prog.Vaddr
				return true
			}
		}
	}
	return false
}

func (et *ElfTable) load() {
	if et.loaded {
		return
	}
	et.loaded = true
	fsElfFilePath := path.Join(et.fs, et.elfFilePath)

	me, err := elf2.NewMMapedElfFile(fsElfFilePath)
	if err != nil {
		et.err = err
		return
	}
	defer me.Close() // todo do not close if it is the selected elf

	if !et.findBase(me) {
		et.err = errElfBaseNotFound
		return
	}
	buildID, err := me.BuildID()
	if err != nil && err != elf2.ErrNoBuildIDSection {
		et.err = err
		return
	}

	symbols := et.options.ElfCache.GetSymbolsByBuildID(buildID)
	if symbols != nil {
		et.table = symbols
		return
	}
	fileInfo, err := os.Stat(fsElfFilePath)
	if err != nil {
		et.err = err
		return
	}
	symbols = et.options.ElfCache.GetSymbolsByStat(statFromFileInfo(fileInfo))
	if symbols != nil {
		et.table = symbols
		return
	}

	debugFilePath := et.findDebugFile(buildID, me)
	if debugFilePath != "" {
		debugMe, err := elf2.NewMMapedElfFile(path.Join(et.fs, debugFilePath))
		if err != nil {
			et.err = err
			return
		}
		defer debugMe.Close() // todo do not close if it is the selected elf

		symbols, err = et.createSymbolTable(debugMe)
		if err != nil {
			et.err = err
			return
		}
		et.table = symbols
		et.options.ElfCache.CacheByBuildID(buildID, symbols)
		return
	}

	symbols, err = et.createSymbolTable(me)
	level.Debug(et.logger).Log("msg", "create symbol table", "f", me.FilePath())
	if err != nil {
		et.err = err
		return
	}

	et.table = symbols
	et.options.ElfCache.CacheByBuildID(buildID, symbols)
	et.options.ElfCache.CacheByStat(statFromFileInfo(fileInfo), symbols)
}

func (et *ElfTable) createSymbolTable(me *elf2.MMapedElfFile) (SymbolNameResolver, error) {
	symTable, symErr := me.NewSymbolTable()
	goTable, goErr := me.NewGoTable()
	if symErr != nil && goErr != nil {
		return nil, fmt.Errorf("s: %w g: %w", symErr, goErr)
	}
	if symErr == nil && goErr == nil {
		return &elf2.GoTableWithFallback{
			GoTable:  goTable,
			SymTable: symTable,
		}, nil
	}
	if symErr == nil {
		return symTable, nil
	}
	if goTable != nil {
		return goTable, nil
	}
	panic("unreachable")
}

func (et *ElfTable) Resolve(pc uint64) string {
	et.load()
	pc -= et.base
	return et.table.Resolve(pc)
}

func (et *ElfTable) Cleanup() {
	if et.table != nil {
		et.table.Cleanup()
	}
}

func (et *ElfTable) findDebugFileWithBuildID(buildID elf2.BuildID) string {
	id := buildID.ID
	if len(id) < 3 || !buildID.GNU() {
		return ""
	}

	debugFile := fmt.Sprintf("/usr/lib/debug/.build-id/%s/%s.debug", id[:2], id[2:])
	fsDebugFile := path.Join(et.fs, debugFile)
	_, err := os.Stat(fsDebugFile)
	if err == nil {
		return debugFile
	}

	return ""
}

func (et *ElfTable) findDebugFile(buildID elf2.BuildID, elfFile *elf2.MMapedElfFile) string {
	// https://sourceware.org/gdb/onlinedocs/gdb/Separate-Debug-Files.html
	// So, for example, suppose you ask GDB to debug /usr/bin/ls, which has a debug link that specifies the file
	// ls.debug, and a build ID whose value in hex is abcdef1234. If the list of the global debug directories
	// includes /usr/lib/debug, then GDB will look for the following debug information files, in the indicated order:
	//
	//- /usr/lib/debug/.build-id/ab/cdef1234.debug
	//- /usr/bin/ls.debug
	//- /usr/bin/.debug/ls.debug
	//- /usr/lib/debug/usr/bin/ls.debug.
	debugFile := et.findDebugFileWithBuildID(buildID)
	if debugFile != "" {
		return debugFile
	}
	debugFile = et.findDebugFileWithDebugLink(elfFile)
	return debugFile
}

func (et *ElfTable) findDebugFileWithDebugLink(elfFile *elf2.MMapedElfFile) string {
	fs := et.fs
	elfFilePath := et.elfFilePath
	debugLinkSection := elfFile.Section(".gnu_debuglink")
	if debugLinkSection == nil {
		return ""
	}
	data, err := elfFile.SectionData(debugLinkSection)
	if err != nil {
		return ""
	}
	if len(data) < 6 {
		return ""
	}
	crc := data[len(data)-4:]
	_ = crc
	debugLink := cString(data)

	// /usr/bin/ls.debug
	fsDebugFile := path.Join(path.Dir(elfFilePath), debugLink)
	_, err = os.Stat(path.Join(fs, fsDebugFile))
	if err == nil {
		return fsDebugFile
	}
	// /usr/bin/.debug/ls.debug
	fsDebugFile = path.Join(path.Dir(elfFilePath), ".debug", debugLink)
	_, err = os.Stat(path.Join(fs, fsDebugFile))
	if err == nil {
		return fsDebugFile
	}
	// /usr/lib/debug/usr/bin/ls.debug.
	fsDebugFile = path.Join("/usr/lib/debug", path.Dir(elfFilePath), debugLink)
	_, err = os.Stat(path.Join(fs, fsDebugFile))
	if err == nil {
		return fsDebugFile
	}

	return ""
}

func cString(bs []byte) string {
	i := 0
	for ; i < len(bs); i++ {
		if bs[i] == 0 {
			break
		}
	}
	return string(bs[:i])
}

type ElfDebugInfo struct {
	SymbolsCount int    `river:"symbols_count,attr,optional"`
	File         string `river:"file,attr,optional"`
}

func (et *ElfTable) DebugInfo() elf2.SymTabDebugInfo {
	return et.table.DebugInfo()
}
