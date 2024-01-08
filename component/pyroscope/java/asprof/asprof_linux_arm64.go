//go:build linux && arm64

package asprof

//go:embed async-profiler-2.9-linux-arm64.tar.gz
var tarGzArchive []byte

// profiler.sh
// jattach
// glibc / libasyncProfiler.so
// musl / libasyncProfiler.so

var version = 209
