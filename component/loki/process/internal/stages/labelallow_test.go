package stages

// This package is ported over from grafana/loki/clients/pkg/logentry/stages.
// We aim to port the stages in steps, to avoid introducing huge amounts of
// new code without being able to slowly review, examine and test them.

import (
	"testing"
	"time"

	"github.com/prometheus/common/model"
	"github.com/stretchr/testify/assert"
)

func TestLabelAllow(t *testing.T) {
	tests := []struct {
		name           string
		config         *LabelAllowConfig
		inputLabels    model.LabelSet
		expectedLabels model.LabelSet
	}{
		{
			name:   "allow single label",
			config: &LabelAllowConfig{Values: []string{"testLabel1"}},
			inputLabels: model.LabelSet{
				"testLabel1": "testValue",
				"testLabel2": "testValue",
			},
			expectedLabels: model.LabelSet{
				"testLabel1": "testValue",
			},
		},
		{
			name:   "allow multiple labels",
			config: &LabelAllowConfig{Values: []string{"testLabel1", "testLabel2"}},
			inputLabels: model.LabelSet{
				"testLabel1": "testValue",
				"testLabel2": "testValue",
				"testLabel3": "testValue",
			},
			expectedLabels: model.LabelSet{
				"testLabel1": "testValue",
				"testLabel2": "testValue",
			},
		},
		{
			name:   "allow non-existing label",
			config: &LabelAllowConfig{Values: []string{"foobar"}},
			inputLabels: model.LabelSet{
				"testLabel1": "testValue",
				"testLabel2": "testValue",
			},
			expectedLabels: model.LabelSet{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			st, err := newLabelAllowStage(*test.config)
			if err != nil {
				t.Fatal(err)
			}
			out := processEntries(st, newEntry(nil, test.inputLabels, "", time.Now()))[0]
			assert.Equal(t, test.expectedLabels, out.Labels)
		})
	}
}
