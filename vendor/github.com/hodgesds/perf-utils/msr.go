package perf

import (
	"os"
	"path/filepath"
)

const (
	// MSRBaseDir is the base dir for MSRs.
	MSRBaseDir = "/dev/cpu"
)

// MSRPaths returns the set of MSR paths.
func MSRPaths() ([]string, error) {
	msrs := []string{}
	err := filepath.Walk(MSRBaseDir, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			return nil
		}
		if path != MSRBaseDir {
			// TODO: replace this with a real recursive walk.
			msrs = append(msrs, path+"/msr")
		}
		return nil
	})
	return msrs, err
}

// MSRs attemps to return all available MSRs.
func MSRs(onErr func(error)) []*MSR {
	paths, err := MSRPaths()
	if err != nil {
		onErr(err)
		return nil
	}
	msrs := []*MSR{}
	for _, path := range paths {
		msr, err := NewMSR(path)
		if err != nil {
			onErr(err)
			continue
		}
		msrs = append(msrs, msr)
	}
	return msrs
}

// MSR represents a Model Specific Register
type MSR struct {
	f *os.File
}

// NewMSR returns a MSR.
func NewMSR(path string) (*MSR, error) {
	f, err := os.OpenFile(path, os.O_RDWR, 0660)
	if err != nil {
		return nil, err
	}
	return &MSR{
		f: f,
	}, nil
}

// Read is used to read a MSR value.
func (m *MSR) Read(off int64) ([]byte, error) {
	b := make([]byte, 8)
	_, err := m.f.ReadAt(b, off)
	return b, err
}

// Close is used to close the MSR.
func (m *MSR) Close() error {
	if m.f != nil {
		return m.f.Close()
	}
	return nil
}
