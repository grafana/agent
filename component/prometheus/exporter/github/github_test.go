package github

import (
	"testing"

	"github.com/grafana/agent/pkg/river"
	"github.com/stretchr/testify/require"
)

func TestUnmarshalRiver(t *testing.T) {
	riverCfg := `
		api_token_file = "/etc/github-api-token"
		repositories = ["grafana/agent"]
		organizations = ["grafana", "prometheus"]
		users = ["jcreixell"]
		api_url = "https://some-other-api.github.com"
`
	var cfg Config
	err := river.Unmarshal([]byte(riverCfg), &cfg)
	require.NoError(t, err)
	require.Equal(t, "/etc/github-api-token", cfg.APITokenFile)
	require.Equal(t, []string{"grafana/agent"}, cfg.Repositories)
	require.Contains(t, cfg.Organizations, "grafana")
	require.Contains(t, cfg.Organizations, "prometheus")
	require.Equal(t, []string{"jcreixell"}, cfg.Users)
	require.Equal(t, "https://some-other-api.github.com", cfg.APIURL)
}

func TestConvert(t *testing.T) {
	cfg := Config{
		APITokenFile:  "/etc/github-api-token",
		Repositories:  []string{"grafana/agent"},
		Organizations: []string{"grafana", "prometheus"},
		Users:         []string{"jcreixell"},
		APIURL:        "https://some-other-api.github.com",
	}

	res := cfg.Convert()
	require.Equal(t, "/etc/github-api-token", res.APITokenFile)
	require.Equal(t, []string{"grafana/agent"}, res.Repositories)
	require.Contains(t, res.Organizations, "grafana")
	require.Contains(t, res.Organizations, "prometheus")
	require.Equal(t, []string{"jcreixell"}, res.Users)
	require.Equal(t, "https://some-other-api.github.com", res.APIURL)
}
