package memcached

import (
	"testing"
	"time"

	"github.com/grafana/agent/component/discovery"
	"github.com/grafana/agent/pkg/integrations/memcached_exporter"
	"github.com/grafana/agent/pkg/river"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRiverUnmarshal(t *testing.T) {
	var exampleRiverConfig = `
address = "localhost:99"
timeout = "5s"`

	var args Arguments
	err := river.Unmarshal([]byte(exampleRiverConfig), &args)
	assert.NoError(t, err)

	expected := Arguments{
		Address: "localhost:99",
		Timeout: 5 * time.Second,
	}

	assert.Equal(t, expected, args)
}

func TestRiverUnmarshalDefaults(t *testing.T) {
	var exampleRiverConfig = ``

	var args Arguments
	err := river.Unmarshal([]byte(exampleRiverConfig), &args)
	assert.NoError(t, err)

	expected := DefaultArguments

	assert.Equal(t, expected, args)
}

func TestRiverConvert(t *testing.T) {
	riverArguments := Arguments{
		Address: "localhost:99",
		Timeout: 5 * time.Second,
	}

	expected := &memcached_exporter.Config{
		MemcachedAddress: "localhost:99",
		Timeout:          5 * time.Second,
	}

	assert.Equal(t, expected, riverArguments.Convert())
}

func TestCustomizeTarget(t *testing.T) {
	args := Arguments{
		Address: "localhost:99",
	}

	baseTarget := discovery.Target{}
	newTargets := customizeTarget(baseTarget, args)
	require.Equal(t, 1, len(newTargets))
	require.Equal(t, "localhost:99", newTargets[0]["instance"])
}
