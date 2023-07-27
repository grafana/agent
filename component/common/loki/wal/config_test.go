package wal

import (
	"testing"
	"time"

	"github.com/grafana/agent/pkg/river"
	"github.com/stretchr/testify/require"
)

func TestWALConfigUnmarshalledThroughRiver(t *testing.T) {
	type testcase struct {
		raw           string
		errorExpected bool
		expected      Config
	}

	for name, tc := range map[string]testcase{
		"default config is wal disabled": {
			raw: "",
			expected: Config{
				Enabled:       false,
				MaxSegmentAge: defaultMaxSegmentAge,
				WatchConfig:   DefaultWatchConfig,
			},
		},
		"wal enabled with defaults": {
			raw: `
			enabled = true
			`,
			expected: Config{
				Enabled:       true,
				MaxSegmentAge: defaultMaxSegmentAge,
				WatchConfig:   DefaultWatchConfig,
			},
		},
		"wal enabled with some overrides": {
			raw: `
			enabled = true
			max_segment_age = "10m"
			min_read_frequency = "11m"
			`,
			expected: Config{
				Enabled:       true,
				MaxSegmentAge: time.Minute * 10,
				WatchConfig: WatchConfig{
					MinReadFrequency: time.Minute * 11,
					MaxReadFrequency: DefaultWatchConfig.MaxReadFrequency,
				},
			},
		},
	} {
		t.Run(name, func(t *testing.T) {
			cfg := Config{}
			err := river.Unmarshal([]byte(tc.raw), &cfg)
			if tc.errorExpected {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tc.expected, cfg)
		})
	}
}
