package loadbalancing_test

import (
	"testing"
	"time"

	"github.com/grafana/agent/component/otelcol/exporter/loadbalancing"
	"github.com/grafana/agent/pkg/river"
	"github.com/open-telemetry/opentelemetry-collector-contrib/exporter/loadbalancingexporter"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/config/configgrpc"
	"go.opentelemetry.io/collector/config/configopaque"
	"go.opentelemetry.io/collector/exporter/exporterhelper"
	"go.opentelemetry.io/collector/exporter/otlpexporter"
)

func TestConfigConversion(t *testing.T) {
	defaultProtocol := loadbalancingexporter.Protocol{
		OTLP: otlpexporter.Config{
			GRPCClientSettings: configgrpc.GRPCClientSettings{
				Endpoint:        "",
				Compression:     "gzip",
				WriteBufferSize: 512 * 1024,
				Headers:         map[string]configopaque.String{},
			},
		},
	}

	tests := []struct {
		testName string
		agentCfg string
		expected loadbalancingexporter.Config
	}{
		{
			testName: "static",
			agentCfg: `
			resolver {
				static {
					hostnames = ["endpoint-1"]
				}
			}
			protocol {
				otlp {
					client {}
				}
			}
			`,
			expected: loadbalancingexporter.Config{
				Resolver: loadbalancingexporter.ResolverSettings{
					Static: &loadbalancingexporter.StaticResolver{
						Hostnames: []string{"endpoint-1"},
					},
					DNS: nil,
				},
				RoutingKey: "traceID",
				Protocol:   defaultProtocol,
			},
		},
		{
			testName: "static with service routing",
			agentCfg: `
			routing_key = "service"
			resolver {
				static {
					hostnames = ["endpoint-1"]
				}
			}
			protocol {
				otlp {
					client {}
				}
			}
			`,
			expected: loadbalancingexporter.Config{
				Resolver: loadbalancingexporter.ResolverSettings{
					Static: &loadbalancingexporter.StaticResolver{
						Hostnames: []string{"endpoint-1"},
					},
					DNS: nil,
				},
				RoutingKey: "service",
				Protocol:   defaultProtocol,
			},
		},
		{
			testName: "static with timeout",
			agentCfg: `
			protocol {
				otlp {
					timeout = "1s"
					client {}
				}
			}
			resolver {
				static {
					hostnames = ["endpoint-1", "endpoint-2:55678"]
				}
			}
			`,
			expected: loadbalancingexporter.Config{
				Protocol: loadbalancingexporter.Protocol{
					OTLP: otlpexporter.Config{
						TimeoutSettings: exporterhelper.TimeoutSettings{
							Timeout: 1 * time.Second,
						},
						GRPCClientSettings: configgrpc.GRPCClientSettings{
							Endpoint:        "",
							Compression:     "gzip",
							WriteBufferSize: 512 * 1024,
							Headers:         map[string]configopaque.String{},
						},
					},
				},
				Resolver: loadbalancingexporter.ResolverSettings{
					Static: &loadbalancingexporter.StaticResolver{
						Hostnames: []string{"endpoint-1", "endpoint-2:55678"},
					},
					DNS: nil,
				},
				RoutingKey: "traceID",
			},
		},
		{
			testName: "dns with defaults",
			agentCfg: `
			resolver {
				dns {
					hostname = "service-1"
				}
			}
			protocol {
				otlp {
					client {}
				}
			}
			`,
			expected: loadbalancingexporter.Config{
				Resolver: loadbalancingexporter.ResolverSettings{
					Static: nil,
					DNS: &loadbalancingexporter.DNSResolver{
						Hostname: "service-1",
						Port:     "4317",
						Interval: 5 * time.Second,
						Timeout:  1 * time.Second,
					},
				},
				RoutingKey: "traceID",
				Protocol:   defaultProtocol,
			},
		},
		{
			testName: "dns with non-defaults",
			agentCfg: `
			resolver {
				dns {
					hostname = "service-1"
					port = "55690"
					interval = "123s"
					timeout = "321s"
				}
			}
			protocol {
				otlp {
					client {}
				}
			}
			`,
			expected: loadbalancingexporter.Config{
				Resolver: loadbalancingexporter.ResolverSettings{
					Static: nil,
					DNS: &loadbalancingexporter.DNSResolver{
						Hostname: "service-1",
						Port:     "55690",
						Interval: 123 * time.Second,
						Timeout:  321 * time.Second,
					},
				},
				RoutingKey: "traceID",
				Protocol:   defaultProtocol,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.testName, func(t *testing.T) {
			var args loadbalancing.Arguments
			require.NoError(t, river.Unmarshal([]byte(tc.agentCfg), &args))
			actual, err := args.Convert()
			require.NoError(t, err)
			require.Equal(t, &tc.expected, actual.(*loadbalancingexporter.Config))
		})
	}
}
