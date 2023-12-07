package eureka

import (
	"testing"
	"time"

	"github.com/grafana/agent/component/common/config"
	"github.com/grafana/river"
	"github.com/grafana/river/rivertypes"
	promcfg "github.com/prometheus/common/config"
	"github.com/prometheus/common/model"
	prom_sd "github.com/prometheus/prometheus/discovery/eureka"
	"github.com/stretchr/testify/require"
)

func TestUnmarshal(t *testing.T) {
	exampleCfg := `
	server = "http://localhost:8080/eureka/v1"
	refresh_interval = "10s"
	basic_auth {
		username = "exampleuser"
		password = "examplepassword"
	}`
	var args Arguments
	err := river.Unmarshal([]byte(exampleCfg), &args)
	require.NoError(t, err)

	require.Equal(t, "http://localhost:8080/eureka/v1", args.Server)
	require.Equal(t, "10s", args.RefreshInterval.String())
	require.Equal(t, "exampleuser", args.HTTPClientConfig.BasicAuth.Username)
	require.Equal(t, rivertypes.Secret("examplepassword"), args.HTTPClientConfig.BasicAuth.Password)
}

func TestValidate(t *testing.T) {
	noServer := `
	refresh_interval = "10s"
	basic_auth {
		username = "exampleuser"
		password = "examplepassword"
	}`

	var args Arguments
	err := river.Unmarshal([]byte(noServer), &args)
	require.Error(t, err)

	emptyServer := `
	server = ""
	refresh_interval = "10s"
	basic_auth {
		username = "exampleuser"
		password = "examplepassword"
	}`
	err = river.Unmarshal([]byte(emptyServer), &args)
	require.Error(t, err)

	invalidServer := `
	server = "localhost"
	refresh_interval = "10s"
	basic_auth {
		username = "exampleuser"
		password = "examplepassword"
	}`
	err = river.Unmarshal([]byte(invalidServer), &args)
	require.Error(t, err)
}

func TestConvert(t *testing.T) {
	args := Arguments{
		Server:          "http://localhost:8080/eureka/v1",
		RefreshInterval: 10 * time.Second,
		HTTPClientConfig: config.HTTPClientConfig{
			BasicAuth: &config.BasicAuth{
				Username: "exampleuser",
				Password: "examplepassword",
			},
			FollowRedirects: false,
			EnableHTTP2:     false,
		},
	}

	sdConfig := args.Convert()

	expected := prom_sd.SDConfig{
		Server:          "http://localhost:8080/eureka/v1",
		RefreshInterval: model.Duration(10 * time.Second),
		HTTPClientConfig: promcfg.HTTPClientConfig{
			BasicAuth: &promcfg.BasicAuth{
				Username: "exampleuser",
				Password: "examplepassword",
			},
			FollowRedirects: false,
			EnableHTTP2:     false,
		},
	}
	require.Equal(t, expected, *sdConfig)
}
