// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2022 Datadog, Inc.

// Reading the fuzz corpus from testdata/ during CI fails on Windows runners
// using Go 1.18, due to carriage return/line feed issues. This is fixed in Go
// 1.19 (see https://go.dev/cl/402074), but we can just skip these tests on Go
// 1.18 + Windows.
//go:build go1.19 || (!windows && go1.18)

package fastdelta_test

import (
	"io"
	"testing"

	"github.com/grafana/agent/component/pyroscope/scrape/internal/fastdelta"
)

// FuzzDelta looks for inputs to delta which cause crashes. This is to account
// for the possibility that the profile format changes in some way, or violates
// any hard-coded assumptions.
func FuzzDelta(f *testing.F) {
	f.Fuzz(func(t *testing.T, b []byte) {
		dc := fastdelta.NewDeltaComputer()
		dc.Delta(b, io.Discard)
	})
}
