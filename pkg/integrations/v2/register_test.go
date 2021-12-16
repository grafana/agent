package integrations

import (
	"testing"
	"time"

	"github.com/go-kit/log"
	v1 "github.com/grafana/agent/pkg/integrations"
	agent_v1 "github.com/grafana/agent/pkg/integrations/agent"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
)

func TestIntegrationRegistration(t *testing.T) {
	setRegistered(t, map[Config]Type{
		&testIntegrationA{}: TypeEither,
		&testIntegrationB{}: TypeEither,
	})

	// This test checks for a few things:
	//
	// 1. Registered integrations will be parseable
	// 2. Registered integrations that are not present will not be unmarshaled to
	//    the list of configs
	// 3. Registered integrations that have defaults may still be parsed
	// 4. Strict parsing should still work as expected.

	var cfgToParse = `
name: John Doe
duration: 500ms
test:
  text: Hello, world!
`

	var fullCfg testFullConfig
	err := yaml.UnmarshalStrict([]byte(cfgToParse), &fullCfg)
	require.NoError(t, err)

	expect := testFullConfig{
		Name:     "John Doe",
		Duration: 500 * time.Millisecond,
		Default:  12345,
		Configs: []Config{
			&testIntegrationA{Text: "Hello, world!", Truth: true},
		},
	}
	require.Equal(t, expect, fullCfg)
}

func TestIntegrationRegistration_Multiple(t *testing.T) {
	setRegistered(t, map[Config]Type{
		&testIntegrationA{}: TypeEither,
		&testIntegrationB{}: TypeEither,
	})

	var cfgToParse = `
name: John Doe
duration: 500ms
test_configs:
  - text: Hello, world!
  - text: Hello again!`

	var fullCfg testFullConfig
	err := yaml.UnmarshalStrict([]byte(cfgToParse), &fullCfg)
	require.NoError(t, err)

	expect := testFullConfig{
		Name:     "John Doe",
		Duration: 500 * time.Millisecond,
		Default:  12345,
		Configs: []Config{
			&testIntegrationA{Text: "Hello, world!", Truth: true},
			&testIntegrationA{Text: "Hello again!", Truth: true},
		},
	}
	require.Equal(t, expect, fullCfg)
}

func TestIntegrationRegistration_Legacy(t *testing.T) {
	setRegistered(t, nil)

	RegisterLegacy(&agent_v1.Config{}, TypeSingleton, func(in v1.Config) UpgradedConfig {
		return &legacyShim{Data: in}
	})

	var cfgToParse = `
name: John Doe
duration: 500ms
agent:
  instance: foo`

	var fullCfg testFullConfig
	err := yaml.UnmarshalStrict([]byte(cfgToParse), &fullCfg)
	require.NoError(t, err)

	require.Len(t, fullCfg.Configs, 1)
	require.IsType(t, &legacyShim{}, fullCfg.Configs[0])

	shim := fullCfg.Configs[0].(*legacyShim)
	require.IsType(t, &agent_v1.Config{}, shim.Data)

	agentConfig := shim.Data.(*agent_v1.Config)
	require.Equal(t, "foo", *agentConfig.Common.InstanceKey)
	require.Equal(t, true, agentConfig.Enabled)
}

type legacyShim struct{ Data v1.Config }

func (s *legacyShim) LegacyConfig() v1.Config              { return s.Data }
func (s *legacyShim) Name() string                         { return s.Data.Name() }
func (s *legacyShim) Identifier(g Globals) (string, error) { return g.AgentIdentifier, nil }
func (s *legacyShim) NewIntegration(log.Logger, Globals) (Integration, error) {
	return NoOpIntegration, nil
}

type testIntegrationA struct {
	Text  string `yaml:"text"`
	Truth bool   `yaml:"truth"`
}

func (i *testIntegrationA) Name() string                       { return "test" }
func (i *testIntegrationA) Identifier(Globals) (string, error) { return "integrationA", nil }
func (i *testIntegrationA) NewIntegration(log.Logger, Globals) (Integration, error) {
	return NoOpIntegration, nil
}

func (i *testIntegrationA) UnmarshalYAML(unmarshal func(interface{}) error) error {
	i.Truth = true
	type plain testIntegrationA
	return unmarshal((*plain)(i))
}

type testIntegrationB struct {
	Text string `yaml:"text"`
}

func (*testIntegrationB) Name() string                       { return "shouldnotbefound" }
func (*testIntegrationB) Identifier(Globals) (string, error) { return "integrationB", nil }
func (*testIntegrationB) NewIntegration(log.Logger, Globals) (Integration, error) {
	return NoOpIntegration, nil
}

type testFullConfig struct {
	// Some random fields that will also be exposed
	Name     string        `yaml:"name"`
	Duration time.Duration `yaml:"duration"`
	Default  int           `yaml:"default"`

	Configs Configs `yaml:"-"`
}

func (c *testFullConfig) UnmarshalYAML(unmarshal func(interface{}) error) error {
	// This default value should not change.
	c.Default = 12345
	return UnmarshalYAML(c, unmarshal)
}
