package pipelinetests

import (
	"github.com/grafana/agent/cmd/internal/pipelinetests/internal/framework"
	"testing"
)

/*
*
//TODO(thampiotr):
- Make a test with logging pipeline
- Make a test with OTEL pipeline
- Make a test with loki.process
- Make a test with relabel rules
*
*/
func TestPipeline_WithEmptyConfig(t *testing.T) {
	framework.RunPipelineTest(t, framework.PipelineTest{
		ConfigFile:           "testdata/empty.river",
		RequireCleanShutdown: true,
	})
}

func TestPipeline_FileNotExists(t *testing.T) {
	framework.RunPipelineTest(t, framework.PipelineTest{
		ConfigFile:           "does_not_exist.river",
		CmdErrContains:       "does_not_exist.river: no such file or directory",
		RequireCleanShutdown: true,
	})
}

func TestPipeline_FileInvalid(t *testing.T) {
	framework.RunPipelineTest(t, framework.PipelineTest{
		ConfigFile:           "testdata/invalid.river",
		CmdErrContains:       "could not perform the initial load successfully",
		RequireCleanShutdown: true,
	})
}
