package stages

// This package is ported over from grafana/loki/clients/pkg/logentry/stages.
// We aim to port the stages in steps, to avoid introducing huge amounts of
// new code without being able to slowly review, examine and test them.

import (
	"fmt"
	"testing"
	"time"

	"github.com/alecthomas/units"
	"github.com/grafana/agent/pkg/util"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	ww "github.com/weaveworks/common/server"
)

// Not all these are tested but are here to make sure the different types marshal without error
var testDropRiver = `
stage.json {
		expressions = { "app" = "", "msg" = "" }
}

stage.drop {
		source      = "src"
		expression  = ".*test.*"
		older_than  = "24h"
		longer_than = "8KB"
}

stage.drop {
		expression = ".*app1.*"
}

stage.drop {
		source = "app"
		value  = "loki"
}

stage.drop {
		longer_than = "10000B"
}
`

func TestDropStage(t *testing.T) {
	// Enable debug logging
	cfg := &ww.Config{}
	require.Nil(t, cfg.LogLevel.Set("debug"))

	tenBytes, _ := units.ParseBase2Bytes("10B")
	oneHour := 1 * time.Hour

	tests := []struct {
		name       string
		config     *DropConfig
		labels     model.LabelSet
		extracted  map[string]interface{}
		t          time.Time
		entry      string
		shouldDrop bool
	}{
		{
			name: "Longer Than Should Drop",
			config: &DropConfig{
				LongerThan: tenBytes,
			},
			labels:     model.LabelSet{},
			extracted:  map[string]interface{}{},
			entry:      "12345678901",
			shouldDrop: true,
		},
		{
			name: "Longer Than Should Not Drop When Equal",
			config: &DropConfig{
				LongerThan: tenBytes,
			},
			labels:     model.LabelSet{},
			extracted:  map[string]interface{}{},
			entry:      "1234567890",
			shouldDrop: false,
		},
		{
			name: "Longer Than Should Not Drop When Less",
			config: &DropConfig{
				LongerThan: tenBytes,
			},
			labels:     model.LabelSet{},
			extracted:  map[string]interface{}{},
			entry:      "123456789",
			shouldDrop: false,
		},
		{
			name: "Older than Should Drop",
			config: &DropConfig{
				OlderThan: oneHour,
			},
			labels:     model.LabelSet{},
			extracted:  map[string]interface{}{},
			t:          time.Now().Add(-2 * time.Hour),
			shouldDrop: true,
		},
		{
			name: "Older than Should Not Drop",
			config: &DropConfig{
				OlderThan: oneHour,
			},
			labels:     model.LabelSet{},
			extracted:  map[string]interface{}{},
			t:          time.Now().Add(-5 * time.Minute),
			shouldDrop: false,
		},
		{
			name: "Matched Source",
			config: &DropConfig{
				Source: "key",
			},
			labels: model.LabelSet{},
			extracted: map[string]interface{}{
				"key": "",
			},
			shouldDrop: true,
		},
		{
			name: "Did not match Source",
			config: &DropConfig{
				Source: "key1",
			},
			labels: model.LabelSet{},
			extracted: map[string]interface{}{
				"key": "val1",
			},
			shouldDrop: false,
		},
		{
			name: "Matched Source and Value",
			config: &DropConfig{
				Source: "key",
				Value:  "val1",
			},
			labels: model.LabelSet{},
			extracted: map[string]interface{}{
				"key": "val1",
			},
			shouldDrop: true,
		},
		{
			name: "Did not match Source and Value",
			config: &DropConfig{
				Source: "key",
				Value:  "val1",
			},
			labels: model.LabelSet{},
			extracted: map[string]interface{}{
				"key": "VALRUE1",
			},
			shouldDrop: false,
		},
		{
			name: "Regex Matched Source and Value",
			config: &DropConfig{
				Source:     "key",
				Expression: ".*val.*",
			},
			labels: model.LabelSet{},
			extracted: map[string]interface{}{
				"key": "val1",
			},
			shouldDrop: true,
		},
		{
			name: "Regex Did not match Source and Value",
			config: &DropConfig{
				Source:     "key",
				Expression: ".*val.*",
			},
			labels: model.LabelSet{},
			extracted: map[string]interface{}{
				"key": "pal1",
			},
			shouldDrop: false,
		},
		{
			name: "Regex No Matching Source",
			config: &DropConfig{
				Source:     "key",
				Expression: ".*val.*",
			},
			labels: model.LabelSet{},
			extracted: map[string]interface{}{
				"pokey": "pal1",
			},
			shouldDrop: false,
		},
		{
			name: "Regex Did Not Match Line",
			config: &DropConfig{
				Expression: ".*val.*",
			},
			labels:     model.LabelSet{},
			entry:      "this is a line which does not match the regex",
			extracted:  map[string]interface{}{},
			shouldDrop: false,
		},
		{
			name: "Regex Matched Line",
			config: &DropConfig{
				Expression: ".*val.*",
			},
			labels:     model.LabelSet{},
			entry:      "this is a line with the word value in it",
			extracted:  map[string]interface{}{},
			shouldDrop: true,
		},
		{
			name: "Match Source and Length Both Match",
			config: &DropConfig{
				Source:     "key",
				LongerThan: tenBytes,
			},
			labels: model.LabelSet{},
			extracted: map[string]interface{}{
				"key": "pal1",
			},
			entry:      "12345678901",
			shouldDrop: true,
		},
		{
			name: "Match Source and Length Only First Matches",
			config: &DropConfig{
				Source:     "key",
				LongerThan: tenBytes,
			},
			labels: model.LabelSet{},
			extracted: map[string]interface{}{
				"key": "pal1",
			},
			entry:      "123456789",
			shouldDrop: false,
		},
		{
			name: "Match Source and Length Only Second Matches",
			config: &DropConfig{
				Source:     "key",
				LongerThan: tenBytes,
			},
			labels: model.LabelSet{},
			extracted: map[string]interface{}{
				"WOOOOOOOOOOOOOO": "pal1",
			},
			entry:      "123456789012",
			shouldDrop: false,
		},
		{
			name: "Everything Must Match",
			config: &DropConfig{
				Source:     "key",
				Expression: ".*val.*",
				OlderThan:  oneHour,
				LongerThan: tenBytes,
			},
			labels: model.LabelSet{},
			extracted: map[string]interface{}{
				"key": "must contain value to match",
			},
			t:          time.Now().Add(-2 * time.Hour),
			entry:      "12345678901",
			shouldDrop: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateDropConfig(tt.config)
			if err != nil {
				t.Error(err)
			}
			logger := util.TestFlowLogger(t)
			m, err := newDropStage(logger, *tt.config, prometheus.DefaultRegisterer)
			require.NoError(t, err)
			out := processEntries(m, newEntry(tt.extracted, tt.labels, tt.entry, tt.t))
			if tt.shouldDrop {
				assert.Len(t, out, 0)
			} else {
				assert.Len(t, out, 1)
			}
		})
	}
}

