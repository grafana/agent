package symtab

import (
	"bytes"
	"debug/elf"
	"encoding/binary"
	"errors"
	"fmt"
	"os"
	"sort"

	"github.com/edsrzf/mmap-go"
)

type MMapedElfFile struct {
	elf.FileHeader
	Sections []elf.SectionHeader
	Progs    []elf.ProgHeader

	symbols  []ElfSymbolIndex
	fpath    string
	mmaped   mmap.MMap
	openFile *os.File
}

type ElfSymbolIndex struct {
	SectionHeaderLink uint32
	NameIndex         uint32
	Value             uint64
}

func NewMMapedElfFile(fpath string) (*MMapedElfFile, error) {
	res := &MMapedElfFile{
		fpath: fpath,
	}
	err := res.ensureOpen()
	if err != nil {
		res.close()
		return nil, err
	}
	elfFile, err := elf.NewFile(bytes.NewReader(res.mmaped))
	if err != nil {
		res.close()
		return nil, err
	}
	progs := make([]elf.ProgHeader, 0, len(elfFile.Progs))
	sections := make([]elf.SectionHeader, 0, len(elfFile.Sections))
	for i := range elfFile.Progs {
		progs = append(progs, elfFile.Progs[i].ProgHeader)
	}
	for i := range elfFile.Sections {
		sections = append(sections, elfFile.Sections[i].SectionHeader)
	}
	res.FileHeader = elfFile.FileHeader
	res.Progs = progs
	res.Sections = sections

	////// todo remove it from here, make it general purpose
	//res.readSymbols()
	//if len(res.symbols) == 0 {
	//	res.close()
	//}
	return res, nil
}

func (e *MMapedElfFile) Resolve(addr uint64) string {
	if len(e.symbols) == 0 {
		return ""
	}
	if addr < e.symbols[0].Value {
		return ""
	}
	i := sort.Search(len(e.symbols), func(i int) bool {
		return addr < e.symbols[i].Value
	})
	i--
	sym := &e.symbols[i]
	name, _ := e.symbolName(sym)
	return name
}

func (e *MMapedElfFile) Cleanup() {
	e.close()
}

func (f *MMapedElfFile) Section(name string) *elf.SectionHeader {
	for i := range f.Sections {
		s := &f.Sections[i]
		if s.Name == name {
			return s
		}
	}
	return nil
}

func (f *MMapedElfFile) readSymbols() error {
	sym, err := f.getSymbols(elf.SHT_SYMTAB)
	if err != nil && err != ErrNoSymbols {
		return err
	}

	dynsym, err := f.getSymbols(elf.SHT_DYNSYM)
	if err != nil && err != ErrNoSymbols {
		return err
	}
	total := len(dynsym) + len(sym)
	if total == 0 {
		return ErrNoSymbols
	}
	all := make([]ElfSymbolIndex, 0, total)
	all = append(all, sym...)
	all = append(all, dynsym...)
	sort.Slice(all, func(i, j int) bool {
		return all[i].Value < all[j].Value
	})
	f.symbols = all
	return nil
}

func (f *MMapedElfFile) sectionByType(typ elf.SectionType) *elf.SectionHeader {
	for i := range f.Sections {
		s := &f.Sections[i]
		if s.Type == typ {
			return s
		}
	}
	return nil
}

func (f *MMapedElfFile) getSymbols(typ elf.SectionType) ([]ElfSymbolIndex, error) {
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

//func (f *MMapedElfFile) getSymbols32(typ elf.SectionType) ([]elf.Symbol, error) {
//	symtabSection := f.sectionByType(typ)
//	if symtabSection == nil {
//		return nil, ErrNoSymbols
//	}
//
//	data, err := symtabSection.Data()
//	if err != nil {
//		return nil, fmt.Errorf("cannot load symbol section: %w", err)
//	}
//	symtab := bytes.NewReader(data)
//	if symtab.Len()%elf.Sym32Size != 0 {
//		return nil, errors.New("length of symbol section is not a multiple of SymSize")
//	}
//
//	//strdata, err := f.stringTable(symtabSection.Link)
//	//if err != nil {
//	//	return nil, nil, fmt.Errorf("cannot load string table section: %w", err)
//	//}
//
//	// The first entry is all zeros.
//	var skip [elf.Sym32Size]byte
//	symtab.Read(skip[:])
//
//	symbols := make([]elf.Symbol, symtab.Len()/elf.Sym32Size)
//
//	i := 0
//	var sym elf.Sym32
//	for symtab.Len() > 0 {
//		binary.Read(symtab, f.ByteOrder, &sym)
//		//str, _ := getString(strdata, int(sym.Name))
//		//symbols[i].Name = str
//		symbols[i].Info = sym.Info
//		symbols[i].Other = sym.Other
//		symbols[i].Section = elf.SectionIndex(sym.Shndx)
//		symbols[i].Value = uint64(sym.Value)
//		symbols[i].Size = uint64(sym.Size)
//		i++
//	}
//
//	return symbols, nil
//}

func (f *MMapedElfFile) getSymbols64(typ elf.SectionType) ([]ElfSymbolIndex, error) {
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

	symbols := make([]ElfSymbolIndex, symtab.Len()/elf.Sym64Size)

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

func (f *MMapedElfFile) symbolName(i *ElfSymbolIndex) (string, error) {
	strSection, err := f.stringTable(i.SectionHeaderLink)
	if err != nil {
		return "", err
	}
	strdata, err := f.SectionData(strSection)
	if err != nil {
		return "", err
	}
	s, b := getString(strdata, int(i.NameIndex))
	if !b {
		return "", fmt.Errorf("elf getString")
	}
	return s, nil
}

func (f *MMapedElfFile) ensureOpen() error {
	if f.mmaped != nil {
		return nil
	}
	return f.open()
}

func (f *MMapedElfFile) close() {
	if f.mmaped != nil {
		f.mmaped.Unmap()
		f.openFile.Close()
		f.openFile = nil
	}
}
func (f *MMapedElfFile) open() error {
	//todo error flag to not retry
	fd, err := os.OpenFile(f.fpath, os.O_RDONLY, 0)
	if err != nil {
		return fmt.Errorf("open elf file %s %w", f.fpath, err)
	}
	mmaped, err := mmap.Map(fd, mmap.RDONLY, 0)
	if err != nil {
		return fmt.Errorf("mmap elf file %s %w", f.fpath, err)
	}
	f.openFile = fd
	f.mmaped = mmaped
	return nil
}

func (f *MMapedElfFile) SectionData(s *elf.SectionHeader) ([]byte, error) {
	if err := f.ensureOpen(); err != nil {
		return nil, err
	}
	from := s.Offset
	to := s.Offset + s.FileSize
	if from > uint64(len(f.mmaped)) || to > uint64(len(f.mmaped)) {
		return nil, fmt.Errorf("section oob %s %v", f.fpath, s)
	}

	return f.mmaped[from:to], nil
}

func (f *MMapedElfFile) stringTable(link uint32) (*elf.SectionHeader, error) {
	if link <= 0 || link >= uint32(len(f.Sections)) {
		return nil, errors.New("section has invalid string table link")
	}
	return &f.Sections[link], nil
}

// getString extracts a string from an ELF string table.
func getString(section []byte, start int) (string, bool) {
	if start < 0 || start >= len(section) {
		return "", false
	}

	for end := start; end < len(section); end++ {
		if section[end] == 0 {
			return string(section[start:end]), true
		}
	}
	return "", false
}
