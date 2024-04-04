//go:build (linux && arm64) || (linux && amd64)

package beyla

import (
	"errors"
	"testing"

	"github.com/grafana/beyla/pkg/beyla"
	"github.com/grafana/beyla/pkg/services"
	"github.com/grafana/beyla/pkg/transform"
	"github.com/grafana/river"
	"github.com/stretchr/testify/require"
)

func TestArguments_UnmarshalRiver(t *testing.T) {
	in := `
		open_port = "80,443,8000-8999"
		executable_name = "test"
		routes {
			unmatched = "wildcard"
			patterns = ["/api/v1/*"]
			ignored_patterns = ["/api/v1/health"]
			ignore_mode = "all"
		}
		attributes {
			kubernetes {
				enable = "true"
			}
		}
		discovery {
			services {
				name = "test"
				namespace = "default"
				open_ports = "80,443"
			}
			services {
				name = "test2"
				namespace = "default"
				open_ports = "80,443"
			}
		}
		output { /* no-op */ }
	`
	var args Arguments
	require.NoError(t, river.Unmarshal([]byte(in), &args))
	cfg, err := args.Convert()
	require.NoError(t, err)
	require.Equal(t, services.PortEnum{Ranges: []services.PortRange{{Start: 80, End: 0}, {Start: 443, End: 0}, {Start: 8000, End: 8999}}}, cfg.Port)
	require.True(t, cfg.Exec.IsSet())
	require.Equal(t, transform.UnmatchType("wildcard"), cfg.Routes.Unmatch)
	require.Equal(t, []string{"/api/v1/*"}, cfg.Routes.Patterns)
	require.Equal(t, []string{"/api/v1/health"}, cfg.Routes.IgnorePatterns)
	require.Equal(t, transform.IgnoreMode("all"), cfg.Routes.IgnoredEvents)
	require.Equal(t, transform.KubeEnableFlag("true"), cfg.Attributes.Kubernetes.Enable)
	require.Len(t, cfg.Discovery.Services, 2)
	require.Equal(t, "test", cfg.Discovery.Services[0].Name)
	require.Equal(t, "default", cfg.Discovery.Services[0].Namespace)
}

func TestArguments_UnmarshalInvalidRiver(t *testing.T) {
	var tests = []struct {
		testname      string
		cfg           string
		expectedError string
	}{
		{
			"invalid regex",
			`
		executable_name = "["
		`,
			"error parsing regexp: missing closing ]: `[`",
		},
		{
			"invalid port range",
			`
		open_port = "-8000"
		`,
			"invalid port range \"-8000\". Must be a comma-separated list of numeric ports or port ranges (e.g. 8000-8999)",
		},
	}
	for _, tt := range tests {
		t.Run(tt.testname, func(t *testing.T) {
			var args Arguments
			require.NoError(t, river.Unmarshal([]byte(tt.cfg), &args))
			_, err := args.Convert()
			require.EqualError(t, err, tt.expectedError)
		})
	}
}

func TestConvert_Routes(t *testing.T) {
	args := Routes{
		Unmatch:        "wildcard",
		Patterns:       []string{"/api/v1/*"},
		IgnorePatterns: []string{"/api/v1/health"},
		IgnoredEvents:  "all",
	}

	expectedConfig := &transform.RoutesConfig{
		Unmatch:        transform.UnmatchType(args.Unmatch),
		Patterns:       args.Patterns,
		IgnorePatterns: args.IgnorePatterns,
		IgnoredEvents:  transform.IgnoreMode(args.IgnoredEvents),
	}

	config := args.Convert()

	require.Equal(t, expectedConfig, config)
}

func TestConvert_Attribute(t *testing.T) {
	args := Attributes{
		Kubernetes: KubernetesDecorator{
			Enable: "true",
		},
	}

	expectedConfig := beyla.Attributes{
		Kubernetes: transform.KubernetesDecorator{
			Enable: transform.KubeEnableFlag(args.Kubernetes.Enable),
		},
	}

	config := args.Convert()

	require.Equal(t, expectedConfig, config)
}

func TestConvert_Discovery(t *testing.T) {
	args := Discovery{
		Services: []Service{
			{
				Name:      "test",
				Namespace: "default",
				OpenPorts: "80",
				Path:      "/api/v1/*",
			},
		},
	}
	config, err := args.Convert()

	require.NoError(t, err)
	require.Len(t, config.Services, 1)
	require.Equal(t, "test", config.Services[0].Name)
	require.Equal(t, "default", config.Services[0].Namespace)
	require.Equal(t, services.PortEnum{Ranges: []services.PortRange{{Start: 80, End: 0}}}, config.Services[0].OpenPorts)
	require.True(t, config.Services[0].Path.IsSet())
}

func TestArguments_Validate(t *testing.T) {
	tests := []struct {
		name     string
		args     Arguments
		expected error
	}{
		{
			name:     "empty arguments",
			args:     Arguments{},
			expected: errors.New("you need to define at least open_port, executable_name, or services in the discovery section"),
		},
		{
			name: "with service discovery",
			args: Arguments{
				Discovery: Discovery{
					Services: []Service{
						{
							Name:      "test",
							Namespace: "default",
							OpenPorts: "80",
							Path:      "/api/v1/*",
						},
					},
				},
			},
			expected: nil,
		},
		{
			name: "with port",
			args: Arguments{
				Port: "80",
			},
			expected: nil,
		},
		{
			name: "with executable name",
			args: Arguments{
				ExecutableName: "test",
			},
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.args.Validate()
			require.Equal(t, tt.expected, err)
		})
	}
}
