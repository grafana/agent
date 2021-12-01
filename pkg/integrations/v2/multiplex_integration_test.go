package integrations

import (
	"testing"

	"github.com/grafana/agent/pkg/util"
	"github.com/stretchr/testify/require"
)

func TestMultiplexIntegration_getControllerConfig(t *testing.T) {
	cfg := muxExample{{"John"}, {"Anne"}}

	i, err := cfg.NewIntegration(IntegrationOptions{Logger: util.TestLogger(t)})
	require.NoError(t, err)

	mux := i.(*multiplexIntegration)
	require.Equal(t, &exampleConfig{"John"}, mux.cfg[0].(*exampleConfig))
	require.Equal(t, &exampleConfig{"Anne"}, mux.cfg[1].(*exampleConfig))
}

type exampleConfig struct {
	FirstName string `yaml:"name"`
}

func (c *exampleConfig) Name() string                                  { return "example" }
func (c *exampleConfig) Identifier(IntegrationOptions) (string, error) { return c.FirstName, nil }
func (c *exampleConfig) NewIntegration(IntegrationOptions) (Integration, error) {
	return NoOpIntegration, nil
}

type muxExample []*exampleConfig

func (m muxExample) Name() string                                  { return "example_configs" }
func (m muxExample) Identifier(IntegrationOptions) (string, error) { return "example_configs", nil }
func (m muxExample) Multiplexed()                                  {}
func (m muxExample) NewIntegration(opts IntegrationOptions) (Integration, error) {
	return NewMultiplexIntegration(m, opts)
}
