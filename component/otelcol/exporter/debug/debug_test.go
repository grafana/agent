package debug_test

import (
	"testing"

	"github.com/grafana/agent/component/otelcol/exporter/debug"
	"github.com/stretchr/testify/require"
	otelcomponent "go.opentelemetry.io/collector/component"
)

func Test(t *testing.T) {
	happyArgs := debug.Arguments{
		Verbosity:          "detailed",
		SamplingInitial:    5,
		SamplingThereafter: 20,
	}
	// Check no errors on converting to exporter args
	otelconf, err := happyArgs.Convert()
	require.NoError(t, err)

	// Check that exporter config is created correctly
	err = otelcomponent.ValidateConfig(otelconf)
	require.NoError(t, err, "error on creating debug exporter config")

	invalidArgs := debug.Arguments{
		Verbosity:          "test",
		SamplingInitial:    5,
		SamplingThereafter: 20,
	}
	// Check error on converting invalid args
	_, err = invalidArgs.Convert()
	require.NotNil(t, err, "no error on invalid arguments")
}
