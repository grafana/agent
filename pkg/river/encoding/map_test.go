package encoding

import (
	"testing"

	"github.com/grafana/agent/pkg/river/internal/value"
	"github.com/stretchr/testify/require"
)

func TestMap(t *testing.T) {
	testMap := make(map[string]string)
	testMap["testBlank"] = ""
	testMap["testValue"] = "value"
	mf, err := newRiverMap(value.Encode(testMap))
	require.NoError(t, err)
	require.True(t, mf.hasValue())
}
