package catchpoint

import (
	"testing"

	"github.com/grafana/agent/static/integrations/catchpoint_exporter"
	"github.com/grafana/river"
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
