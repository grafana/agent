//go:build linux && arm64

package asprof

import _ "embed"

var glibcDistribution []byte
var glibcDistributionName = "TODO"

var muslDistribution []byte
var muslDistributionName = "TODO"

func (p *Profiler) TargetLibPath(dist *Distribution, pid int) string {
	f := ProcFile{Path: dist.LibPath(), PID: pid}
	return f.ProcRootPath()
}
