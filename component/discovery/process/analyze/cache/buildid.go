package cache

import (
	"bytes"
	"debug/elf"
	"encoding/hex"
	"errors"
	"fmt"
)

// copypaste from https://github.com/grafana/pyroscope/blob/8a7fe2b80c219bfda9be685ff27ca1dee4218a42/ebpf/symtab/elf/buildid.go#L31

//type BuildID struct {
//	ID  string
//	Typ string
//}
//
//func GNUBuildID(s string) BuildID {
//	return BuildID{ID: s, Typ: "gnu"}
//}
//func GoBuildID(s string) BuildID {
//	return BuildID{ID: s, Typ: "go"}
//}
//
//func (b *BuildID) Empty() bool {
//	return b.ID == "" || b.Typ == ""
//}
//
//func (b *BuildID) GNU() bool {
//	return b.Typ == "gnu"
//}

var (
	ErrNoBuildIDSection = fmt.Errorf("build ID section not found")
)

func BuildID(f *elf.File) (string, error) {
	id, err := GNUBuildID(f)
	if err != nil && !errors.Is(err, ErrNoBuildIDSection) {
		return "", err
	}
	if id != "" {
		return id, nil
	}
	id, err = GoBuildID(f)
	if err != nil && !errors.Is(err, ErrNoBuildIDSection) {
		return "", err
	}
	if id != "" {
		return id, nil
	}

	return "", ErrNoBuildIDSection
}

var goBuildIDSep = []byte("/")

func GoBuildID(f *elf.File) (string, error) {
	buildIDSection := f.Section(".note.go.buildid")
	if buildIDSection == nil {
		return "", ErrNoBuildIDSection
	}

	data, err := buildIDSection.Data()
	if err != nil {
		return "", fmt.Errorf("reading .note.go.buildid %w", err)
	}
	if len(data) < 17 {
		return "", fmt.Errorf(".note.gnu.build-id is too small")
	}

	data = data[16 : len(data)-1]
	if len(data) < 40 || bytes.Count(data, goBuildIDSep) < 2 {
		return "", fmt.Errorf("wrong .note.go.buildid ")
	}
	id := string(data)
	if id == "redacted" {
		return "", fmt.Errorf("blacklisted  .note.go.buildid ")
	}
	return id, nil
}

func GNUBuildID(f *elf.File) (string, error) {
	buildIDSection := f.Section(".note.gnu.build-id")
	if buildIDSection == nil {
		return "", ErrNoBuildIDSection
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
	rawBuildID := data[16:]
	if len(rawBuildID) != 20 && len(rawBuildID) != 8 { // 8 is xxhash, for example in Container-Optimized OS
		return "", fmt.Errorf(".note.gnu.build-id has wrong size ")
	}
	buildIDHex := hex.EncodeToString(rawBuildID)
	return buildIDHex, nil
}
