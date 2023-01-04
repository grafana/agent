package cluster

import (
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestAgent_OmmitEmptyFields(t *testing.T) {
	var cfg Config
	yml, err := yaml.Marshal(&cfg)
	require.NoError(t, err)
	require.Equal(t, "{}\n", string(yml))
}
