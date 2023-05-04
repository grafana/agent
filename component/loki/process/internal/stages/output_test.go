package stages

// This package is ported over from grafana/loki/clients/pkg/logentry/stages.
// We aim to port the stages in steps, to avoid introducing huge amounts of
// new code without being able to slowly review, examine and test them.

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/go-kit/log"
	"github.com/grafana/agent/pkg/util"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var testOutputRiver = `
stage.json {
    expressions = { "out" = "message" }
}
stage.output {
    source = "out"
}
`

var testOutputLogLine = `
{
	"time":"2012-11-01T22:08:41+00:00",
	"app":"loki",
	"component": ["parser","type"],
	"level" : "WARN",
	"nested" : {"child":"value"},
	"message" : "this is a log line"
}
`
var testOutputLogLineWithMissingKey = `
{
	"time":"2012-11-01T22:08:41+00:00",
	"app":"loki",
	"component": ["parser","type"],
	"level" : "WARN",
	"nested" : {"child":"value"}
}
`

func TestPipeline_Output(t *testing.T) {
	logger := util.TestFlowLogger(t)
	pl, err := NewPipeline(logger, loadConfig(testOutputRiver), nil, prometheus.DefaultRegisterer)
	require.NoError(t, err)

	out := processEntries(pl, newEntry(nil, nil, testOutputLogLine, time.Now()))[0]
	assert.Equal(t, "this is a log line", out.Line)
}

func TestPipelineWithMissingKey_Output(t *testing.T) {
	var buf bytes.Buffer
	w := log.NewSyncWriter(&buf)
	logger := log.NewLogfmtLogger(w)
	pl, err := NewPipeline(logger, loadConfig(testOutputRiver), nil, prometheus.DefaultRegisterer)
	require.NoError(t, err)

	_ = processEntries(pl, newEntry(nil, nil, testOutputLogLineWithMissingKey, time.Now()))
	expectedLog := "level=debug msg=\"extracted output could not be converted to a string\" err=\"Can't convert <nil> to string\" type=null"
	if !(strings.Contains(buf.String(), expectedLog)) {
		t.Errorf("\nexpected: %s\n+actual: %s", expectedLog, buf.String())
	}
}

func TestOutputValidation(t *testing.T) {
	emptyConfig := OutputConfig{Source: ""}
	_, err := newOutputStage(nil, emptyConfig)
	require.Equal(t, err, ErrOutputSourceRequired)
}

func TestOutputStage_Process(t *testing.T) {
	cfg := OutputConfig{
		Source: "out",
	}
	extractedValues := map[string]interface{}{
		"something": "notimportant",
		"out":       "outmessage",
	}
	wantOutput := "outmessage"

	st, err := newOutputStage(nil, cfg)
	require.NoError(t, err)
	out := processEntries(st, newEntry(extractedValues, nil, "replaceme", time.Time{}))[0]

	assert.Equal(t, wantOutput, out.Line)
}
