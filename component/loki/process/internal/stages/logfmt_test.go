package stages

// This package is ported over from grafana/loki/clients/pkg/logentry/stages.
// We aim to port the stages in steps, to avoid introducing huge amounts of
// new code without being able to slowly review, examine and test them.

import (
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"

	"github.com/grafana/agent/pkg/util"
	util_log "github.com/grafana/loki/pkg/util/log"
)

var testLogfmtRiverSingleStageWithoutSource = `
stage.logfmt {
		mapping = { "out" = "message", "app" = "", "duration" = "", "unknown" = "" }
}`

var testLogfmtRiverMultiStageWithSource = `
stage.logfmt {
		mapping = { "extra" = "" }
}
stage.logfmt {
		mapping = { "user" = "" }
		source  = "extra"
}`

func TestLogfmt(t *testing.T) {
	var testLogfmtLogLine = `
		time=2012-11-01T22:08:41+00:00 app=loki	level=WARN duration=125 message="this is a log line" extra="user=foo""
	`
	t.Parallel()

	tests := map[string]struct {
		config          string
		entry           string
		expectedExtract map[string]interface{}
	}{
		"successfully run a pipeline with 1 logfmt stage without source": {
			testLogfmtRiverSingleStageWithoutSource,
			testLogfmtLogLine,
			map[string]interface{}{
				"out":      "this is a log line",
				"app":      "loki",
				"duration": "125",
			},
		},
		"successfully run a pipeline with 2 logfmt stages with source": {
			testLogfmtRiverMultiStageWithSource,
			testLogfmtLogLine,
			map[string]interface{}{
				"extra": "user=foo",
				"user":  "foo",
			},
		},
	}

	for testName, testData := range tests {
		testData := testData

		t.Run(testName, func(t *testing.T) {
			t.Parallel()

			pl, err := NewPipeline(util_log.Logger, loadConfig(testData.config), nil, prometheus.DefaultRegisterer)
			assert.NoError(t, err)
			out := processEntries(pl, newEntry(nil, nil, testData.entry, time.Now()))[0]
			assert.Equal(t, testData.expectedExtract, out.Extracted)
		})
	}
}

func TestLogfmtConfigValidation(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		config           LogfmtConfig
		wantMappingCount int
		err              error
	}{
		"no mapping": {
			LogfmtConfig{},
			0,
			ErrMappingRequired,
		},
		"valid without source": {
			LogfmtConfig{
				Mapping: map[string]string{
					"foo1": "foo",
					"foo2": "",
				},
			},
			2,
			nil,
		},
		"valid with source": {
			LogfmtConfig{
				Mapping: map[string]string{
					"foo1": "foo",
					"foo2": "",
				},
				Source: "log",
			},
			2,
			nil,
		},
	}
	for tName, tt := range tests {
		tt := tt
		t.Run(tName, func(t *testing.T) {
			got, err := validateLogfmtConfig(&tt.config)
			if tt.err != nil {
				assert.EqualError(t, err, tt.err.Error())
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.wantMappingCount, len(got))
		})
	}
}

var testLogfmtLogFixture = `
	time=2012-11-01T22:08:41+00:00
	app=loki
	level=WARN
	nested="child=value"
	message="this is a log line"
`

func TestLogfmtParser_Parse(t *testing.T) {
	t.Parallel()
	logger := util.TestFlowLogger(t)
	tests := map[string]struct {
		config          LogfmtConfig
		extracted       map[string]interface{}
		entry           string
		expectedExtract map[string]interface{}
	}{
		"successfully decode logfmt on entry": {
			LogfmtConfig{
				Mapping: map[string]string{
					"time":    "",
					"app":     "",
					"level":   "",
					"nested":  "",
					"message": "",
				},
			},
			map[string]interface{}{},
			testLogfmtLogFixture,
			map[string]interface{}{
				"time":    "2012-11-01T22:08:41+00:00",
				"app":     "loki",
				"level":   "WARN",
				"nested":  "child=value",
				"message": "this is a log line",
			},
		},
		"successfully decode logfmt on extracted[source]": {
			LogfmtConfig{
				Mapping: map[string]string{
					"time":    "",
					"app":     "",
					"level":   "",
					"nested":  "",
					"message": "",
				},
				Source: "log",
			},
			map[string]interface{}{
				"log": testLogfmtLogFixture,
			},
			"{}",
			map[string]interface{}{
				"time":    "2012-11-01T22:08:41+00:00",
				"app":     "loki",
				"level":   "WARN",
				"nested":  "child=value",
				"message": "this is a log line",
				"log":     testLogfmtLogFixture,
			},
		},
		"missing extracted[source]": {
			LogfmtConfig{
				Mapping: map[string]string{
					"app": "",
				},
				Source: "log",
			},
			map[string]interface{}{},
			testLogfmtLogFixture,
			map[string]interface{}{},
		},
		"invalid logfmt on entry": {
			LogfmtConfig{
				Mapping: map[string]string{
					"expr1": "",
				},
			},
			map[string]interface{}{},
			"{\"invalid\":\"logfmt\"}",
			map[string]interface{}{},
		},
		"invalid logfmt on extracted[source]": {
			LogfmtConfig{
				Mapping: map[string]string{
					"app": "",
				},
				Source: "log",
			},
			map[string]interface{}{
				"log": "not logfmt",
			},
			testLogfmtLogFixture,
			map[string]interface{}{
				"log": "not logfmt",
			},
		},
		"nil source": {
			LogfmtConfig{
				Mapping: map[string]string{
					"app": "",
				},
				Source: "log",
			},
			map[string]interface{}{
				"log": nil,
			},
			testLogfmtLogFixture,
			map[string]interface{}{
				"log": nil,
			},
		},
	}
	for tName, tt := range tests {
		tt := tt
		t.Run(tName, func(t *testing.T) {
			t.Parallel()
			p, err := New(logger, nil, StageConfig{LogfmtConfig: &tt.config}, nil)
			assert.NoError(t, err)
			out := processEntries(p, newEntry(tt.extracted, nil, tt.entry, time.Now()))[0]

			assert.Equal(t, tt.expectedExtract, out.Extracted)
		})
	}
}
