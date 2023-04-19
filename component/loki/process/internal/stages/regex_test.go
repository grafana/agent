package stages

// This package is ported over from grafana/loki/clients/pkg/logentry/stages.
// We aim to port the stages in steps, to avoid introducing huge amounts of
// new code without being able to slowly review, examine and test them.

import (
	"bytes"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/go-kit/log"
	"github.com/grafana/agent/pkg/util"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/model"
	"github.com/stretchr/testify/assert"
)

var protocolStr = "protocol"

var testRegexRiverSingleStageWithoutSource = `
stage.regex {
    expression =  "^(?P<ip>\\S+) (?P<identd>\\S+) (?P<user>\\S+) \\[(?P<timestamp>[\\w:/]+\\s[+\\-]\\d{4})\\] \"(?P<action>\\S+)\\s?(?P<path>\\S+)?\\s?(?P<protocol>\\S+)?\" (?P<status>\\d{3}|-) (?P<size>\\d+|-)\\s?\"?(?P<referer>[^\"]*)\"?\\s?\"?(?P<useragent>[^\"]*)?\"?$"
}
`

var testRegexRiverMultiStageWithSource = `
stage.regex {
    expression = "^(?P<ip>\\S+) (?P<identd>\\S+) (?P<user>\\S+) \\[(?P<timestamp>[\\w:/]+\\s[+\\-]\\d{4})\\] \"(?P<action>\\S+)\\s?(?P<path>\\S+)?\\s?(?P<protocol>\\S+)?\" (?P<status>\\d{3}|-) (?P<size>\\d+|-)\\s?\"?(?P<referer>[^\"]*)\"?\\s?\"?(?P<useragent>[^\"]*)?\"?$"
}
stage.regex {
    expression = "^HTTP\\/(?P<protocol_version>[0-9\\.]+)$"
    source     = "protocol"
}
`

var testRegexRiverSourceWithMissingKey = `
stage.json {
    expressions = { "time" = "" }
}
stage.regex {
    expression = "^(?P<year>\\d+)"
    source     = "time"
}
`

var testRegexLogLineWithMissingKey = `
{
	"app":"loki",
	"component": ["parser","type"],
	"level": "WARN"
}
`

var testRegexLogLine = `11.11.11.11 - frank [25/Jan/2000:14:00:01 -0500] "GET /1986.js HTTP/1.1" 200 932 "-" "Mozilla/5.0 (Windows; U; Windows NT 5.1; de; rv:1.9.1.7) Gecko/20091221 Firefox/3.5.7 GTB6"`

func init() {
	Debug = true
}

func TestPipeline_Regex(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		config          string
		entry           string
		expectedExtract map[string]interface{}
	}{
		"successfully run a pipeline with 1 regex stage without source": {
			testRegexRiverSingleStageWithoutSource,
			testRegexLogLine,
			map[string]interface{}{
				"ip":        "11.11.11.11",
				"identd":    "-",
				"user":      "frank",
				"timestamp": "25/Jan/2000:14:00:01 -0500",
				"action":    "GET",
				"path":      "/1986.js",
				"protocol":  "HTTP/1.1",
				"status":    "200",
				"size":      "932",
				"referer":   "-",
				"useragent": "Mozilla/5.0 (Windows; U; Windows NT 5.1; de; rv:1.9.1.7) Gecko/20091221 Firefox/3.5.7 GTB6",
			},
		},
		"successfully run a pipeline with 2 regex stages with source": {
			testRegexRiverMultiStageWithSource,
			testRegexLogLine,
			map[string]interface{}{
				"ip":               "11.11.11.11",
				"identd":           "-",
				"user":             "frank",
				"timestamp":        "25/Jan/2000:14:00:01 -0500",
				"action":           "GET",
				"path":             "/1986.js",
				"protocol":         "HTTP/1.1",
				"protocol_version": "1.1",
				"status":           "200",
				"size":             "932",
				"referer":          "-",
				"useragent":        "Mozilla/5.0 (Windows; U; Windows NT 5.1; de; rv:1.9.1.7) Gecko/20091221 Firefox/3.5.7 GTB6",
			},
		},
	}

	for testName, testData := range tests {
		testData := testData

		t.Run(testName, func(t *testing.T) {
			t.Parallel()

			logger := util.TestFlowLogger(t)
			pl, err := NewPipeline(logger, loadConfig(testData.config), nil, prometheus.DefaultRegisterer)
			if err != nil {
				t.Fatal(err)
			}

			out := processEntries(pl, newEntry(nil, nil, testData.entry, time.Now()))[0]
			assert.Equal(t, testData.expectedExtract, out.Extracted)
		})
	}
}

