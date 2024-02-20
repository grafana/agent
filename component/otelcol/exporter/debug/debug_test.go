package debug_test

import (
	"fmt"
	"testing"

	"github.com/grafana/agent/component/otelcol/exporter/debug"
	"github.com/grafana/river"
	"github.com/stretchr/testify/require"
	otelcomponent "go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config/configtelemetry"
	debugexporter "go.opentelemetry.io/collector/exporter/debugexporter"
)

func Test(t *testing.T) {
	tests := []struct {
		testName       string
		args           string
		expectedReturn debugexporter.Config
		errorMsg       string
	}{
		{
			testName: "defaultConfig",
			args: `
			verbosity = "basic"
			sampling_initial = 2
			sampling_thereafter = 500
			`,
			expectedReturn: debugexporter.Config{
				Verbosity:          configtelemetry.LevelBasic,
				SamplingInitial:    2,
				SamplingThereafter: 500,
			},
		},

		{
			testName: "validConfig",
			args: ` 
				verbosity = "detailed"
				sampling_initial = 5
				sampling_thereafter = 20
			`,
			expectedReturn: debugexporter.Config{
				Verbosity:          configtelemetry.LevelDetailed,
				SamplingInitial:    5,
				SamplingThereafter: 20,
			},
		},

		{
			testName: "invalidConfig",
			args: `
				verbosity = "test"
				sampling_initial = 5
				sampling_thereafter = 20
			`,
			errorMsg: "error in conversion to config arguments",
		},
	}

	for _, tc := range tests {
		t.Run(tc.testName, func(t *testing.T) {
			var args debug.Arguments
			err := river.Unmarshal([]byte(tc.args), &args)
			require.NoError(t, err)

			actualPtr, err := args.Convert()
			if tc.errorMsg != "" {
				require.ErrorContains(t, err, tc.errorMsg)
				return
			}

			require.NoError(t, err)

			actual := actualPtr.(*debugexporter.Config)
			fmt.Printf("Passed conversion")

			require.NoError(t, otelcomponent.ValidateConfig(actual))

			require.Equal(t, tc.expectedReturn, *actual)
		})
	}
}
