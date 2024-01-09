package asprof

import (
	"bytes"
	"crypto/sha1"
	_ "embed"
	"encoding/hex"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/prometheus/procfs"
)

var fsMutex sync.Mutex

// separte dirs for glibc & musl
type Distribution struct {
	extractedDir string
	version      int
}

func (d *Distribution) binaryLauncher() bool {
	return d.version >= 210
}

func (d *Distribution) LibPath() string {
	if d.binaryLauncher() {
		return filepath.Join(d.extractedDir, "lib/libasyncProfiler.so")
	}
	return filepath.Join(d.extractedDir, "build/libasyncProfiler.so")
}

func (d *Distribution) JattachPath() string {
	if d.binaryLauncher() {
		return ""
	}
	return filepath.Join(d.extractedDir, "build/jattach")
}

func (d *Distribution) Launcher() string {
	if d.binaryLauncher() {
		return filepath.Join(d.extractedDir, "bin/asprof")
	}
	return filepath.Join(d.extractedDir, "profiler.sh")
}

type Profiler struct {
	tmpDir       string
	extractOnce  sync.Once
	glibcDist    *Distribution
	muslDist     *Distribution
	extractError error
	tmpDirMarker any
	archiveHash  string
	archive      Archive
}

type Archive struct {
	data    []byte
	version int
}

func NewProfiler(tmpDir string, archive Archive) *Profiler {
	res := &Profiler{tmpDir: tmpDir, glibcDist: new(Distribution), muslDist: new(Distribution), tmpDirMarker: "grafana-agent-asprof"}
	sum := sha1.Sum(archive.data)
	hexSum := hex.EncodeToString(sum[:])
	res.archiveHash = hexSum
	res.glibcDist.version = archive.version
	res.muslDist.version = archive.version
	res.archive = archive
	return res
}

func (p *Profiler) Execute(dist *Distribution, argv []string) (string, string, error) {
	stdout := bytes.NewBuffer(nil)
	stderr := bytes.NewBuffer(nil)

	exe := dist.Launcher()
	cmd := exec.Command(exe, argv...)

	cmd.Stdout = stdout
	cmd.Stderr = stderr
	err := cmd.Start()
	if err != nil {
		return stdout.String(), stderr.String(), fmt.Errorf("asprof failed to start %s: %w", exe, err)
	}
	err = cmd.Wait()
	if err != nil {
		return stdout.String(), stderr.String(), fmt.Errorf("asprof failed to run %s: %w", exe, err)
	}
	return stdout.String(), stderr.String(), nil
}

func (p *Profiler) CopyLib(dist *Distribution, pid int) error {
	fsMutex.Lock()
	defer fsMutex.Unlock()
	libData, err := os.ReadFile(dist.LibPath())
	if err != nil {
		return err
	}
	procRoot := ProcessPath("/", pid)
	procRootFile, err := os.Open(procRoot)
	if err != nil {
		return fmt.Errorf("failed to open proc root %s: %w", procRoot, err)
	}

	return writeFile(procRootFile, dist.LibPath(), libData)
}

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

func (p *Profiler) ExtractDistributions() error {
	p.extractOnce.Do(func() {
		p.extractError = p.extractDistributions()
	})
	return p.extractError
}

func (p *Profiler) extractDistributions() error {
	fsMutex.Lock()
	defer fsMutex.Unlock()
	muslDistName, glibcDistName := p.getDistNames()

	var launcher, jattach, glibc, musl []byte
	err := readTarGZ(p.archive.data, func(name string, fi fs.FileInfo, data []byte) error {
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
		return err
	}
	if launcher == nil || glibc == nil || musl == nil {
		return fmt.Errorf("failed to find libasyncProfiler in tar.gz")
	}
	if !p.glibcDist.binaryLauncher() {
		if jattach == nil {
			return fmt.Errorf("failed to find jattach in tar.gz")
		}
	}
	fileMap := map[string][]byte{}
	fileMap[filepath.Join(glibcDistName, p.glibcDist.Launcher())] = launcher
	fileMap[filepath.Join(glibcDistName, p.glibcDist.LibPath())] = glibc
	fileMap[filepath.Join(muslDistName, p.muslDist.Launcher())] = launcher
	fileMap[filepath.Join(muslDistName, p.muslDist.LibPath())] = musl
	if !p.glibcDist.binaryLauncher() {
		fileMap[filepath.Join(glibcDistName, p.glibcDist.JattachPath())] = jattach
		fileMap[filepath.Join(muslDistName, p.muslDist.JattachPath())] = jattach
	}
	tmpDirFile, err := os.Open(p.tmpDir)
	if err != nil {
		return fmt.Errorf("failed to open tmp dir %s: %w", p.tmpDir, err)
	}
	defer tmpDirFile.Close()

	if err = checkTempDirPermissions(tmpDirFile); err != nil {
		return err
	}

	for path, data := range fileMap {
		if err := writeFile(tmpDirFile, path, data); err != nil {
			return err
		}
	}
	p.glibcDist.extractedDir = filepath.Join(p.tmpDir, glibcDistName)
	p.muslDist.extractedDir = filepath.Join(p.tmpDir, muslDistName)
	return nil
}

func (p *Profiler) getDistNames() (string, string) {
	muslDistName := fmt.Sprintf("%s-%s-%s", p.tmpDirMarker,
		"musl",
		p.archiveHash)
	glibcDistName := fmt.Sprintf("%s-%s-%s", p.tmpDirMarker,
		"glibc",
		p.archiveHash)
	return muslDistName, glibcDistName
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
