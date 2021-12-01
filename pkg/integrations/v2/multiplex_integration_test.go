package integrations

import (
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestMultiplexIntegration_YAML(t *testing.T) {
	mux := NewMultiplexConfig("example_configs", &exampleConfig{})

	input := `
- name: John
  age: 35
- name: Anne
  age: 34`

	err := yaml.Unmarshal([]byte(input), mux)
	require.NoError(t, err)

	expect := `
- name: John
  age: 35
  alive: true
- name: Anne
  age: 34
  alive: true`

	bb, err := yaml.Marshal(mux)
	require.NoError(t, err)
	require.YAMLEq(t, expect, string(bb))

	require.Equal(t, &exampleConfig{"John", 35, true}, mux.(*multiplexConfig).configs[0].(*exampleConfig))
	require.Equal(t, &exampleConfig{"Anne", 34, true}, mux.(*multiplexConfig).configs[1].(*exampleConfig))
}

type exampleConfig struct {
	FirstName string `yaml:"name"`
	Age       int    `yaml:"age"`
	Alive     bool   `yaml:"alive"`
}

func (c *exampleConfig) UnmarshalYAML(unmarshal func(interface{}) error) error {
	c.Alive = true

	type config exampleConfig
	return unmarshal((*config)(c))
}

func (c *exampleConfig) Name() string { return "example" }

func (c *exampleConfig) Identifier(IntegrationOptions) (string, error) { return c.FirstName, nil }

func (c *exampleConfig) NewIntegration(IntegrationOptions) (Integration, error) {
	return NoOpIntegration, nil
}
