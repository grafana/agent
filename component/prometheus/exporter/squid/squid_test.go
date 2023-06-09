package squid

import (
	"testing"

	"github.com/grafana/agent/component/discovery"
	"github.com/grafana/agent/pkg/integrations/squid_exporter"
	"github.com/grafana/agent/pkg/river"
	"github.com/grafana/agent/pkg/river/rivertypes"
	"github.com/prometheus/common/config"
	"github.com/stretchr/testify/require"
)

func TestRiverUnmarshal(t *testing.T) {
	riverConfig := `
	address = "some_address"
	username     = "some_user"
	password     = "some_password"
	`

	var args Arguments
	err := river.Unmarshal([]byte(riverConfig), &args)
	require.NoError(t, err)

	expected := Arguments{
		SquidAddr:     "some_address",
		SquidUser:     "some_user",
		SquidPassword: rivertypes.Secret("some_password"),
	}

	require.Equal(t, expected, args)
}

func TestConvert(t *testing.T) {
	riverConfig := `
	address = "some_address"
	username     = "some_user"
	password     = "some_password"
	`
	var args Arguments
	err := river.Unmarshal([]byte(riverConfig), &args)
	require.NoError(t, err)

	res := args.Convert()

	expected := squid_exporter.Config{
		Address:  "some_address",
		Username: "some_user",
		Password: config.Secret("some_password"),
	}
	require.Equal(t, expected, *res)
}

func TestCustomizeTarget(t *testing.T) {
	args := Arguments{
		SquidAddr: "some_address",
	}

	baseTarget := discovery.Target{}
	newTargets := customizeTarget(baseTarget, args)
	require.Equal(t, 1, len(newTargets))
	require.Equal(t, "some_address", newTargets[0]["instance"])
}