func TestDropPipeline(t *testing.T) {
	registry := prometheus.NewRegistry()
	plName := "test_drop_pipeline"
	logger := util.TestFlowLogger(t)
	pl, err := NewPipeline(logger, loadConfig(testDropRiver), &plName, registry)
	require.NoError(t, err)
	out := processEntries(pl,
		newEntry(nil, nil, testMatchLogLineApp1, time.Now()),
		newEntry(nil, nil, testMatchLogLineApp2, time.Now()),
	)

	// Only the second line will go through.
	assert.Len(t, out, 1)
	assert.Equal(t, out[0].Line, testMatchLogLineApp2)
}

var (
	dropVal          = "msg"
	dropRegex        = ".*blah"
	dropInvalidRegex = "(?P<ts[0-9]+).*"
)

func Test_validateDropConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  *DropConfig
		wantErr error
	}{
		{
			name:    "ErrEmpty",
			config:  &DropConfig{},
			wantErr: ErrDropStageEmptyConfig,
		},
		{
			name: "Invalid Config",
			config: &DropConfig{
				Value:      dropVal,
				Expression: dropRegex,
			},
			wantErr: ErrDropStageInvalidConfig,
		},
		{
			name: "Invalid Regex",
			config: &DropConfig{
				Expression: dropInvalidRegex,
			},
			wantErr: fmt.Errorf("%s: %s", ErrDropStageInvalidRegex.Error(), "error parsing regexp: invalid named capture: `(?P<ts[0-9]+).*`"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := validateDropConfig(tt.config); ((err != nil) && (err.Error() != tt.wantErr.Error())) || (err == nil && tt.wantErr != nil) {
				t.Errorf("validateDropConfig() error = %v, wantErr = %v", err, tt.wantErr)
			}
		})
	}
}
