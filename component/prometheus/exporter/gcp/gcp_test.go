package gcp

import (
	"fmt"
	"strings"
	"testing"

	"github.com/grafana/agent/pkg/integrations/gcp_exporter"
	"github.com/grafana/agent/pkg/river"
	"github.com/stretchr/testify/require"
)

func TestUnmarshalRiver(t *testing.T) {
	expected := DefaultArguments
	expected.ProjectIDs = []string{
		"foo",
		"bar",
	}
	expected.MetricPrefixes = []string{
		"pubsub.googleapis.com/snapshot",
		"pubsub.googleapis.com/subscription/num_undelivered_messages",
		"pubsub.googleapis.com/subscription/oldest_unacked_message_age",
	}

	riverCfg := fmt.Sprintf(`
		project_ids = [%s]
		metrics_prefixes = [%s]
`,
		"\""+strings.Join(expected.ProjectIDs, "\", \"")+"\"",
		"\""+strings.Join(expected.MetricPrefixes, "\", \"")+"\"",
	)

	var args Arguments
	err := river.Unmarshal([]byte(riverCfg), &args)
	require.NoError(t, err)

	require.Equal(t, expected, args)
}

func TestConvertConfig(t *testing.T) {
	args := DefaultArguments

	res := args.Convert()
	require.NotNil(t, res)
	require.Equal(t, gcp_exporter.DefaultConfig, *res)

}
