//go:build linux && amd64

package asprof

import (
	_ "embed"
)

//go:embed async-profiler-2.9-linux-x64.tar.gz
var tarGzArchive []byte

// profiler.sh
// jattach
// glibc / libasyncProfiler.so
// musl / libasyncProfiler.so

var version = 209
