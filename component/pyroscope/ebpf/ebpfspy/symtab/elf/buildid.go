package elf

import (
	"bytes"
	"encoding/hex"
	"fmt"
)

type BuildID struct {
	id  string
	typ string
}

func (b *BuildID) ID() string {
	return b.id
}
func GNUBuildID(s string) BuildID {
	return BuildID{id: s, typ: "gnu"}
}
func GoBuildID(s string) BuildID {
	return BuildID{id: s, typ: "go"}
}

func (b *BuildID) Empty() bool {
	return b.id == "" || b.typ == ""
}

func (b *BuildID) GNU() bool {
	return b.typ == "gnu"
}

var (
	ErrNoBuildIDSection = fmt.Errorf("build ID section not found")
)

func (elfFile *MMapedElfFile) BuildID() (BuildID, error) {
	id, err := elfFile.GNUBuildID()
	if err != nil && err != ErrNoBuildIDSection {
		return BuildID{}, err
	}
	if !id.Empty() {
		return id, nil
	}
	id, err = elfFile.GoBuildID()
	if err != nil && err != ErrNoBuildIDSection {
		return BuildID{}, err
	}
	if !id.Empty() {
		return id, nil
	}

	return BuildID{}, ErrNoBuildIDSection
}

var goBuildIDSep = []byte("/")

func (elfFile *MMapedElfFile) GoBuildID() (BuildID, error) {
	buildIDSection := elfFile.Section(".note.go.buildid")
	if buildIDSection == nil {
		return BuildID{}, ErrNoBuildIDSection
	}
	data, err := elfFile.SectionData(buildIDSection)
	if err != nil {
		return BuildID{}, fmt.Errorf("reading .note.go.buildid %w", err)
	}

	if len(data) < 40 || bytes.Count(data, goBuildIDSep) < 2 {
		return BuildID{}, fmt.Errorf("wrong .note.go.buildid %s", elfFile.fpath)
	}
	id := string(data)
	return GoBuildID(id), nil
}

func (elfFile *MMapedElfFile) GNUBuildID() (BuildID, error) {
	buildIDSection := elfFile.Section(".note.gnu.build-id")
	if buildIDSection == nil {
		return BuildID{}, ErrNoBuildIDSection
	}

	data, err := elfFile.SectionData(buildIDSection)
	if err != nil {
		return BuildID{}, fmt.Errorf("reading .note.gnu.build-id %w", err)
	}
	if len(data) < 16 {
		return BuildID{}, fmt.Errorf(".note.gnu.build-id is too small")
	}
	if !bytes.Equal([]byte("GNU"), data[12:15]) {
		return BuildID{}, fmt.Errorf(".note.gnu.build-id is not a GNU build-id")
	}
	rawBuildID := data[16:]
	if len(rawBuildID) < 20 {
		return BuildID{}, fmt.Errorf(".note.gnu.build-id is too small %s", elfFile.fpath)
	}
	buildIDHex := hex.EncodeToString(rawBuildID)
	return GNUBuildID(buildIDHex), nil
}
