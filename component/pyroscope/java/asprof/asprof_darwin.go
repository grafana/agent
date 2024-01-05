//go:build darwin

package asprof

import (
	_ "embed"
	"path/filepath"
)

//go:embed async-profiler-3.0-ea-macos-arm64.tar.gz
var distribution []byte

var macDist = &Distribution{
	targz:   distribution,
	fname:   "async-profiler-3.0-ea-macos-arm64.tar.gz",
	version: 300,
}

func (p *Profiler) Extract() error {
	p.unpackOnce.Do(func() {
		lib, launcher, err := getLibAndLauncher(distribution)
		if err != nil {
			p.unpackError = err
			return
		}
		err := d.Extract(p.tmpDir, lib, launcher)
		if err != nil {
			p.unpackError = err
			break
		}

	})
	return p.unpackError
}

func DistributionForProcess(pid int) (*Distribution, error) {
	return macDist, nil
}

func (d *Distribution) LibPath() string {
	return filepath.Join(d.extractedDir, "lib/libasyncProfiler.dylib")
}

func ProcessPath(path string, pid int) string {
	return path
}
