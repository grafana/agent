package snmp

import (
	"testing"
	"time"

	"github.com/grafana/agent/component"
	"github.com/grafana/agent/component/discovery"
	"github.com/grafana/agent/pkg/river"

	"github.com/prometheus/common/model"
	snmp_config "github.com/prometheus/snmp_exporter/config"
	"github.com/stretchr/testify/require"
)

func TestUnmarshalRiver(t *testing.T) {
	riverCfg := `
		config_file = "modules.yml"
		target "network_switch_1" {
			address = "192.168.1.2"
			module = "if_mib"
			walk_params = "public"
		}
		target "network_router_2" {
			address = "192.168.1.3"
			module = "mikrotik"
			walk_params = "private"
		}
		walk_param "private" {
			version = "2"
			auth {
				community = "secret"
			}
		}
		walk_param "public" {
			version = "2"
			auth {
				community = "public"
			}
		}		
`
	var cfg Config
	err := river.Unmarshal([]byte(riverCfg), &cfg)
	require.NoError(t, err)
	require.Equal(t, "modules.yml", cfg.ConfigFile)
	require.Equal(t, 2, len(cfg.Targets))

	require.Contains(t, "network_switch_1", cfg.Targets[0].Name)
	require.Contains(t, "192.168.1.2", cfg.Targets[0].Target)
	require.Contains(t, "if_mib", cfg.Targets[0].Module)
	require.Contains(t, "public", cfg.Targets[0].WalkParams)

	require.Contains(t, "network_router_2", cfg.Targets[1].Name)
	require.Contains(t, "192.168.1.3", cfg.Targets[1].Target)
	require.Contains(t, "mikrotik", cfg.Targets[1].Module)
	require.Contains(t, "private", cfg.Targets[1].WalkParams)

	require.Equal(t, 2, len(cfg.WalkParams))

	require.Contains(t, "private", cfg.WalkParams[0].Name)
	require.Contains(t, "secret", cfg.WalkParams[0].Auth.Community)

	require.Contains(t, "public", cfg.WalkParams[1].Name)
	require.Contains(t, "public", cfg.WalkParams[1].Auth.Community)

}

func TestConvertConfig(t *testing.T) {
	cfg := Config{
		ConfigFile: "modules.yml",
		Targets:    TargetBlock{{Name: "network_switch_1", Target: "192.168.1.2", Module: "if_mib"}},
		WalkParams: WalkParams{{Name: "public", Version: 2, Auth: Auth{Community: "public"}}},
	}

	res := cfg.Convert()
	require.Equal(t, "modules.yml", res.SnmpConfigFile)
	require.Equal(t, 1, len(res.SnmpTargets))
	require.Equal(t, "network_switch_1", res.SnmpTargets[0].Name)
}

func TestConvertTargets(t *testing.T) {
	targets := TargetBlock{{
		Name:   "network_switch_1",
		Target: "192.168.1.2",
		Module: "if_mib",
	}}

	res := targets.Convert()
	require.Equal(t, 1, len(res))
	require.Equal(t, "network_switch_1", res[0].Name)
	require.Equal(t, "192.168.1.2", res[0].Target)
	require.Equal(t, "if_mib", res[0].Module)
}

func TestConvertWalkParams(t *testing.T) {
	walkParams := WalkParams{{
		Name:                    "public",
		Version:                 2,
		MaxRepetitions:          uint32(10),
		Retries:                 3,
		Timeout:                 time.Duration(5 * time.Second),
		UseUnconnectedUDPSocket: true,
	}}

	res := walkParams.Convert()
	require.Equal(t, 1, len(res))
	require.Equal(t, 2, res["public"].Version)
	require.Equal(t, uint32(10), res["public"].MaxRepetitions)
	require.Equal(t, 3, res["public"].Retries)
	require.Equal(t, time.Duration(5*time.Second), res["public"].Timeout)
	require.Equal(t, true, res["public"].UseUnconnectedUDPSocket)
}

func TestConvertAuth(t *testing.T) {
	auth := Auth{
		Community:     "public",
		SecurityLevel: "authPriv",
		Username:      "user",
		AuthProtocol:  "MD5",
		PrivProtocol:  "DES",
		Password:      "password",
		PrivPassword:  "password",
		ContextName:   "context",
	}
	res := auth.Convert()
	require.Equal(t, snmp_config.Secret("public"), res.Community)
	require.Equal(t, "authPriv", res.SecurityLevel)
	require.Equal(t, "user", res.Username)
	require.Equal(t, "MD5", res.AuthProtocol)
	require.Equal(t, "DES", res.PrivProtocol)
	require.Equal(t, snmp_config.Secret("password"), res.Password)
	require.Equal(t, snmp_config.Secret("password"), res.PrivPassword)
	require.Equal(t, "context", res.ContextName)
}

func TestBuildSNMPTargets(t *testing.T) {
	cfg := Config{
		ConfigFile: "modules.yml",
		Targets:    TargetBlock{{Name: "network_switch_1", Target: "192.168.1.2", Module: "if_mib", WalkParams: "public"}},
		WalkParams: WalkParams{{Name: "public", Version: 2, Auth: Auth{Community: "public"}}},
	}
	baseTarget := discovery.Target{
		model.SchemeLabel:                   "http",
		model.MetricsPathLabel:              "component/prometheus.exporter.snmp.default/metrics",
		"instance":                          "prometheus.exporter.snmp.default",
		"job":                               "integrations/snmp",
		"__meta_agent_integration_name":     "snmp",
		"__meta_agent_integration_instance": "prometheus.exporter.snmp.default",
	}
	args := component.Arguments(cfg)
	targets := buildSNMPTargets(baseTarget, args)
	require.Equal(t, 1, len(targets))
	require.Equal(t, "integrations/snmp/network_switch_1", targets[0]["job"])
	require.Equal(t, "192.168.1.2", targets[0]["__param_target"])
	require.Equal(t, "if_mib", targets[0]["__param_module"])
	require.Equal(t, "public", targets[0]["__param_walk_params"])
}
