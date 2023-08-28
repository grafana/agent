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

func TestLabelDrop(t *testing.T) {
	tests := []struct {
		name           string
		config         *LabelDropConfig
		inputLabels    model.LabelSet
		expectedLabels model.LabelSet
	}{
		{
			name:   "drop one label",
			config: &LabelDropConfig{Values: []string{"testLabel1"}},
			inputLabels: model.LabelSet{
				"testLabel1": "testValue",
				"testLabel2": "testValue",
			},
			expectedLabels: model.LabelSet{
				"testLabel2": "testValue",
			},
		},
		{
			name:   "drop two labels",
			config: &LabelDropConfig{Values: []string{"testLabel1", "testLabel2"}},
			inputLabels: model.LabelSet{
				"testLabel1": "testValue",
				"testLabel2": "testValue",
			},
			expectedLabels: model.LabelSet{},
		},
		{
			name:   "drop non-existing label",
			config: &LabelDropConfig{Values: []string{"foobar"}},
			inputLabels: model.LabelSet{
				"testLabel1": "testValue",
				"testLabel2": "testValue",
			},
			expectedLabels: model.LabelSet{
				"testLabel1": "testValue",
				"testLabel2": "testValue",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			st, err := newLabelDropStage(*test.config)
			if err != nil {
				t.Fatal(err)
			}
			out := processEntries(st, newEntry(nil, test.inputLabels, "", time.Now()))[0]
			assert.Equal(t, test.expectedLabels, out.Labels)
		})
	}
}
