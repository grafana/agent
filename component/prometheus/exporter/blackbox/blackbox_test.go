package blackbox

import (
	"testing"
	"time"

	"github.com/grafana/agent/component"
	"github.com/grafana/agent/component/discovery"
	"github.com/grafana/agent/pkg/river"
	"github.com/prometheus/common/model"
	"github.com/stretchr/testify/require"
)

func TestUnmarshalRiver(t *testing.T) {
	riverCfg := `
		config_file = "modules.yml"
		target "target_a" {
			address = "http://example.com"
			module = "http_2xx"
		}
		target "target_b" {
			address = "http://grafana.com"
			module = "http_2xx"
		}
		probe_timeout_offset = "0.5s"
`
	var args Arguments
	err := river.Unmarshal([]byte(riverCfg), &args)
	require.NoError(t, err)
	require.Equal(t, "modules.yml", args.ConfigFile)
	require.Equal(t, 2, len(args.Targets))
	require.Equal(t, 500*time.Millisecond, args.ProbeTimeoutOffset)
	require.Contains(t, "target_a", args.Targets[0].Name)
	require.Contains(t, "http://example.com", args.Targets[0].Target)
	require.Contains(t, "http_2xx", args.Targets[0].Module)
	require.Contains(t, "target_b", args.Targets[1].Name)
	require.Contains(t, "http://grafana.com", args.Targets[1].Target)
	require.Contains(t, "http_2xx", args.Targets[1].Module)
}

func TestConvertConfig(t *testing.T) {
	args := Arguments{
		ConfigFile:         "modules.yml",
		Targets:            TargetBlock{{Name: "target_a", Target: "http://example.com", Module: "http_2xx"}},
		ProbeTimeoutOffset: 1 * time.Second,
	}

	res := args.Convert()
	require.Equal(t, "modules.yml", res.BlackboxConfigFile)
	require.Equal(t, 1, len(res.BlackboxTargets))
	require.Contains(t, "target_a", res.BlackboxTargets[0].Name)
	require.Contains(t, "http://example.com", res.BlackboxTargets[0].Target)
	require.Contains(t, "http_2xx", res.BlackboxTargets[0].Module)
	require.Equal(t, 1.0, res.ProbeTimeoutOffset)
}

func TestConvertTargets(t *testing.T) {
	targets := TargetBlock{{
		Name:   "target_a",
		Target: "http://example.com",
		Module: "http_2xx",
	}}

	res := targets.Convert()
	require.Equal(t, 1, len(res))
	require.Contains(t, "target_a", res[0].Name)
	require.Contains(t, "http://example.com", res[0].Target)
	require.Contains(t, "http_2xx", res[0].Module)
}

func TestBuildBlackboxTargets(t *testing.T) {
	baseArgs := Arguments{
		ConfigFile:         "modules.yml",
		Targets:            TargetBlock{{Name: "target_a", Target: "http://example.com", Module: "http_2xx"}},
		ProbeTimeoutOffset: 1.0,
	}
	baseTarget := discovery.Target{
		model.SchemeLabel:                   "http",
		model.MetricsPathLabel:              "component/prometheus.exporter.blackbox.default/metrics",
		"instance":                          "prometheus.exporter.blackbox.default",
		"job":                               "integrations/blackbox",
		"__meta_agent_integration_name":     "blackbox",
		"__meta_agent_integration_instance": "prometheus.exporter.blackbox.default",
	}
	args := component.Arguments(baseArgs)
	targets := buildBlackboxTargets(baseTarget, args)
	require.Equal(t, 1, len(targets))
	require.Equal(t, "integrations/blackbox/target_a", targets[0]["job"])
	require.Equal(t, "http://example.com", targets[0]["__param_target"])
	require.Equal(t, "http_2xx", targets[0]["__param_module"])
}
