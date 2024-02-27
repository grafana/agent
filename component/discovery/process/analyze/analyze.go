package analyze

import (
	"debug/elf"
	"io"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
)

const (
	labelValueTrue  = "true"
	labelValueFalse = "false"
)

type Results struct {
	Labels map[string]string
}

type Input struct {
	PID     uint32
	PIDs    string
	File    io.ReaderAt
	ElfFile *elf.File
}

// analyzerFunc is called with a particular pid and a reader into its binary.
//
// If an error occurs analyzing the binary/process information it is returned.
// If there is strong evidence that this process has been detected, the
// analyzer can return io.EOF and it will skip all following analyzers.
type analyzerFunc func(input Input, analysis *Results) error

func Analyze(logger log.Logger, input Input) *Results {
	res := &Results{
		Labels: make(map[string]string),
	}
	for _, a := range []analyzerFunc{
		analyzeBinary,
		analyzeGo,
		analyzePython,
		analyzeDotNet,
		analyzeJava,
	} {
		if err := a(input, res); err == io.EOF {
			break
		} else if err != nil {
			level.Warn(logger).Log("msg", "error during", "func", "todo", "err", err)
		}
	}

	return res
}
