package catchpoint

import (
	"testing"

	"github.com/grafana/agent/pkg/integrations/catchpoint_exporter"
	"github.com/grafana/agent/pkg/river"
	"github.com/grafana/agent/pkg/river/rivertypes"
	config_util "github.com/prometheus/common/config"
	"github.com/stretchr/testify/require"
)

func TestRiverUnmarshal(t *testing.T) {
	riverConfig := `
	port        = "3030"
  verbose     = true
  webhookpath = "/nondefault-webhook-path"
	`

	var args Arguments
	err := river.Unmarshal([]byte(riverConfig), &args)
	require.NoError(t, err)

	expected := Arguments{
		Verbose:     true,
		Port:        "3030",
		WebhookPath: "/nondefault-webhook-path",
	}

	require.Equal(t, expected, args)
}

func TestConvert(t *testing.T) {
	riverConfig := `
	port        = "3030"
  verbose     = true
  webhookpath = "/nondefault-webhook-path"
	`

	var args Arguments
	err := river.Unmarshal([]byte(riverConfig), &args)
	require.NoError(t, err)

	res := args.Convert()

	expected := catchpoint_exporter.Config{
		Verbose:     true,
		Port:        "3030",
		WebhookPath: "/nondefault-webhook-path",
	}
	require.Equal(t, expected, *res)
}
