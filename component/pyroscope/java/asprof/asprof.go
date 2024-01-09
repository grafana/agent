package asprof

import (
	"bytes"
	_ "embed"
	"fmt"
	"os"
	"os/exec"
	"sync"
)

var fsMutex sync.Mutex

type Profiler struct {
	tmpDir     string
	unpackOnce sync.Once

	glibcDist   *Distribution
	muslDist    *Distribution
	unpackError error
}

func NewProfiler(tmpDir string) *Profiler {
	return &Profiler{tmpDir: tmpDir, glibcDist: new(Distribution), muslDist: new(Distribution)}
}

type Distribution struct {
	extractedDir string
}

func binaryLauncher() bool {
	return version >= 210
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
	data, err := os.ReadFile(dist.LibPath())
	if err != nil {
		return err
	}
	path := ProcessPath(dist.LibPath(), pid)
	return writeFile("/", path, data)
}
