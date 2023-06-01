package symtab

import (
	"bytes"
	"debug/elf"
	"encoding/hex"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"golang.org/x/exp/slices"
)

type ElfTable struct {
	table              *SymTab
	base               uint64
	typ                elf.Type
	executables        []elf.ProgHeader
	buildID            string
	loaded             bool
	fs                 string
	symbolFile         string
	symbolFileFileInfo stat
	elfCache           *ElfCache
	logger             log.Logger
}

type ElfTableOptions struct {
	UseDebugFiles bool
	ElfCache      *ElfCache
}

func NewElfTable(logger log.Logger, fs string, elfFilePath string, options ElfTableOptions) (*ElfTable, error) {
	fsElfFilePath := path.Join(fs, elfFilePath)
	elfFile, err := elf.Open(fsElfFilePath)
	if err != nil {
		return nil, fmt.Errorf("open elf file %s: %w", fsElfFilePath, err)
	}
	defer elfFile.Close()
	res := &ElfTable{
		logger:   logger,
		typ:      elfFile.Type,
		elfCache: options.ElfCache,
	}
	for _, prog := range elfFile.Progs {
		if prog.Type == elf.PT_LOAD && (prog.ProgHeader.Flags&elf.PF_X != 0) {
			res.executables = append(res.executables, prog.ProgHeader)
		}
	}
	res.buildID, _ = getBuildID(elfFile)

	symbolsFile := elfFilePath

	if options.UseDebugFiles {
		debugFile, fileInfo := findDebugFile(fs, elfFilePath, res.buildID, elfFile)
		if debugFile != "" {
			symbolsFile = debugFile
			res.symbolFileFileInfo = fileInfo
		}
	}

	res.fs = fs
	res.symbolFile = symbolsFile
	return res, nil
}

func (t *ElfTable) Rebase(base uint64) {
	t.base = base
	if t.table != nil {
		t.table.Rebase(base)
	}
}

func (t *ElfTable) load() {
	if t.loaded {
		return
	}
	symbols := t.elfCache.GetSymbolsByBuildID(t.buildID)
	if len(symbols) == 0 {
		if t.symbolFileFileInfo.dev == 0 && t.symbolFileFileInfo.ino == 0 {
			fileInfo, err := os.Stat(path.Join(t.fs, t.symbolFile))
			if err == nil && fileInfo != nil {
				t.symbolFileFileInfo = statFromFileInfo(fileInfo)
			}
		}
		symbols = t.elfCache.GetSymbolsByStat(t.symbolFileFileInfo)
	}
	if len(symbols) == 0 {
		elfFile, err := elf.Open(path.Join(t.fs, t.symbolFile))
		if err != nil {
			t.table = NewSymTab(nil)
			t.loaded = true
			return
		}
		defer elfFile.Close()

		level.Debug(t.logger).Log(
			"msg", "get elf symbols",
			"symbolFile", t.symbolFile,
			"buildID", t.buildID,
			"fs", t.fs,
		)

		symbols = getElfSymbols(t.symbolFile, elfFile)

		t.elfCache.CacheByBuildID(t.buildID, symbols)
		t.elfCache.CacheByStat(t.symbolFileFileInfo, symbols)
	} else {
		level.Debug(t.logger).Log(
			"msg", "get cached elf symbols",
			"symbolFile", t.symbolFile,
			"buildID", t.buildID,
			"fs", t.fs,
		)
	}
	t.table = NewSymTab(symbols)
	t.table.Rebase(t.base)
	t.loaded = true
}

func (t *ElfTable) Resolve(pc uint64) *Symbol {
	t.load()
	return t.table.Resolve(pc)
}

func getElfSymbols(elfPath string, elfFile *elf.File) []Symbol {
	symtab := getELFSymbolsFromSymtab(elfPath, elfFile)
	if len(symtab) > 0 {
		return symtab
	}
	pclntab, err := getELFSymbolsFromPCLN(elfPath, elfFile)
	if err != nil {
		return symtab
	}
	return pclntab
}

