package stages

// This package is ported over from grafana/loki/clients/pkg/logentry/stages.
// We aim to port the stages in steps, to avoid introducing huge amounts of
// new code without being able to slowly review, examine and test them.

import (
	"fmt"
	"testing"
	"time"

	"github.com/grafana/agent/pkg/util"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
)

var testMatchRiver = `
stage.json {
		expressions = { "app" = "" }
}

stage.labels {
		values = { "app" = "" }
}

stage.match {
		selector = "{app=\"loki\"}"
		stage.json {
				expressions = { "msg" = "message" }
		}
		action = "keep"
}

stage.match {
		pipeline_name = "app2"
		selector = "{app=\"poki\"}"
		stage.json {
				expressions = { "msg" = "msg" }
		}
		action = "keep"
}

stage.output {
		source = "msg"
}
`

var testMatchLogLineApp1 = `
{
	"time":"2012-11-01T22:08:41+00:00",
	"app":"loki",
	"component": ["parser","type"],
	"level" : "WARN",
	"message" : "app1 log line"
}
`

var testMatchLogLineApp2 = `
{
	"time":"2012-11-01T22:08:41+00:00",
	"app":"poki",
	"component": ["parser","type"],
	"level" : "WARN",
	"msg" : "app2 log line"
}
`

func TestMatchStage(t *testing.T) {
	registry := prometheus.NewRegistry()
	plName := "test_match_pipeline"
	logger := util.TestFlowLogger(t)
	pl, err := NewPipeline(logger, loadConfig(testMatchRiver), &plName, registry)
	if err != nil {
		t.Fatal(err)
	}

	in := make(chan Entry)

	out := pl.Run(in)

	in <- newEntry(nil, nil, testMatchLogLineApp1, time.Now())

	e := <-out

	assert.Equal(t, "app1 log line", e.Line)

	// Process the second log line which should extract the output from the `msg` field
	e.Line = testMatchLogLineApp2
	e.Extracted = map[string]interface{}{}
	in <- e
	e = <-out
	assert.Equal(t, "app2 log line", e.Line)
	close(in)
}

