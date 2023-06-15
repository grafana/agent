package elf

import (
	"bytes"
	"debug/elf"
	"encoding/binary"
	"errors"
	"fmt"
	"sort"

	"github.com/grafana/agent/component/pyroscope/ebpf/ebpfspy/symtab/gosym"
)

// symbols from .symtab, .dynsym

type SymbolIndex struct {
	Name  Name
	Value uint64
}

type SectionLinkIndex uint8

var sectionTypeSym SectionLinkIndex = 0
var sectionTypeDynSym SectionLinkIndex = 1

type Name uint32

func NewName(NameIndex uint32, linkIndex SectionLinkIndex) Name {
	return Name((NameIndex & 0x7fffffff) | uint32(linkIndex)<<31)
}

func (n *Name) NameIndex() uint32 {
	return uint32(*n) & 0x7fffffff
}

func (n *Name) LinkIndex() SectionLinkIndex {
	return SectionLinkIndex(*n >> 31)
}

type FlatSymbolIndex struct {
	Links  []uint32
	Names  []Name
	Values gosym.PCIndex
}
type SymbolTable struct {
	Index FlatSymbolIndex
	File  *MMapedElfFile
}

func (e *SymbolTable) Refresh() {

}

func (e *SymbolTable) DebugString() string {
	return fmt.Sprintf("SymbolTable{ f = %s , sz = %d }", e.File.FilePath(), e.Index.Values.Length())
}

func (e *SymbolTable) Resolve(addr uint64) string {
	if len(e.Index.Names) == 0 {
		return ""
	}
	i := e.Index.Values.FindIndex(addr)
	if i == -1 {
		return ""
	}
	name, _ := e.symbolName(i)
	return name
}

func (e *SymbolTable) Cleanup() {
	e.File.Close()
}

func (f *MMapedElfFile) NewSymbolTable() (*SymbolTable, error) {
	sym, sectionSym, err := f.getSymbols(elf.SHT_SYMTAB)
	if err != nil && err != ErrNoSymbols {
		return nil, err
	}

	dynsym, sectionDynSym, err := f.getSymbols(elf.SHT_DYNSYM)
	if err != nil && err != ErrNoSymbols {
		return nil, err
	}
	total := len(dynsym) + len(sym)
	if total == 0 {
		return nil, ErrNoSymbols
	}
	all := make([]SymbolIndex, 0, total) // todo avoid allocation
	all = append(all, sym...)
	all = append(all, dynsym...)

	sort.Slice(all, func(i, j int) bool {
		if all[i].Value == all[j].Value {
			return all[i].Name < all[j].Name
		}
		return all[i].Value < all[j].Value
	})

	res := &SymbolTable{Index: FlatSymbolIndex{
		Links: []uint32{
			sectionSym,    // should be at 0 - SectionTypeSym
			sectionDynSym, // should be at 1 - SectionTypeDynSym
		},
		Names:  make([]Name, total),
		Values: gosym.NewPCIndex(total),
	}, File: f}
	for i := range all {
		res.Index.Names[i] = all[i].Name
		res.Index.Values.Set(i, all[i].Value)
	}
	return res, nil
}

func (f *MMapedElfFile) getSymbols(typ elf.SectionType) ([]SymbolIndex, uint32, error) {
	switch f.Class {
	case elf.ELFCLASS64:
		return f.getSymbols64(typ)

	case elf.ELFCLASS32:
		return f.getSymbols32(typ)
	}

	return nil, 0, errors.New("not implemented")
}

// ErrNoSymbols is returned by File.Symbols and File.DynamicSymbols
// if there is no such section in the File.
var ErrNoSymbols = errors.New("no symbol section")

func (f *MMapedElfFile) getSymbols64(typ elf.SectionType) ([]SymbolIndex, uint32, error) {
	symtabSection := f.sectionByType(typ)
	if symtabSection == nil {
		return nil, 0, ErrNoSymbols
	}
	var linkIndex SectionLinkIndex
	if typ == elf.SHT_DYNSYM {
		linkIndex = sectionTypeDynSym
	} else {
		linkIndex = sectionTypeSym
	}

	data, err := f.SectionData(symtabSection)
	if err != nil {
		return nil, 0, fmt.Errorf("cannot load symbol section: %w", err)
	}
	symtab := bytes.NewReader(data)
	if symtab.Len()%elf.Sym64Size != 0 {
		return nil, 0, errors.New("length of symbol section is not a multiple of Sym64Size")
	}

	// The first entry is all zeros.
	var skip [elf.Sym64Size]byte
	symtab.Read(skip[:])

	symbols := make([]SymbolIndex, symtab.Len()/elf.Sym64Size)

	var sym elf.Sym64
	i := 0
	for symtab.Len() > 0 {
		binary.Read(symtab, f.ByteOrder, &sym)
		if sym.Value != 0 && sym.Info&0xf == byte(elf.STT_FUNC) {
			symbols[i].Value = sym.Value
			if sym.Name >= 0x7fffffff {
				return nil, 0, fmt.Errorf("wrong sym name")
			}
			symbols[i].Name = NewName(sym.Name, linkIndex)
			i++
		}
	}

	return symbols[:i], symtabSection.Link, nil
}

func (f *MMapedElfFile) getSymbols32(typ elf.SectionType) ([]SymbolIndex, uint32, error) {
	symtabSection := f.sectionByType(typ)
	if symtabSection == nil {
		return nil, 0, ErrNoSymbols
	}
	var linkIndex SectionLinkIndex
	if typ == elf.SHT_DYNSYM {
		linkIndex = sectionTypeDynSym
	} else {
		linkIndex = sectionTypeSym
	}

	data, err := f.SectionData(symtabSection)
	if err != nil {
		return nil, 0, fmt.Errorf("cannot load symbol section: %w", err)
	}
	symtab := bytes.NewReader(data)
	if symtab.Len()%elf.Sym32Size != 0 {
		return nil, 0, errors.New("length of symbol section is not a multiple of Sym64Size")
	}

	// The first entry is all zeros.
	var skip [elf.Sym32Size]byte
	symtab.Read(skip[:])

	symbols := make([]SymbolIndex, symtab.Len()/elf.Sym32Size)

	var sym elf.Sym32
	i := 0
	for symtab.Len() > 0 {
		binary.Read(symtab, f.ByteOrder, &sym)
		if sym.Value != 0 && sym.Info&0xf == byte(elf.STT_FUNC) {
			symbols[i].Value = uint64(sym.Value)
			if sym.Name >= 0x7fffffff {
				return nil, 0, fmt.Errorf("wrong sym name")
			}
			symbols[i].Name = NewName(sym.Name, linkIndex)
			i++
		}
	}

	return symbols[:i], symtabSection.Link, nil
}

func (f *SymbolTable) symbolName(idx int) (string, error) {
	linkIndex := f.Index.Names[idx].LinkIndex()
	SectionHeaderLink := f.Index.Links[linkIndex]
	strSection, err := f.File.stringTable(SectionHeaderLink)
	if err != nil {
		return "", err
	}
	NameIndex := f.Index.Names[idx].NameIndex()
	s, b := f.File.getString(int(NameIndex) + int(strSection.Offset))
	if !b {
		return "", fmt.Errorf("elf getString")
	}
	return s, nil
}
