package stages

import (
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/model"
	"github.com/stretchr/testify/require"

	"github.com/grafana/loki/pkg/push"
	util_log "github.com/grafana/loki/pkg/util/log"
)

var pipelineStagesNonIndexedLabelsFromLogfmt = `
stage.logfmt {
	mapping = { "app" = ""}
}

stage.non_indexed_labels { 
	values = {"app" = ""}
}
`

var pipelineStagesNonIndexedLabelsFromJSON = `
stage.json {
	expressions = {app = ""}
}

stage.non_indexed_labels { 
	values = {"app" = ""}
}
`

var pipelineStagesNonIndexedLabelsWithRegexParser = `
stage.regex {
	expression = "^(?s)(?P<time>\\S+?) (?P<stream>stdout|stderr) (?P<flags>\\S+?) (?P<content>.*)$"
}

stage.non_indexed_labels { 
	values = {"stream" = ""}
}
`

var pipelineStagesNonIndexedLabelsFromJSONWithTemplate = `
stage.json {
	expressions = {app = ""}
}

stage.template {
    source   = "app"
    template = "{{ ToUpper .Value }}"
}

stage.non_indexed_labels { 
	values = {"app" = ""}
}
`

var pipelineStagesNonIndexedAndRegularLabelsFromJSON = `
stage.json {
	expressions = {app = "", component = "" }
}

stage.non_indexed_labels { 
	values = {"app" = ""}
}

stage.labels { 
	values = {"component" = ""}
}
`

func Test_NonIndexedLabelsStage(t *testing.T) {
	tests := map[string]struct {
		pipelineStagesYaml       string
		logLine                  string
		expectedNonIndexedLabels push.LabelsAdapter
		expectedLabels           model.LabelSet
	}{
		"expected non-indexed labels to be extracted with logfmt parser and to be added to entry": {
			pipelineStagesYaml:       pipelineStagesNonIndexedLabelsFromLogfmt,
			logLine:                  "app=loki component=ingester",
			expectedNonIndexedLabels: push.LabelsAdapter{push.LabelAdapter{Name: "app", Value: "loki"}},
		},
		"expected non-indexed labels to be extracted with json parser and to be added to entry": {
			pipelineStagesYaml:       pipelineStagesNonIndexedLabelsFromJSON,
			logLine:                  `{"app":"loki" ,"component":"ingester"}`,
			expectedNonIndexedLabels: push.LabelsAdapter{push.LabelAdapter{Name: "app", Value: "loki"}},
		},
		"expected non-indexed labels to be extracted with regexp parser and to be added to entry": {
			pipelineStagesYaml:       pipelineStagesNonIndexedLabelsWithRegexParser,
			logLine:                  `2019-01-01T01:00:00.000000001Z stderr P i'm a log message!`,
			expectedNonIndexedLabels: push.LabelsAdapter{push.LabelAdapter{Name: "stream", Value: "stderr"}},
		},
		"expected non-indexed labels to be extracted with json parser and to be added to entry after rendering the template": {
			pipelineStagesYaml:       pipelineStagesNonIndexedLabelsFromJSONWithTemplate,
			logLine:                  `{"app":"loki" ,"component":"ingester"}`,
			expectedNonIndexedLabels: push.LabelsAdapter{push.LabelAdapter{Name: "app", Value: "LOKI"}},
		},
		"expected non-indexed and regular labels to be extracted with json parser and to be added to entry": {
			pipelineStagesYaml:       pipelineStagesNonIndexedAndRegularLabelsFromJSON,
			logLine:                  `{"app":"loki" ,"component":"ingester"}`,
			expectedNonIndexedLabels: push.LabelsAdapter{push.LabelAdapter{Name: "app", Value: "loki"}},
			expectedLabels:           model.LabelSet{model.LabelName("component"): model.LabelValue("ingester")},
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			pl, err := NewPipeline(util_log.Logger, loadConfig(test.pipelineStagesYaml), nil, prometheus.DefaultRegisterer)
			require.NoError(t, err)

			result := processEntries(pl, newEntry(nil, nil, test.logLine, time.Now()))[0]
			require.Equal(t, test.expectedNonIndexedLabels, result.NonIndexedLabels)
			if test.expectedLabels != nil {
				require.Equal(t, test.expectedLabels, result.Labels)
			} else {
				require.Empty(t, result.Labels)
			}
		})
	}
}
