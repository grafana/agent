// Package featuregate automatically enables upstream feature gates that
// otelcol components require. The package must be imported for the feature
// gates to be enabled.
package featuregate

import (
	"os"

	"go.opentelemetry.io/collector/featuregate"

	_ "go.opentelemetry.io/collector/obsreport" // telemetry.useOtelForInternalMetrics
)

// TODO(rfratto): this package should be updated occasionally to remove feature
// gates which no longer exist.
//
// Once all feature gates are removed, this package can be removed as well.

func init() {
	_ = enableFeatureGates(featuregate.GetRegistry())
}

func enableFeatureGates(reg *featuregate.Registry) error {
	// TODO(marctc): temporary workaround to fix issue with traces' metrics not
	// being collected even flow is not enabled.
	return reg.Apply(map[string]bool{
		"telemetry.useOtelForInternalMetrics": isFlowRunning(),
	})
}

func isFlowRunning() bool {
	key, found := os.LookupEnv("AGENT_MODE")
	if !found {
		key, found := os.LookupEnv("EXPERIMENTAL_ENABLE_FLOW")
		if !found {
			return false
		}
		return key == "true" || key == "1"
	}

	switch key {
	case "flow":
		return true
	}
	return false
}
