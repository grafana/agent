package featuregate

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCheckAllowed(t *testing.T) {
	type args struct {
		stability    Stability
		minStability Stability
		featureName  string
	}
	tests := []struct {
		name        string
		args        args
		errContains string
	}{
		{
			name: "undefined stability",
			args: args{
				stability:    StabilityUndefined,
				minStability: StabilityStable,
				featureName:  "component do.all.things",
			},
			errContains: "stability levels must be defined: got <invalid_stability_level> as stability of component do.all.things",
		},
		{
			name: "too low stability",
			args: args{
				stability:    StabilityBeta,
				minStability: StabilityStable,
				featureName:  "component do.all.things",
			},
			errContains: "component do.all.things is at stability level \"beta\", which is below the minimum allowed stability level \"stable\"",
		},
		{
			name: "equal stability",
			args: args{
				stability:    StabilityBeta,
				minStability: StabilityBeta,
				featureName:  "component do.all.things",
			},
		},
		{
			name: "higher stability",
			args: args{
				stability:    StabilityStable,
				minStability: StabilityBeta,
				featureName:  "component do.all.things",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := CheckAllowed(tt.args.stability, tt.args.minStability, tt.args.featureName)
			if tt.errContains == "" {
				require.NoError(t, err)
			} else {
				require.Contains(t, err.Error(), tt.errContains)
			}
		})
	}
}
