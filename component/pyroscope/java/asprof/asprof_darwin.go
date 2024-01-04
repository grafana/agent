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

func AllDistributions() []*Distribution {
	return []*Distribution{macDist}
}

func DistributionForProcess(pid int) *Distribution {
	return macDist
}

func (d *Distribution) LibPath() string {
	return filepath.Join(d.extractedDir, "lib/libasyncProfiler.dylib")
}

func ProcessPath(path string, pid int) string {
	return path
}
