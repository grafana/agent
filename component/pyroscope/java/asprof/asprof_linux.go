//go:build linux

package asprof

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/prometheus/procfs"
)

func DistributionForProcess(pid int) (*Distribution, error) {
	proc, err := procfs.NewProc(pid)
	if err != nil {
		return nil, fmt.Errorf("failed to select dist for pid %d %w", pid, err)
	}
	maps, err := proc.ProcMaps()
	if err != nil {
		return nil, fmt.Errorf("failed to select dist for pid %d %w", pid, err)
	}
	musl := false
	glibc := false
	for _, m := range maps {
		if strings.Contains(m.Pathname, "/lib/ld-musl-x86_64.so.1") ||
			strings.Contains(m.Pathname, "/lib/ld-musl-aarch64.so.1") {
			musl = true
		}
		if strings.HasSuffix(m.Pathname, "/libc.so.6") {
			glibc = true
		}
	}
	if musl && glibc {
		return nil, fmt.Errorf("failed to select dist for pid %d: both musl and glibc found", pid)
	}
	if musl {
		return muslDist, nil
	}
	if glibc {
		return glibcDist, nil
	}
	return nil, fmt.Errorf("failed to select dist for pid %d: neither musl nor glibc found", pid)
}

func (d *Distribution) LibPath() string {
	return filepath.Join(d.extractedDir, "lib/libasyncProfiler.so")
}

func (p *Profiler) ExtractDistributions() error {
	p.unpackOnce.Do(func() {
		glibcLib, glibcLauncher, err := getLibAndLauncher(glibcDist.targz)
		if err != nil {
			p.unpackError = err
			return
		}
		muslLib, muslLauncher, err := getLibAndLauncher(muslDist.targz)
		if err != nil {
			p.unpackError = err
			return
		}
		currentProcessDist, err := DistributionForProcess(os.Getpid())
		if err != nil {
			p.unpackError = err
			return
		}
		if currentProcessDist == muslDist {
			glibcLauncher = muslLauncher
		} else {
			muslLauncher = glibcLauncher
		}
		err = glibcDist.write(p.tmpDir, glibcLib, glibcLauncher)
		if err != nil {
			p.unpackError = err
			return
		}
		err = muslDist.write(p.tmpDir, muslLib, muslLauncher)
		if err != nil {
			p.unpackError = err
			return
		}
	})
	return p.unpackError
}

func ProcessPath(path string, pid int) string {
	f := ProcFile{path, pid}
	return f.ProcRootPath()
}

type ProcFile struct {
	Path string
	PID  int
}

func (f *ProcFile) ProcRootPath() string {
	return filepath.Join("/proc", strconv.Itoa(f.PID), "root", f.Path)
}