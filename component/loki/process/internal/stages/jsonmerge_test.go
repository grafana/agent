package stages

import (
	"encoding/json"
	"reflect"
	"testing"
	"time"

	"github.com/grafana/agent/pkg/util"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var testJSONMergeRiver = `
stage.jsonmerge {
	source = ""
	values = ["number", "my_string"]
	output = "multi_values_result"
}
stage.jsonmerge {
	source = ""
	values = ["object"]
	output = "object_result"
}`

func TestJSONMergeStageProcess(t *testing.T) {
	registry := prometheus.NewRegistry()
	logger := util.TestFlowLogger(t)

	pl, err := NewPipeline(logger, loadConfig(testJSONMergeRiver), &plName, registry)
	require.NoError(t, err)

	ls := model.LabelSet{
		"filename": "hello.log",
	}
	testTime := time.Now()
	out := processEntries(pl,
		newEntry(map[string]interface{}{"number": 1, "my_string": "i'm a string"}, ls, `{"hello":"world"}`, testTime),
		newEntry(map[string]interface{}{"object": map[string]interface{}{"numbers": []int{1, 2, 3}}}, ls, `{"number":1}`, testTime),
		newEntry(map[string]interface{}{"my_string": "i'm a string"}, ls, `{"hello":{"object":true}}`, testTime),
	)

	mustJSONUnmarshal := func(data any) map[string]interface{} {
		d, ok := data.(string)
		require.True(t, ok, "data must be a string")
		var m map[string]interface{}
		require.NoError(t, json.Unmarshal([]byte(d), &m))
		return m
	}

	require.Len(t, out, 3, "should equal number of entries")
	assert.Equal(t,
		map[string]interface{}{
			"hello":     "world",
			"number":    float64(1),
			"my_string": "i'm a string",
		},
		mustJSONUnmarshal(out[0].Extracted["multi_values_result"]),
	)
	assert.Equal(t,
		map[string]interface{}{
			"number": float64(1),
			"object": map[string]interface{}{
				"numbers": []interface{}{float64(1), float64(2), float64(3)},
			},
		},
		mustJSONUnmarshal(out[1].Extracted["object_result"]),
	)
	assert.Equal(t,
		map[string]interface{}{
			"my_string": "i'm a string",
			"hello": map[string]interface{}{
				"object": true,
			}},
		mustJSONUnmarshal(out[2].Extracted["multi_values_result"]),
	)
}

func TestExtractSourceObject(t *testing.T) {
	type args struct {
		extracted map[string]interface{}
		entry     *string
	}
	tests := []struct {
		name    string
		source  string
		args    args
		want    map[string]interface{}
		wantErr bool
	}{
		{
			name:   "extract from log line when source is empty",
			source: "",
			args: args{
				extracted: map[string]interface{}{},
				entry:     toPtr(`{"hello": "world"}`),
			},
			want: map[string]interface{}{
				"hello": "world",
			},
			wantErr: false,
		},
		{
			name:   "extract from log line which is not JSON should error",
			source: "",
			args: args{
				extracted: map[string]interface{}{},
				entry:     toPtr(`hello="world"`),
			},
			want:    nil,
			wantErr: true,
		},
		{
			name:   "extract string from extracted",
			source: "hello",
			args: args{
				extracted: map[string]interface{}{
					"hello":     "world",
					"something": "else",
				},
				entry: nil,
			},
			want: map[string]interface{}{
				"hello": "world",
			},
			wantErr: false,
		},
		{
			name:   "extract object from extracted",
			source: "hello",
			args: args{
				extracted: map[string]interface{}{
					"hello": map[string]interface{}{
						"world": []int{1, 2, 3},
					},
				},
				entry: nil,
			},
			want: map[string]interface{}{
				"hello": map[string]interface{}{
					"world": []int{1, 2, 3},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			j := &jsonmergeStage{
				config: JSONMergeConfig{Source: tt.source},
			}
			got, err := j.extractSourceObject(tt.args.extracted, tt.args.entry)
			if (err != nil) != tt.wantErr {
				t.Errorf("jsonmergeStage.extractSourceObject() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("jsonmergeStage.extractSourceObject() = %v, want %v", got, tt.want)
			}
		})
	}
}

func toPtr[T any](t T) *T {
	return &t
}
