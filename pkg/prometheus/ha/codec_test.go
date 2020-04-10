package ha

import (
	"testing"

	"github.com/grafana/agent/pkg/prometheus/instance"
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