func TestPipelineWithMissingKey_Regex(t *testing.T) {
	var buf bytes.Buffer
	w := log.NewSyncWriter(&buf)
	logger := log.NewLogfmtLogger(w)
	pl, err := NewPipeline(logger, loadConfig(testRegexRiverSourceWithMissingKey), nil, prometheus.DefaultRegisterer)
	if err != nil {
		t.Fatal(err)
	}
	_ = processEntries(pl, newEntry(nil, nil, testRegexLogLineWithMissingKey, time.Now()))[0]

	expectedLog := "level=debug component=stage type=regex msg=\"failed to convert source value to string\" source=time err=\"Can't convert <nil> to string\" type=null"
	if !(strings.Contains(buf.String(), expectedLog)) {
		t.Errorf("\nexpected: %s\n+actual: %s", expectedLog, buf.String())
	}
}

func TestRegexConfig_validate(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		config interface{}
		err    error
	}{
		"empty config": {
			nil,
			ErrExpressionRequired,
		},
		"missing regex_expression": {
			map[string]interface{}{},
			ErrExpressionRequired,
		},
		"invalid regex_expression": {
			map[string]interface{}{
				"expression": "(?P<ts[0-9]+).*",
			},
			errors.New(ErrCouldNotCompileRegex.Error() + ": error parsing regexp: invalid named capture: `(?P<ts[0-9]+).*`"),
		},
		"empty source": {
			map[string]interface{}{
				"expression": "(?P<ts>[0-9]+).*",
				"source":     "",
			},
			ErrEmptyRegexStageSource,
		},
		"valid without source": {
			map[string]interface{}{
				"expression": "(?P<ts>[0-9]+).*",
			},
			nil,
		},
		"valid with source": {
			map[string]interface{}{
				"expression": "(?P<ts>[0-9]+).*",
				"source":     "log",
			},
			nil,
		},
	}
	for tName, tt := range tests {
		tt := tt
		t.Run(tName, func(t *testing.T) {
			c, err := parseRegexConfig(tt.config)
			if err != nil {
				t.Fatalf("failed to create config: %s", err)
			}
			_, err = validateRegexConfig(*c)
			if (err != nil) != (tt.err != nil) {
				t.Errorf("RegexConfig.validate() expected error = %v, actual error = %v", tt.err, err)
				return
			}
			if (err != nil) && (err.Error() != tt.err.Error()) {
				t.Errorf("RegexConfig.validate() expected error = %v, actual error = %v", tt.err, err)
				return
			}
		})
	}
}

var regexLogFixture = `11.11.11.11 - frank [25/Jan/2000:14:00:01 -0500] "GET /1986.js HTTP/1.1" 200 932 "-" "Mozilla/5.0 (Windows; U; Windows NT 5.1; de; rv:1.9.1.7) Gecko/20091221 Firefox/3.5.7 GTB6"`

