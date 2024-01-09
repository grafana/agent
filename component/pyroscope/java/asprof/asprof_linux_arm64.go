//go:build linux && arm64

package asprof

//go:embed async-profiler-3.0-ea-linux-arm64.tar.gz
var tarGzArchive []byte

// asprof
// glibc / libasyncProfiler.so
// musl / libasyncProfiler.so

var version = 300
