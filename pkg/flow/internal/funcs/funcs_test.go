package funcs_test

import (
	"testing"

	"github.com/grafana/agent/pkg/flow/internal/funcs"
	"github.com/stretchr/testify/require"
	"github.com/zclconf/go-cty/cty"
)

func TestEnvFunc(t *testing.T) {
	t.Setenv("TEST_VAR", "HELLO_WORLD")

	res, err := funcs.EnvFunc.Call([]cty.Value{cty.StringVal("TEST_VAR")})
	require.NoError(t, err)
	require.Equal(t, "HELLO_WORLD", res.AsString())
}
