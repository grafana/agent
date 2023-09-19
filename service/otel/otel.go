// Package otel implements the otel service for Flow.
// This service registers feature gates will be used by the otelcol components
// based on upstream Collector components.
package otel

import (
	"context"
	"fmt"

	"github.com/grafana/agent/pkg/util"
	"github.com/grafana/agent/service"
)

// ServiceName defines the name used for the otel service.
const ServiceName = "otel"

type Service struct{}

var _ service.Service = (*Service)(nil)

func New() *Service {
	return &Service{}
}

// Data implements service.Service. It returns nil, as the otel service does
// not have any runtime data.
func (*Service) Data() any {
	return nil
}

// Definition implements service.Service.
func (*Service) Definition() service.Definition {
	return service.Definition{
		Name:       ServiceName,
		ConfigType: nil, // otel does not accept configuration
		DependsOn:  []string{},
	}
}

// Run implements service.Service.
func (*Service) Run(ctx context.Context, host service.Host) error {
	err := util.SetupFlowModeOtelFeatureGates()
	if err != nil {
		return err
	}

	<-ctx.Done()
	return nil
}

// Update implements service.Service.
func (*Service) Update(newConfig any) error {
	return fmt.Errorf("otel service does not support configuration")
}
