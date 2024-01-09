//go:build linux

package asprof

import (
	"crypto/sha1"
	_ "embed"
	"encoding/hex"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/prometheus/procfs"
)

func (p *Profiler) DistributionForProcess(pid int) (*Distribution, error) {
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
		if strings.Contains(m.Pathname, "x86_64-linux-gnu/libc-") {
			glibc = true
		}
	}
	if musl && glibc {
		return nil, fmt.Errorf("failed to select dist for pid %d: both musl and glibc found", pid)
	}
	if musl {
		return p.muslDist, nil
	}
	if glibc {
		return p.glibcDist, nil
	}
	if _, err := os.Stat(ProcessPath("/lib/ld-musl-x86_64.so.1", pid)); err == nil {
		return p.muslDist, nil
	}
	if _, err := os.Stat(ProcessPath("/lib/ld-musl-aarch64.so.1", pid)); err == nil {
		return p.muslDist, nil
	}
	if _, err := os.Stat(ProcessPath("/lib64/ld-linux-x86-64.so.2", pid)); err == nil {
		return p.glibcDist, nil
	}
	return nil, fmt.Errorf("failed to select dist for pid %d: neither musl nor glibc found", pid)
}

func (d *Distribution) LibPath() string {
	if binaryLauncher() {
		return filepath.Join(d.extractedDir, "lib/libasyncProfiler.so")
	}
	return filepath.Join(d.extractedDir, "build/libasyncProfiler.so")
}

func (d *Distribution) JattachPath() string {
	if binaryLauncher() {
		return ""
	}
	return filepath.Join(d.extractedDir, "build/jattach")
}

func (d *Distribution) Launcher() string {
	if binaryLauncher() {
		return filepath.Join(d.extractedDir, "bin/asprof")
	}
	return filepath.Join(d.extractedDir, "profiler.sh")
}

func (p *Profiler) ExtractDistributions() error {
	p.unpackOnce.Do(func() {
		p.extractDistributions()
	})
	return p.unpackError
}

func (p *Profiler) extractDistributions() {
	fsMutex.Lock()
	defer fsMutex.Unlock()
	sum := sha1.Sum(tarGzArchive)
	hexSum := hex.EncodeToString(sum[:])
	muslDistName := fmt.Sprintf("%s-%s-%s", tmpDirMarker,
		"musl",
		hexSum)
	glibcDistName := fmt.Sprintf("%s-%s-%s", tmpDirMarker,
		"glibc",
		hexSum)

	var launcher, jattach, glibc, musl []byte
	err := readTarGZ(tarGzArchive, func(name string, fi fs.FileInfo, data []byte) error {
		fmt.Printf("gz file %s\n", name)
		if name == "profiler.sh" || name == "asprof" {
			launcher = data
		}
		if name == "jattach" {
			jattach = data
		}
		if strings.Contains(name, "glibc/libasyncProfiler.so") {
			glibc = data
		}
		if strings.Contains(name, "musl/libasyncProfiler.so") {
			musl = data
		}
		return nil
	})
	if err != nil {
		p.unpackError = err
		return
	}
	if launcher == nil || glibc == nil || musl == nil {
		p.unpackError = fmt.Errorf("failed to find libasyncProfiler in tar.gz")
		return
	}
	if !binaryLauncher() {
		if jattach == nil {
			p.unpackError = fmt.Errorf("failed to find jattach in tar.gz")
			return
		}
	}
	fileMap := map[string][]byte{}
	fileMap[filepath.Join(glibcDistName, p.glibcDist.Launcher())] = launcher
	fileMap[filepath.Join(glibcDistName, p.glibcDist.LibPath())] = glibc
	fileMap[filepath.Join(muslDistName, p.muslDist.Launcher())] = launcher
	fileMap[filepath.Join(muslDistName, p.muslDist.LibPath())] = musl
	if !binaryLauncher() {
		fileMap[filepath.Join(glibcDistName, p.glibcDist.JattachPath())] = jattach
		fileMap[filepath.Join(muslDistName, p.muslDist.JattachPath())] = jattach
	}
	for path, data := range fileMap {
		err := writeFile(p.tmpDir, path, data)
		if err != nil {
			p.unpackError = err
			return
		}
	}
	p.glibcDist.extractedDir = filepath.Join(p.tmpDir, glibcDistName)
	p.muslDist.extractedDir = filepath.Join(p.tmpDir, muslDistName)
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