func TestRegexParser_Parse(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		config          RegexConfig
		extracted       map[string]interface{}
		entry           string
		expectedExtract map[string]interface{}
	}{
		"successfully match expression on entry": {
			RegexConfig{
				Expression: "^(?P<ip>\\S+) (?P<identd>\\S+) (?P<user>\\S+) \\[(?P<timestamp>[\\w:/]+\\s[+\\-]\\d{4})\\] \"(?P<action>\\S+)\\s?(?P<path>\\S+)?\\s?(?P<protocol>\\S+)?\" (?P<status>\\d{3}|-) (?P<size>\\d+|-)\\s?\"?(?P<referer>[^\"]*)\"?\\s?\"?(?P<useragent>[^\"]*)?\"?$",
			},
			map[string]interface{}{},
			regexLogFixture,
			map[string]interface{}{
				"ip":        "11.11.11.11",
				"identd":    "-",
				"user":      "frank",
				"timestamp": "25/Jan/2000:14:00:01 -0500",
				"action":    "GET",
				"path":      "/1986.js",
				"protocol":  "HTTP/1.1",
				"status":    "200",
				"size":      "932",
				"referer":   "-",
				"useragent": "Mozilla/5.0 (Windows; U; Windows NT 5.1; de; rv:1.9.1.7) Gecko/20091221 Firefox/3.5.7 GTB6",
			},
		},
		"successfully match expression on extracted[source]": {
			RegexConfig{
				Expression: "^HTTP\\/(?P<protocol_version>.*)$",
				Source:     &protocolStr,
			},
			map[string]interface{}{
				"protocol": "HTTP/1.1",
			},
			regexLogFixture,
			map[string]interface{}{
				"protocol":         "HTTP/1.1",
				"protocol_version": "1.1",
			},
		},
		"failed to match expression on entry": {
			RegexConfig{
				Expression: "^(?s)(?P<time>\\S+?) (?P<stream>stdout|stderr) (?P<flags>\\S+?) (?P<message>.*)$",
			},
			map[string]interface{}{},
			"blahblahblah",
			map[string]interface{}{},
		},
		"failed to match expression on extracted[source]": {
			RegexConfig{
				Expression: "^HTTP\\/(?P<protocol_version>.*)$",
				Source:     &protocolStr,
			},
			map[string]interface{}{
				"protocol": "unknown",
			},
			"unknown/unknown",
			map[string]interface{}{
				"protocol": "unknown",
			},
		},
		"case insensitive": {
			RegexConfig{
				Expression: "(?i)(?P<bad>panic:|core_dumped|failure|error|attack| bad |illegal |denied|refused|unauthorized|fatal|failed|Segmentation Fault|Corrupted)",
			},
			map[string]interface{}{},
			"A Terrible Error has occurred!!!",
			map[string]interface{}{
				"bad": "Error",
			},
		},
		"missing extracted[source]": {
			RegexConfig{
				Expression: "^HTTP\\/(?P<protocol_version>.*)$",
				Source:     &protocolStr,
			},
			map[string]interface{}{},
			"blahblahblah",
			map[string]interface{}{},
		},
		"invalid data type in extracted[source]": {
			RegexConfig{
				Expression: "^HTTP\\/(?P<protocol_version>.*)$",
				Source:     &protocolStr,
			},
			map[string]interface{}{
				"protocol": true,
			},
			"unknown/unknown",
			map[string]interface{}{
				"protocol": true,
			},
		},
	}
	for tName, tt := range tests {
		tt := tt
		t.Run(tName, func(t *testing.T) {
			t.Parallel()
			logger := util.TestFlowLogger(t)
			p, err := New(logger, nil, StageConfig{RegexConfig: &tt.config}, nil)
			if err != nil {
				t.Fatalf("failed to create regex parser: %s", err)
			}
			out := processEntries(p, newEntry(tt.extracted, nil, tt.entry, time.Now()))[0]
			assert.Equal(t, tt.expectedExtract, out.Extracted)
		})
	}
}

func BenchmarkRegexStage(b *testing.B) {
	benchmarks := []struct {
		name   string
		config RegexConfig
		entry  string
	}{
		{"apache common log",
			RegexConfig{
				Expression: "^(?P<ip>\\S+) (?P<identd>\\S+) (?P<user>\\S+) \\[(?P<timestamp>[\\w:/]+\\s[+\\-]\\d{4})\\] \"(?P<action>\\S+)\\s?(?P<path>\\S+)?\\s?(?P<protocol>\\S+)?\" (?P<status>\\d{3}|-) (?P<size>\\d+|-)\\s?\"?(?P<referer>[^\"]*)\"?\\s?\"?(?P<useragent>[^\"]*)?\"?$"},
			regexLogFixture,
		},
	}
	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			logger := util.TestFlowLogger(b)
			stage, err := New(logger, nil, StageConfig{RegexConfig: &bm.config}, nil)
			if err != nil {
				panic(err)
			}
			labels := model.LabelSet{}
			ts := time.Now()
			extr := map[string]interface{}{}

			in := make(chan Entry)
			out := stage.Run(in)
			go func() {
				for range out {
				}
			}()
			for i := 0; i < b.N; i++ {
				in <- newEntry(extr, labels, bm.entry, ts)
			}
			close(in)
		})
	}
}
