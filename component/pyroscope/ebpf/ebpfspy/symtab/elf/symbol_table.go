package elf

import (
	"bytes"
	"debug/elf"
	"encoding/binary"
	"errors"
	"fmt"
	"sort"
)

// symbols from .symtab, .dynsym

type SymbolIndex struct {
	SectionHeaderLink uint32
	NameIndex         uint32
	Value             uint64
}

type SymbolTable struct {
	// todo make it 3 separate tables
	Symbols []SymbolIndex
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
	name, _ := e.symbolName(sym)
	return name
}

func (e *SymbolTable) Cleanup() {
	e.File.Close()
}

func (f *MMapedElfFile) ReadSymbols() (*SymbolTable, error) {
	sym, err := f.getSymbols(elf.SHT_SYMTAB)
	if err != nil && err != ErrNoSymbols {
		return nil, err
	}

	dynsym, err := f.getSymbols(elf.SHT_DYNSYM)
	if err != nil && err != ErrNoSymbols {
		return nil, err
	}
	total := len(dynsym) + len(sym)
	if total == 0 {
		return nil, ErrNoSymbols
	}
	all := make([]SymbolIndex, 0, total)
	all = append(all, sym...)
	all = append(all, dynsym...)
	sort.Slice(all, func(i, j int) bool {
		return all[i].Value < all[j].Value
	})
	//f.Symbols = all
	return &SymbolTable{Symbols: all, File: f}, nil
}

func (f *MMapedElfFile) getSymbols(typ elf.SectionType) ([]SymbolIndex, error) {
	switch f.Class {
	case elf.ELFCLASS64:
		return f.getSymbols64(typ)

		//case elf.ELFCLASS32://todo
		//	return f.getSymbols32(typ)
	}

	return nil, errors.New("not implemented")
}

// ErrNoSymbols is returned by File.Symbols and File.DynamicSymbols
// if there is no such section in the File.
var ErrNoSymbols = errors.New("no symbol section")

func (f *MMapedElfFile) getSymbols64(typ elf.SectionType) ([]SymbolIndex, error) {
	symtabSection := f.sectionByType(typ)
	if symtabSection == nil {
		return nil, ErrNoSymbols
	}

	data, err := f.SectionData(symtabSection)
	if err != nil {
		return nil, fmt.Errorf("cannot load symbol section: %w", err)
	}
	symtab := bytes.NewReader(data)
	if symtab.Len()%elf.Sym64Size != 0 {
		return nil, errors.New("length of symbol section is not a multiple of Sym64Size")
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
			symbols[i].SectionHeaderLink = symtabSection.Link
			symbols[i].NameIndex = sym.Name
			i++
		}
	}

	return symbols[:i], nil
}

func (f *SymbolTable) symbolName(i *SymbolIndex) (string, error) {
	strSection, err := f.File.stringTable(i.SectionHeaderLink)
	if err != nil {
		return "", err
	}
	strdata, err := f.File.SectionData(strSection)
	if err != nil {
		return "", err
	}
	s, b := getString(strdata, int(i.NameIndex))
	if !b {
		return "", fmt.Errorf("elf getString")
	}
	return s, nil
}
