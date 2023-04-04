package memcached

import (
	"testing"
	"time"

	"github.com/grafana/agent/pkg/integrations/memcached_exporter"
	"github.com/grafana/agent/pkg/river"
	"github.com/stretchr/testify/assert"
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
