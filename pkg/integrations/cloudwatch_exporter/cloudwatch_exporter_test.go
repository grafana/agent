package cloudwatch_exporter

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCloudwatchExporterConfigInstanceKey(t *testing.T) {
	cfg1 := &Config{
		STSRegion: "us-east-2",
	}
	cfg2 := &Config{
		STSRegion: "us-east-3",
	}

	cfg1Hash, err := cfg1.InstanceKey("")
	require.NoError(t, err)
	cfg2Hash, err := cfg2.InstanceKey("")
	require.NoError(t, err)

	assert.NotEqual(t, cfg1Hash, cfg2Hash)

	// test that making them equal in values leads to the same instance key
	cfg2.STSRegion = "us-east-2"
	cfg2Hash, err = cfg2.InstanceKey("")
	require.NoError(t, err)

	assert.Equal(t, cfg1Hash, cfg2Hash)
}