func getELFSymbolsFromSymtab(elfPath string, elfFile *elf.File) []Symbol {
	symtab, _ := elfFile.Symbols()
	dynsym, _ := elfFile.DynamicSymbols()
	var symbols []Symbol
	add := func(t []elf.Symbol) {
		for _, sym := range t {
			if sym.Value != 0 && sym.Info&0xf == byte(elf.STT_FUNC) {
				symbols = append(symbols, Symbol{
					Name:   sym.Name,
					Start:  sym.Value,
					Module: elfPath,
				})
			}
		}
	}
	add(symtab)
	add(dynsym)
	slices.SortFunc(symbols, func(a, b Symbol) bool {
		if a.Start == b.Start {
			return strings.Compare(a.Name, b.Name) < 0
		}
		return a.Start < b.Start
	})
	return symbols
}

func getBuildID(elfFile *elf.File) (string, error) {
	buildIDSection := elfFile.Section(".note.gnu.build-id")
	if buildIDSection == nil {
		return "", fmt.Errorf(".note.gnu.build-id section not found")
	}
	data, err := buildIDSection.Data()
	if err != nil {
		return "", fmt.Errorf("reading .note.gnu.build-id %w", err)
	}
	if len(data) < 16 {
		return "", fmt.Errorf(".note.gnu.build-id is too small")
	}
	if !bytes.Equal([]byte("GNU"), data[12:15]) {
		return "", fmt.Errorf(".note.gnu.build-id is not a GNU build-id")
	}
	buildID := hex.EncodeToString(data[16:])
	return buildID, nil
}

func findDebugFileWithBuildID(fs string, buildID string) (string, stat) {
	if len(buildID) < 3 {
		return "", stat{}
	}

	debugFile := fmt.Sprintf("/usr/lib/debug/.build-id/%s/%s.debug", buildID[:2], buildID[2:])
	fsDebugFile := path.Join(fs, debugFile)
	fileInfo, err := os.Stat(fsDebugFile)
	if err == nil {
		return debugFile, statFromFileInfo(fileInfo)
	}

	return "", stat{}
}

func findDebugFile(fs string, elfFilePath string, buildID string, elfFile *elf.File) (string, stat) {
	// https://sourceware.org/gdb/onlinedocs/gdb/Separate-Debug-Files.html
	// So, for example, suppose you ask GDB to debug /usr/bin/ls, which has a debug link that specifies the file
	// ls.debug, and a build ID whose value in hex is abcdef1234. If the list of the global debug directories
	// includes /usr/lib/debug, then GDB will look for the following debug information files, in the indicated order:
	//
	//- /usr/lib/debug/.build-id/ab/cdef1234.debug
	//- /usr/bin/ls.debug
	//- /usr/bin/.debug/ls.debug
	//- /usr/lib/debug/usr/bin/ls.debug.
	debugFile, fileInfo := findDebugFileWithBuildID(fs, buildID)
	if debugFile != "" {
		return debugFile, fileInfo
	}
	debugFile, fileInfo, _ = findDebugFileWithDebugLink(fs, elfFilePath, elfFile)
	return debugFile, fileInfo
}

func findDebugFileWithDebugLink(fs string, elfFilePath string, elfFile *elf.File) (string, stat, error) {
	debugLinkSection := elfFile.Section(".gnu_debuglink")
	if debugLinkSection == nil {
		return "", stat{}, fmt.Errorf("")
	}
	data, err := debugLinkSection.Data()
	if err != nil {
		return "", stat{}, fmt.Errorf("reading .gnu_debuglink %w", err)
	}
	if len(data) < 6 {
		return "", stat{}, fmt.Errorf(".gnu_debuglink is too small")
	}
	crc := data[len(data)-4:]
	_ = crc
	debugLink := cString(data)

	// /usr/bin/ls.debug
	fsDebugFile := path.Join(path.Dir(elfFilePath), debugLink)
	fileInfo, err := os.Stat(path.Join(fs, fsDebugFile))
	if err == nil {
		return fsDebugFile, statFromFileInfo(fileInfo), nil
	}
	// /usr/bin/.debug/ls.debug
	fsDebugFile = path.Join(path.Dir(elfFilePath), ".debug", debugLink)
	fileInfo, err = os.Stat(path.Join(fs, fsDebugFile))
	if err == nil {
		return fsDebugFile, statFromFileInfo(fileInfo), nil
	}
	// /usr/lib/debug/usr/bin/ls.debug.
	fsDebugFile = path.Join("/usr/lib/debug", path.Dir(elfFilePath), debugLink)
	fileInfo, err = os.Stat(path.Join(fs, fsDebugFile))
	if err == nil {
		return fsDebugFile, statFromFileInfo(fileInfo), nil
	}

	return "", stat{}, nil
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