func TestMatcher(t *testing.T) {
	t.Parallel()
	tests := []struct {
		selector string
		labels   map[string]string
		action   string

		shouldDrop bool
		shouldRun  bool
		wantErr    bool
	}{
		{`{foo="bar"} |= "foo"`, map[string]string{"foo": "bar"}, MatchActionKeep, false, true, false},
		{`{foo="bar"} |~ "foo"`, map[string]string{"foo": "bar"}, MatchActionKeep, false, true, false},
		{`{foo="bar"} |= "bar"`, map[string]string{"foo": "bar"}, MatchActionKeep, false, false, false},
		{`{foo="bar"} |~ "bar"`, map[string]string{"foo": "bar"}, MatchActionKeep, false, false, false},
		{`{foo="bar"} != "bar"`, map[string]string{"foo": "bar"}, MatchActionKeep, false, true, false},
		{`{foo="bar"} !~ "bar"`, map[string]string{"foo": "bar"}, MatchActionKeep, false, true, false},
		{`{foo="bar"} != "foo"`, map[string]string{"foo": "bar"}, MatchActionKeep, false, false, false},
		{`{foo="bar"} |= "foo"`, map[string]string{"foo": "bar"}, MatchActionDrop, true, false, false},
		{`{foo="bar"} |~ "foo"`, map[string]string{"foo": "bar"}, MatchActionDrop, true, false, false},
		{`{foo="bar"} |= "bar"`, map[string]string{"foo": "bar"}, MatchActionDrop, false, false, false},
		{`{foo="bar"} |~ "bar"`, map[string]string{"foo": "bar"}, MatchActionDrop, false, false, false},
		{`{foo="bar"} != "bar"`, map[string]string{"foo": "bar"}, MatchActionDrop, true, false, false},
		{`{foo="bar"} !~ "bar"`, map[string]string{"foo": "bar"}, MatchActionDrop, true, false, false},
		{`{foo="bar"} != "foo"`, map[string]string{"foo": "bar"}, MatchActionDrop, false, false, false},
		{`{foo="bar"} !~ "[]"`, map[string]string{"foo": "bar"}, MatchActionDrop, false, false, true},
		{"foo", map[string]string{"foo": "bar"}, MatchActionKeep, false, false, true},
		{"{}", map[string]string{"foo": "bar"}, MatchActionKeep, false, false, true},
		{"{", map[string]string{"foo": "bar"}, MatchActionKeep, false, false, true},
		{"", map[string]string{"foo": "bar"}, MatchActionKeep, false, true, true},
		{`{foo="bar"}`, map[string]string{"foo": "bar"}, MatchActionKeep, false, true, false},
		{`{foo=""}`, map[string]string{"foo": "bar"}, MatchActionKeep, false, false, false},
		{`{foo=""}`, map[string]string{}, MatchActionKeep, false, true, false},
		{`{foo!="bar"}`, map[string]string{"foo": "bar"}, MatchActionKeep, false, false, false},
		{`{foo!="bar"}`, map[string]string{"foo": "bar"}, MatchActionDrop, false, false, false},
		{`{foo="bar",bar!="test"}`, map[string]string{"foo": "bar"}, MatchActionKeep, false, true, false},
		{`{foo="bar",bar!="test"}`, map[string]string{"foo": "bar"}, MatchActionDrop, true, false, false},
		{`{foo="bar",bar!="test"}`, map[string]string{"foo": "bar", "bar": "test"}, MatchActionKeep, false, false, false},
		{`{foo="bar",bar=~"te.*"}`, map[string]string{"foo": "bar", "bar": "test"}, MatchActionDrop, true, false, false},
		{`{foo="bar",bar=~"te.*"}`, map[string]string{"foo": "bar", "bar": "test"}, MatchActionKeep, false, true, false},
		{`{foo="bar",bar!~"te.*"}`, map[string]string{"foo": "bar", "bar": "test"}, MatchActionKeep, false, false, false},
		{`{foo="bar",bar!~"te.*"}`, map[string]string{"foo": "bar", "bar": "test"}, MatchActionDrop, false, false, false},

		{`{foo=""}`, map[string]string{}, MatchActionKeep, false, true, false},
	}

	for _, tt := range tests {
		name := fmt.Sprintf("%s/%s/%s", tt.selector, tt.labels, tt.action)

		t.Run(name, func(t *testing.T) {
			// Build a match config which has a simple label stage that when matched will add the test_label to
			// the labels in the pipeline.
			var stages []StageConfig
			if tt.action != MatchActionDrop {
				stages = []StageConfig{
					{
						LabelsConfig: &LabelsConfig{
							Values: map[string]*string{"test_label": nil},
						},
					},
				}
			}
			matchConfig := MatchConfig{
				tt.selector,
				stages,
				tt.action,
				"",
				"",
			}
			logger := util.TestFlowLogger(t)
			s, err := newMatcherStage(logger, nil, matchConfig, prometheus.DefaultRegisterer)
			if (err != nil) != tt.wantErr {
				t.Errorf("withMatcher() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if s != nil {
				out := processEntries(s, newEntry(map[string]interface{}{
					"test_label": "unimportant value",
				}, toLabelSet(tt.labels), "foo", time.Now()))

				if tt.shouldDrop {
					if len(out) != 0 {
						t.Errorf("stage should have been dropped but got %v", out)
					}
					return
				}
				// test_label should only be in the label set if the stage ran
				if _, ok := out[0].Labels["test_label"]; ok {
					if !tt.shouldRun {
						t.Error("stage ran but should have not")
					}
				}
			}
		})
	}
}

func TestValidateMatcherConfig(t *testing.T) {
	emptyStages := []StageConfig{}
	defaultStage := []StageConfig{{MatchConfig: &MatchConfig{}}}
	tests := []struct {
		name    string
		cfg     *MatchConfig
		wantErr bool
	}{
		{"pipeline name required", &MatchConfig{}, true},
		{"selector required", &MatchConfig{Selector: ""}, true},
		{"nil stages without dropping", &MatchConfig{PipelineName: "", Selector: `{app="foo"}`, Action: MatchActionKeep, Stages: nil}, true},
		{"empty stages without dropping", &MatchConfig{Selector: `{app="foo"}`, Action: MatchActionKeep, Stages: emptyStages}, true},
		{"stages with dropping", &MatchConfig{Selector: `{app="foo"}`, Action: MatchActionDrop, Stages: defaultStage}, true},
		{"empty stages dropping", &MatchConfig{Selector: `{app="foo"}`, Action: MatchActionDrop, Stages: emptyStages}, false},
		{"stages without dropping", &MatchConfig{Selector: `{app="foo"}`, Action: MatchActionKeep, Stages: defaultStage}, false},
		{"bad selector", &MatchConfig{Selector: `{app="foo}`, Action: MatchActionKeep, Stages: defaultStage}, true},
		{"bad action", &MatchConfig{Selector: `{app="foo}`, Action: "nope", Stages: emptyStages}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := validateMatcherConfig(tt.cfg)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateMatcherConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
