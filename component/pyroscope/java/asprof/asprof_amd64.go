package asprof

import _ "embed"

//go:embed async-profiler-3.0-ea-linux-x64.tar.gz
var glibcDistribution []byte
var glibcDistributionName = "async-profiler-3.0-ea-linux-x64"

// todo
var muslDistribution []byte
var muslDistributionName = "TODO"
