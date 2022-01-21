package v2

import (
	"context"
	"net/http"

	"github.com/grafana/agent/pkg/integrations/shared"

	"github.com/grafana/agent/pkg/integrations/v2/common"

	"github.com/grafana/agent/pkg/util"
)

// funcIntegration is a function that implements Integration.
type funcIntegration func(ctx context.Context) error

func (fi funcIntegration) MetricsHandler() (http.Handler, error) {
	panic("implement me")
}

func (fi funcIntegration) ScrapeConfigs() []shared.ScrapeConfig {
	panic("implement me")
}

func (fi funcIntegration) Run(ctx context.Context) error {
	return fi(ctx)
}

// RunIntegration implements Integration.
func (fi funcIntegration) RunIntegration(ctx context.Context) error { return fi(ctx) }

// NoOpIntegration is an Integration that does nothing.
var NoOpIntegration = funcIntegration(func(ctx context.Context) error {
	<-ctx.Done()
	return nil
})

// CompareConfigs will return true if a and b are equal. If neither a or b
// implement ComparableConfig, then configs are compared by marshaling to YAML
// and comparing the results.
func CompareConfigs(a, b Config) bool {
	if a, ok := a.(ComparableConfig); ok {
		return a.ConfigEquals(b)
	}
	if b, ok := b.(ComparableConfig); ok {
		return b.ConfigEquals(a)
	}
	return util.CompareYAML(a, b)
}

type IntegrationConfig interface {
	Cfg() Config
	Common() common.MetricsConfig
}

// IntegrationConfigs is an interface that allows access to any actively running configs
type IntegrationConfigs interface {
	ActiveConfigs() []Config
}
