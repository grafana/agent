//go:build linux && amd64

package asprof

import (
	_ "embed"
)

//go:embed async-profiler-3.0-ea-linux-x64.tar.gz
var glibcArchive []byte

var glibcDist = &Distribution{
	targz:   glibcArchive,
	fname:   "async-profiler-3.0-ea-linux-x64.tar.gz",
	version: 300,
}

//go:embed async-profiler-3.0-ea-linux-musl-x64.tar.gz
var muslArchive []byte

var muslDist = &Distribution{
	targz:   muslArchive,
	fname:   "async-profiler-3.0-ea-linux-musl-x64.tar.gz",
	version: 300,
}