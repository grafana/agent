package featuregate

import (
	"github.com/grafana/agent/pkg/util"
	_ "go.opentelemetry.io/collector/obsreport"
)

// An init function just for Flow mode.
func init() {
	err := util.SetupFlowModeOtelFeatureGates()
	if err != nil {
		panic(err)
	}
}
