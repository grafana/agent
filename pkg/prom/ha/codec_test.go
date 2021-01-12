package ha

import (
	"testing"

	"github.com/grafana/agent/pkg/prom/instance"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
)

func TestCodec(t *testing.T) {
	exampleConfig := `name: 'test'
host_filter: false
scrape_configs:
  - job_name: process-1
    static_configs:
      - targets: ['process-1:80']
        labels:
          cluster: 'local'
          origin: 'agent'`

	var in instance.Config
	err := yaml.Unmarshal([]byte(exampleConfig), &in)
	require.NoError(t, err)

	c := &yamlCodec{}
	bb, err := c.Encode(in)
	require.NoError(t, err)

	out, err := c.Decode(bb)
	require.NoError(t, err)
	require.Equal(t, &in, out)
}

func TestCodec_RetainsSecrets_Password(t *testing.T) {
	exampleConfig := `name: 'test'
host_filter: false
scrape_configs:
  - job_name: process-1
    static_configs:
      - targets: ['process-1:80']
        labels:
          cluster: 'local'
          origin: 'agent'
remote_write:
  - url: http://cortex:9090/api/prom/push
    basic_auth:
      username: test_username
      password: test_pass`

	var in instance.Config
	err := yaml.Unmarshal([]byte(exampleConfig), &in)
	require.NoError(t, err)

	c := &yamlCodec{}
	bb, err := c.Encode(in)
	require.NoError(t, err)

	out, err := c.Decode(bb)
	require.NoError(t, err)
	require.Equal(t, &in, out)
}

func TestCodec_RetainsSecrets_BearerToken(t *testing.T) {
	exampleConfig := `name: 'test'
host_filter: false
scrape_configs:
  - job_name: process-1
    static_configs:
      - targets: ['process-1:80']
        labels:
          cluster: 'local'
          origin: 'agent'
remote_write:
  - url: http://cortex:9090/api/prom/push
    bearer_token: test_bearer`

	var in instance.Config
	err := yaml.Unmarshal([]byte(exampleConfig), &in)
	require.NoError(t, err)

	c := &yamlCodec{}
	bb, err := c.Encode(in)
	require.NoError(t, err)

	out, err := c.Decode(bb)
	require.NoError(t, err)
	require.Equal(t, &in, out)
}

// TestCodec_Decode_Nil makes sure that if Decode is called with an empty value,
// which may happen when a key is deleted, that no error occurs and instead an
// nil value is returned.
func TestCodec_Decode_Nil(t *testing.T) {
	c := &yamlCodec{}

	input := [][]byte{nil, make([]byte, 0)}
	for _, bb := range input {
		out, err := c.Decode(bb)
		require.Nil(t, err)
		require.Nil(t, out)
	}
}
