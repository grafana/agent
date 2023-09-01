package util

import (
	"fmt"

	"go.opentelemetry.io/collector/featuregate"
)

func EnableOtelFeatureGates(fgNames ...string) error {
	fgReg := featuregate.GlobalRegistry()

	for _, fg := range fgNames {
		err := fgReg.Set(fg, true)
		if err != nil {
			return fmt.Errorf("error setting Otel feature gate: %w", err)
		}
	}

	return nil
}
