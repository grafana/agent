package featuregate

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/featuregate"
)

func Test_enableFeatureGates(t *testing.T) {
	err := enableFeatureGates(featuregate.GetRegistry())
	require.NoError(t, err, "enableFeatureGates should not have a failed. Did a feature gate get removed from upstream?")
}
