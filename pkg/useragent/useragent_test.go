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
		Operator bool
		GOOS     string
	}{
		{
			Name:     "basic",
			Mode:     "",
			Expected: "GrafanaAgent/v1.2.3 (static;linux)",
			GOOS:     "linux",
		},
		{
			Name:     "flow",
			Mode:     "flow",
			Expected: "GrafanaAgent/v1.2.3 (flow;windows)",
			GOOS:     "windows",
		},
		{
			Name:     "static",
			Mode:     "static",
			Expected: "GrafanaAgent/v1.2.3 (static;darwin)",
			GOOS:     "darwin",
		},
		{
			Name: "unknown",
			Mode: "blahlahblah",
			// unknown mode, should not happen. But we will substitute 'unknown' to avoid allowing arbitrary cardinality.
			Expected: "GrafanaAgent/v1.2.3 (unknown;freebsd)",
			GOOS:     "freebsd",
		},
		{
			Name:     "operator",
			Mode:     "static",
			Operator: true,
			Expected: "GrafanaAgent/v1.2.3 (static;linux;operator)",
			GOOS:     "linux",
		},
	}
	for _, tst := range tests {
		t.Run(tst.Name, func(t *testing.T) {
			goos = tst.GOOS
			if tst.Operator {
				t.Setenv(operatorEnv, "1")
			} else {
				t.Setenv(operatorEnv, "")
			}
			t.Setenv(modeEnv, tst.Mode)
			actual := UserAgent()
			require.Equal(t, tst.Expected, actual)
		})
	}
}
