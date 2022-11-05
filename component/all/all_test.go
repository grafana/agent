package all

import (
	"testing"

	"github.com/grafana/agent/component/discovery"

	"github.com/stretchr/testify/require"

	"github.com/grafana/agent/component/local/file"
	"github.com/grafana/agent/pkg/river"
)

func TestComponent(t *testing.T) {
	md := river.MetadataDict{Types: make([]river.DataType, 0)}
	c, err := md.GenerateComponent("local.file", false, file.Arguments{}, file.Exports{})
	require.NoError(t, err)
	require.NotNil(t, c)

	c1, err := md.GenerateComponent("discovery.kubernetes", false, nil, discovery.Exports{})
	require.NoError(t, err)
	require.NotNil(t, c1)
	require.True(t, c1.ArgumentField == "")

}
