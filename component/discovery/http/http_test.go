package http

import (
	"net/url"
	"testing"
	"time"

	"github.com/grafana/agent/component/common/config"
	"github.com/grafana/agent/pkg/river"
	"github.com/prometheus/common/model"
	"github.com/stretchr/testify/require"
	"gotest.tools/assert"
)

func TestRiverConfig(t *testing.T) {
	var exampleRiverConfig = `
	url = "https://www.example.com:12345/foo"
	refresh_interval = "14s"
	basic_auth {
		username = "123"
		password = "456"
	}
`
	var args Arguments
	err := river.Unmarshal([]byte(exampleRiverConfig), &args)
	require.NoError(t, err)
	assert.Equal(t, args.HTTPClientConfig.BasicAuth.Username, "123")
}

func TestConvert(t *testing.T) {
	args := DefaultArguments
	u, err := url.Parse("https://www.example.com:12345/foo")
	require.NoError(t, err)
	args.URL = config.URL{URL: u}

	sd := args.Convert()
	assert.Equal(t, "https://www.example.com:12345/foo", sd.URL)
	assert.Equal(t, model.Duration(60*time.Second), sd.RefreshInterval)
	assert.Equal(t, true, sd.HTTPClientConfig.EnableHTTP2)
}
