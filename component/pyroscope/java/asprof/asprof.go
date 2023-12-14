package asprof

import (
	"bytes"
	"crypto/sha1"
	_ "embed"
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"sync"

	"k8s.io/utils/path"
)

// option1: embed the tar.gz file todo make this a module outside of agent repo
// option2: distribute the tar.gz file with the agent docker image

type Profiler struct {
	tmpDir      string
	unpackOnce  sync.Once
	unpackError error

	glibcDir string
	muslDir  string

	mutex sync.Mutex
}

func NewProfiler(tmpDir string) *Profiler {
	return &Profiler{tmpDir: tmpDir}
}

type Distribution bool

var Glibc Distribution = true
var Musl Distribution = false

func (p *Profiler) Extract() error {
	p.unpackOnce.Do(func() {
		sum := sha1.Sum(glibcDistribution)
		p.glibcDir = filepath.Join(p.tmpDir, "asprof-glibc-"+hex.EncodeToString(sum[:]))
		if p.unpackError = extractTarGZ(glibcDistribution, p.glibcDir); p.unpackError != nil {
			return
		}
		if err := os.Chmod(filepath.Join(p.glibcDir, glibcDistributionName, "bin", "asprof"), 0700); err != nil {
			p.unpackError = err
			return
		}
		//sum = sha1.Sum(muslDistribution)
		//p.muslDir = filepath.Join(p.tmpDir, "asprof-musl-"+hex.EncodeToString(sum[:]))
		//if p.unpackError = extractTarGZ(muslDistribution, p.muslDir); p.unpackError != nil {
		//	return
		//}
	})
	return p.unpackError
}

func (p *Profiler) LibPath(dist Distribution) string {
	if dist == Glibc {
		return filepath.Join(p.glibcDir, glibcDistributionName, "lib/libasyncProfiler.so")
	}
	return "TODO"
}

func (p *Profiler) TargetLibPath(dist Distribution, pid int) string {
	if dist == Glibc {
		f := File{Path: p.LibPath(dist), PID: pid}
		return f.ProcRootPath()
	}
	return "TODO"
}

func (p *Profiler) AsprofPath() string {
	return filepath.Join(p.glibcDir, glibcDistributionName, "bin", "asprof")
}

func (p *Profiler) Execute(dist Distribution, argv []string) (string, string, error) {
	stdout := bytes.NewBuffer(nil)
	stderr := bytes.NewBuffer(nil)

	exe := p.AsprofPath()
	cmd := exec.Command(exe, argv...)
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	err := cmd.Start()
	if err != nil {
		return "", "", fmt.Errorf("asprof failed to start %s: %w", exe, err)
	}
	err = cmd.Wait()
	if err != nil {
		return "", "", fmt.Errorf("asprof failed to run %s: %w", exe, err)
	}
	return stdout.String(), stderr.String(), nil
}
func (p *Profiler) CopyLib(dist Distribution, pid int) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	src := p.LibPath(dist)

	libBytes, err := os.ReadFile(src)
	if err != nil {
		return err
	}

	dst := p.TargetLibPath(dist, pid)
	targetExists, err := path.Exists(path.CheckSymlinkOnly, dst)
	if err != nil {
		return err
	}
	if targetExists {
		targetLibBytes, err := os.ReadFile(dst)
		if err != nil {
			return err
		}
		if !bytes.Equal(libBytes, targetLibBytes) {
			return fmt.Errorf("file %s already exists and is different", dst)
		}
		return nil
	} else {
		fd, err := os.OpenFile(dst, os.O_CREATE|os.O_EXCL, 0644)
		if err != nil {
			return fmt.Errorf("failed to create file %s: %w", dst, err)
		}
		defer fd.Close()
		//todo check fd was not manipulated with symlinks
		n, err := fd.Write(libBytes)
		if err != nil {
			return fmt.Errorf("failed to write to file %s: %w", dst, err)
		}
		if n != len(libBytes) {
			return fmt.Errorf("failed to write to file %s %d", dst, n)
		}
		return nil
	}
}

// todo use timeout argument

type File struct {
	Path string
	PID  int
}

func (f *File) ProcRootPath() string {
	return filepath.Join("/proc", strconv.Itoa(f.PID), "root", f.Path)
}

func (f *File) Read() ([]byte, error) {
	return os.ReadFile(f.ProcRootPath())
}

func (f *File) Delete() error {
	return os.Remove(f.ProcRootPath())
}
