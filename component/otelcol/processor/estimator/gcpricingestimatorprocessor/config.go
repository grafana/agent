package gcpricingestimatorprocessor

import (
	"go.opentelemetry.io/collector/component"
)

var (
	_ component.Config = (*Config)(nil)
)

type Config struct{}

// ID implements config.Processor
func (*Config) ID() component.ID {
	panic("unimplemented")
}

// SetIDName implements config.Processor
func (*Config) SetIDName(idName string) {
	panic("unimplemented")
}

// Validate implements config.Processor
func (*Config) Validate() error {
	panic("unimplemented")
}
