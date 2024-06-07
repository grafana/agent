package catchpoint

import (
	"testing"

	"github.com/grafana/agent/static/integrations/catchpoint_exporter"
	"github.com/grafana/river"
	"github.com/stretchr/testify/require"
)

func TestRiverUnmarshal(t *testing.T) {
	riverConfig := `
  port               = "3030"
  verbose_logging     = true
  webhook_path        = "/nondefault-webhook-path"
	`

	var args Arguments
	err := river.Unmarshal([]byte(riverConfig), &args)
	require.NoError(t, err)

	expected := Arguments{
		VerboseLogging: true,
		Port:           "3030",
		WebhookPath:    "/nondefault-webhook-path",
	}

	require.Equal(t, expected, args)
}

func TestConvert(t *testing.T) {
	riverConfig := `
  port               = "3030"
  verbose_logging     = true
  webhook_path        = "/nondefault-webhook-path"
	`

	var args Arguments
	err := river.Unmarshal([]byte(riverConfig), &args)
	require.NoError(t, err)

	res := args.Convert()

	expected := catchpoint_exporter.Config{
		VerboseLogging: true,
		Port:           "3030",
		WebhookPath:    "/nondefault-webhook-path",
	}
	require.Equal(t, expected, *res)
}
