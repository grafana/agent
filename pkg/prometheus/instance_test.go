package prometheus

import (
	"testing"

	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/config"
	sd_config "github.com/prometheus/prometheus/discovery/config"
	"github.com/prometheus/prometheus/discovery/targetgroup"
	"github.com/stretchr/testify/require"
)

// TODO(rfratto): test the instance.
//
// 1. mock the WAL, ensure that samples get sent to it
// 2. test truncateLoop, getRemoteWriteTimestamp
//
// No need to go crazy and mock all external dependencies,
// just create a temp directory for the WAL, create a custom
// registerer exposed over httptest so can just scrape it
// just like a "real" external target.

func TestInstanceConfig_ApplyDefaults(t *testing.T) {
	global := config.DefaultGlobalConfig
	cfg := &InstanceConfig{
		Name: "instance",
		ScrapeConfigs: []*config.ScrapeConfig{{
			JobName: "scrape",
			ServiceDiscoveryConfig: sd_config.ServiceDiscoveryConfig{
				StaticConfigs: []*targetgroup.Group{{
					Targets: []model.LabelSet{{
						model.AddressLabel: model.LabelValue("127.0.0.1:12345"),
					}},
					Labels: model.LabelSet{"cluster": "localhost"},
				}},
			},
		}},
	}

	cfg.ApplyDefaults(&global)
	for _, sc := range cfg.ScrapeConfigs {
		require.Equal(t, sc.ScrapeInterval, global.ScrapeInterval)
		require.Equal(t, sc.ScrapeTimeout, global.ScrapeTimeout)
		require.Equal(t, sc.RelabelConfigs, DefaultRelabelConfigs)
	}
}
