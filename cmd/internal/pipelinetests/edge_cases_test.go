package pipelinetests

import (
	"testing"

	"github.com/grafana/agent/cmd/internal/pipelinetests/internal/framework"
)

/*
*
//TODO(thampiotr):
- Make a test with OTEL pipeline
- Make a test with loki.process
- Make a test with relabel rules
*
*/
func TestPipeline_WithEmptyConfig(t *testing.T) {
	framework.PipelineTest{
		ConfigFile:           "testdata/empty.river",
		RequireCleanShutdown: true,
	}.RunTest(t)
}

func TestPipeline_FileNotExists(t *testing.T) {
	framework.PipelineTest{
		ConfigFile:           "does_not_exist.river",
		CmdErrContains:       "does_not_exist.river: no such file or directory",
		RequireCleanShutdown: true,
	}.RunTest(t)
}

func TestPipeline_FileInvalid(t *testing.T) {
	framework.PipelineTest{
		ConfigFile:           "testdata/invalid.river",
		CmdErrContains:       "could not perform the initial load successfully",
		RequireCleanShutdown: true,
	}.RunTest(t)
}
