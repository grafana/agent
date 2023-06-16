// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

/*
Package elf implements access to ELF object files.

# Security

This package is not designed to be hardened against adversarial inputs, and is
outside the scope of https://go.dev/security/policy. In particular, only basic
validation is done when parsing object files. As such, care should be taken when
parsing untrusted inputs, as parsing malformed files may consume significant
resources, or cause panics.
*/

// Copied from here https://github.com/golang/go/blob/go1.20.5/src/debug/elf/file.go#L585
// modified to not read symbol names in memory and return []SymbolIndex

package elf

import (
	"bytes"
	"debug/elf"
	"encoding/binary"
	"errors"
	"fmt"
)

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
	_, _ = symtab.Read(skip[:])

	symbols := make([]SymbolIndex, symtab.Len()/elf.Sym64Size)

	var sym elf.Sym64
	i := 0
	for symtab.Len() > 0 {
		_ = binary.Read(symtab, f.ByteOrder, &sym)
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
	_, _ = symtab.Read(skip[:])

	symbols := make([]SymbolIndex, symtab.Len()/elf.Sym32Size)

	var sym elf.Sym32
	i := 0
	for symtab.Len() > 0 {
		_ = binary.Read(symtab, f.ByteOrder, &sym)
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
