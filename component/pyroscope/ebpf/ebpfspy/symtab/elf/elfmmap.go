package elf

import (
	"bytes"
	"debug/elf"
	"errors"
	"fmt"
	"os"
	"runtime"

	"github.com/edsrzf/mmap-go"
)

type MMapedElfFile struct {
	elf.FileHeader
	Sections []elf.SectionHeader
	Progs    []elf.ProgHeader

	fpath  string
	err    error
	mmaped mmap.MMap
	fd     *os.File
}

func NewMMapedElfFile(fpath string) (*MMapedElfFile, error) {
	res := &MMapedElfFile{
		fpath: fpath,
	}
	err := res.ensureOpen()
	if err != nil {
		res.Close()
		return nil, err
	}
	elfFile, err := elf.NewFile(bytes.NewReader(res.mmaped))
	if err != nil {
		res.Close()
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

	runtime.SetFinalizer(res, (*MMapedElfFile).Finalize)
	return res, nil
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

func (f *MMapedElfFile) sectionByType(typ elf.SectionType) *elf.SectionHeader {
	for i := range f.Sections {
		s := &f.Sections[i]
		if s.Type == typ {
			return s
		}
	}
	return nil
}

func (f *MMapedElfFile) ensureOpen() error {
	if f.mmaped != nil {
		return nil
	}
	return f.open()
}

func (f *MMapedElfFile) Finalize() {
	if f.mmaped != nil {
		println("ebpf mmaped elf not closed")
	}
	f.Close()
}
func (f *MMapedElfFile) Close() {
	if f.mmaped != nil {
		f.mmaped.Unmap()
		f.fd.Close()
		f.fd = nil
	}
}
func (f *MMapedElfFile) open() error {
	if f.err != nil {
		return fmt.Errorf("failed previously %w", f.err)
	}
	fd, err := os.OpenFile(f.fpath, os.O_RDONLY, 0)
	if err != nil {
		f.err = err
		return fmt.Errorf("open elf file %s %w", f.fpath, err)
	}
	mmaped, err := mmap.Map(fd, mmap.RDONLY, 0)
	if err != nil {
		fd.Close()
		f.err = err
		return fmt.Errorf("mmap elf file %s %w", f.fpath, err)
	}
	f.fd = fd
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

func (f *MMapedElfFile) FilePath() string {
	return f.fpath
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
