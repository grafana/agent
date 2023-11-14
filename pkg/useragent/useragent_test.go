package useragent

import (
	"testing"

	"github.com/grafana/agent/pkg/build"
	"github.com/stretchr/testify/require"
)

func TestUserAgent(t *testing.T) {
	build.Version = "v1.2.3"
	tests := []struct {
		Name     string
		Mode     string
		Expected string
	}{
		{
			Name:     "basic",
			Mode:     "",
			Expected: "GrafanaAgent/v1.2.3(static)",
		},
		{
			Name:     "flow",
			Mode:     "flow",
			Expected: "GrafanaAgent/v1.2.3(flow)",
		},
		{
			Name:     "static",
			Mode:     "static",
			Expected: "GrafanaAgent/v1.2.3(static)",
		},
		{
			Name: "unknown",
			Mode: "blahlahblah",
			// unknown mode, just leave it out of user-agent. Don't want arbitrary values to get sent here.
			Expected: "GrafanaAgent/v1.2.3",
		},
	}
	for _, tst := range tests {
		t.Run(tst.Name, func(t *testing.T) {
			t.Setenv("AGENT_MODE", tst.Mode)
			actual := UserAgent()
			require.Equal(t, tst.Expected, actual)
		})
	}
}
