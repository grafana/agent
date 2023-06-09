package elf

import "sort"

type ElfSymbolIndex struct {
	SectionHeaderLink uint32
	NameIndex         uint32
	Value             uint64
}

type SymbolTable struct {
	Symbols []ElfSymbolIndex
	File    *MMapedElfFile
}

func (e *SymbolTable) Resolve(addr uint64) string {
	if len(e.Symbols) == 0 {
		return ""
	}
	if addr < e.Symbols[0].Value {
		return ""
	}
	i := sort.Search(len(e.Symbols), func(i int) bool {
		return addr < e.Symbols[i].Value
	})
	i--
	sym := &e.Symbols[i]
	name, _ := e.File.symbolName(sym)
	return name
}

func (e *SymbolTable) Cleanup() {
	e.File.Close()
}
