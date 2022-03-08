package magic

import (
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestNodeMapper(t *testing.T) {
	testYaml := `
person:
  name: bob
  age: 16
`
	node := &yaml.Node{}
	err := yaml.Unmarshal([]byte(testYaml), node)
	require.NoError(t, err)
	nm := make([]*NodeMapping, 0)
	nm = buildNodeMap(nil, node.Content[0].Content[0], node.Content[0].Content[1], nm)
	require.NoError(t, err)
}
