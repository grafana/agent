package asprof

import (
	"bytes"
	_ "embed"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"

	"k8s.io/utils/path"
)

// option1: embed the tar.gz file todo make this a module outside of agent repo
// option2: distribute the tar.gz file with the agent docker image

type Profiler struct {
	tmpDir     string
	unpackOnce sync.Once

	mutex       sync.Mutex
	unpackError error
}

func NewProfiler(tmpDir string) *Profiler {
	return &Profiler{tmpDir: tmpDir}
}

type Distribution struct {
	targz        []byte
	fname        string
	version      int
	extractedDir string
}

func (d *Distribution) AsprofPath() string {
	if d.version < 300 {
		return filepath.Join(d.extractedDir, "bin/profiler.sh")
	}
	return filepath.Join(d.extractedDir, "bin/asprof")
}

func (p *Profiler) Execute(dist *Distribution, argv []string) (string, string, error) {
	stdout := bytes.NewBuffer(nil)
	stderr := bytes.NewBuffer(nil)

	exe := dist.AsprofPath()
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
	p.mutex.Lock()
	defer p.mutex.Unlock()
	src := dist.LibPath()

	libBytes, err := os.ReadFile(src)
	if err != nil {
		return err
	}

	dst := ProcessPath(dist.LibPath(), pid)
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
		libDir := filepath.Dir(dst)
		err = os.MkdirAll(libDir, 0755)
		if err != nil {
			return fmt.Errorf("failed to create directory %s: %w", libDir, err)
		}
		binDir := filepath.Join(filepath.Dir(libDir), "bin")
		err = os.MkdirAll(binDir, 0755) // asprof tries to load as bin/../lib/libasyncProfiler.so
		if err != nil {
			return fmt.Errorf("failed to create directory %s: %w", binDir, err)
		}
		fd, err := os.OpenFile(dst, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0644)
		if err != nil {
			return fmt.Errorf("failed to create file %s: %w", dst, err)
		}
		defer fd.Close()
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